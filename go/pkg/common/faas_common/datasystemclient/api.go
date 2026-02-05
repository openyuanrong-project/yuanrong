/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package datasystemclient is data system client used for communicating with data system worker.
// To use data system, you should export the data system lib path. Please refer to the Dockerfile of the frontend.
// The lib should copied to home/sn/bin/datasystem/lib. Please refer to
// functioncore/build/common/common_compile.sh and the Dockerfile of the frontend.
// NOTE: To change the version of data system, must revise the version in the common_compile.sh, test.sh and the go.mod
package datasystemclient

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"go.uber.org/zap"
	
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/data"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

// Config - required parameter
type Config struct {
	TenantID     string
	NodeIP       string
	NoNeedGenKey bool
	NeedEncrypt  bool
	KeyPrefix    string
	Cluster      string
	Limit        uint64
	DataKey      []byte

	invalidIP       []string
	useLastUsedNode bool
}

const (
	errKeyNotFound      = 3
	errOutOfMemory      = 6
	errDsWorkerNotReady = 8
	errTryAgain         = 19
	errRPCCancelled     = 1000
	errRPCUnavailable   = 1002
	errAsyncQueueFull   = 2003

	errDsClientNil = 11001 // 数据系统错误目前到6000，errDsClientNil表示frontend中析构了该client，为frontend封装，因此从11000开始计数，避开数据系统节点返回的错误码区间
)

const (
	defaultUploadTTLSecond     = 24 * 60 * 60
	defaultExecuteTTLSecond    = 30 * 60
	defaultDataSystemTimeoutMs = 60 * 1000
	defaultDownloadLimit       = 32 * 1024 * 1024
	defaultSubscribeTimeoutMs  = 100
	// DefaultDataSystemPort -
	DefaultDataSystemPort = 31501
)

const (
	// StreamEndElement -
	StreamEndElement = `\xE0\xFF\xE0\xFF`
	// StreamEndElementSize -
	StreamEndElementSize = 16
)

var (
	// ErrKeyNotFound - the key on data system is not found
	ErrKeyNotFound = errors.New("key not found on data system")
	// ErrValueSizeExceeded - the value size exceeded the limit
	ErrValueSizeExceeded = errors.New("value size exceeded the limit")
	// UploadTTLSecond - upload data to data system timeout in uploadFile function
	UploadTTLSecond uint32
	// ExecuteTTLSecond - upload data to data system timeout in execute function
	// should greater than the time of algorithm execute
	ExecuteTTLSecond uint32
	// UploadWriteMode - L2Cache policy used during uploading in uploadFile function
	UploadWriteMode api.WriteModeEnum
	// ExecuteWriteMode - L2Cache policy used during uploading in execute function
	ExecuteWriteMode api.WriteModeEnum

	nodeEtcdKeyPrefix = "/datasystem/cluster/"

	timeoutMs int
	port      int

	clientMap             = concurrentMap{mp: make(map[string]*nodeIP2ClientMap)}
	localClientLibruntime api.LibruntimeAPI
)

// SubscribeParam -
type SubscribeParam struct {
	StreamName       string
	TimeoutMs        uint32
	ExpectReceiveNum int32
	Callback         func()
	TraceId          string
}

func initDataSystemCommon(cfg *types.DataSystemConfig, stopCh <-chan struct{}) {
	port = DefaultDataSystemPort
	timeoutMs = cfg.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = defaultDataSystemTimeoutMs
	}
	var dataSystemKeyPrefixList []string
	if len(cfg.Clusters) != 0 {
		for _, cluster := range cfg.Clusters {
			dataSystemKeyPrefix := "/" + cluster + nodeEtcdKeyPrefix
			dataSystemKeyPrefixList = append(dataSystemKeyPrefixList, dataSystemKeyPrefix)
		}
	} else {
		dataSystemKeyPrefixList = append(dataSystemKeyPrefixList, nodeEtcdKeyPrefix)
	}
	log.GetLogger().Infof("init data system success,timeoutMs: %d", timeoutMs)
	go StartWatch(dataSystemKeyPrefixList, stopCh)
}

func setLocalClient(rt api.LibruntimeAPI) {
	localClientLibruntime = rt
}

// InitDataSystemLibruntime - init data system before using api
func InitDataSystemLibruntime(cfg *types.DataSystemConfig, rt api.LibruntimeAPI, stopCh <-chan struct{}) {
	setLocalClient(rt)
	initDataSystemCommon(cfg, stopCh)
}

func getClient(cfg *Config, traceId string) (DsClientImpl, bool, error) {
	logger := log.GetLogger().With(zap.Any("traceId", traceId))
	cache, err := getDataSystemCacheByCluster(cfg.Cluster)
	if err != nil {
		logger.Warnf("getDataSystemCacheByCluster failed, err: %s", err.Error())
		return DsClientImpl{}, false, err
	}
	if cfg.NodeIP == "" || !cache.ifNodeExist(cfg.NodeIP) {
		// if node ip is empty, get a random dataSystem node
		if cfg.useLastUsedNode {
			cfg.NodeIP, err = cache.getLastUsedNodeWithInvalidNode(cfg.invalidIP)
		} else {
			cfg.NodeIP, err = cache.getRandomNodeWithInvalidNode(cfg.invalidIP)
		}
		if err != nil {
			return DsClientImpl{}, false, err
		}
		logger.Infof("get a random node ip: %s", cfg.NodeIP)

	}

	// double check
	if client, existed := clientMap.get(cfg.TenantID, cfg.NodeIP); existed {
		return client, false, nil
	}

	client, err := clientMap.getOrCreate(cfg.TenantID, cfg.NodeIP)
	if err != nil {
		cache.invalidateNode(cfg.NodeIP)
		clientMap.deleteClient(cfg.TenantID, cfg.NodeIP)
		logger.Warnf("Failed to create the client. replace the node and try again.: %s, err: %s",
			cfg.NodeIP, err.Error())
		return DsClientImpl{}, true, err
	}
	return client, false, nil
}

func getDataSystemCacheByCluster(cluster string) (*Cache, error) {
	if cluster == "" {
		cluster = noCluster
	}
	cacheData, ok := dataSystemCache.Load(cluster)
	if !ok {
		log.GetLogger().Errorf("no datasystem node in cluster %s", cluster)
		return nil, errors.New("no data system node is available")
	}
	cache, ok := cacheData.(*Cache)
	if !ok {
		return nil, errors.New("dataSystem cache is invalid")
	}
	return cache, nil
}

// shouldRetry -  数据系统相关错误码处理逻辑：1.需要重试的错误码以白名单方式处理，2.不在白名单中的错误码直接返回失败
func shouldRetry(code int) bool {
	retryCode := map[int]struct{}{
		errOutOfMemory:      {},
		errAsyncQueueFull:   {},
		errRPCCancelled:     {},
		errTryAgain:         {},
		errDsClientNil:      {},
		errRPCUnavailable:   {},
		errDsWorkerNotReady: {},
	}
	_, ok := retryCode[code]
	return ok
}

// UploadWithoutKeyRetry - the key is returned by data system
func UploadWithoutKeyRetry(value []byte, config *Config, param api.SetParam, traceID string) (string, error) {
	for {
		genKey, needRetry, err := uploadWithoutKey(value, config, param, traceID)
		if err == nil {
			return genKey, nil
		}
		if needRetry {
			log.GetLogger().Debugf("upload without key will retry, failed ip: %s", config.NodeIP)
			config.invalidIP = append(config.invalidIP, config.NodeIP)
			config.NodeIP = ""
			continue
		}
		return "", err
	}
}

func uploadWithoutKey(value []byte, config *Config, param api.SetParam, traceID string) (string, bool, error) {
	dsClient, retry, err := getClient(config, traceID)
	if err != nil {
		return "", retry, err
	}
	if dsClient.kvClient == nil {
		return "", false, fmt.Errorf("dsclient is nil")
	}
	runtime.LockOSThread()
	dsClient.kvClient.SetTraceID(traceID)
	key, status := dsClient.kvClient.KVSetWithoutKey(value, param)
	runtime.UnlockOSThread()
	if status.IsError() {
		if shouldRetry(status.Code) {
			log.GetLogger().Warnf("uploadWithoutKey dsClient(nodeIP: %s) is unavailable, code: %d, err: %s,"+
				" retry other clients, traceID: %s", config.NodeIP, status.Code, status.Err, traceID)
			return "", true, status.Err
		}
		log.GetLogger().Warnf("dsClient(nodeIP: %s) is unavailable, code: %d, can't retry err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return "", false, status.Err
	}
	return key, false, nil
}

// UploadWithKeyRetry - the key is returned by data system 返回的是数据系统生成的key，不带前缀拼接
func UploadWithKeyRetry(value []byte, config *Config, param api.SetParam, traceID string) (string, error) {
	for {
		genKey, retry, err := uploadWithKey(value, config, param, traceID)
		if err == nil {
			return genKey, nil
		}
		if retry {
			log.GetLogger().Debugf("upload with key will retry, failed ip: %s,traceID: %s",
				config.NodeIP, traceID)
			config.invalidIP = append(config.invalidIP, config.NodeIP)
			config.NodeIP = ""
			continue
		}
		log.GetLogger().Errorf("upload with key failed err: %s,NodeIP: %s,traceID: %s", err.Error(),
			config.NodeIP, traceID)
		return "", err
	}
}

func uploadWithKey(value []byte, config *Config, param api.SetParam, traceID string) (string, bool, error) {
	dsClient, retry, err := getClient(config, traceID)
	if err != nil {
		return "", retry, err
	}
	if dsClient.kvClient == nil {
		return "", false, fmt.Errorf("dsclient is nil")
	}
	var key string
	var genKey string
	key, genKey, err = getDataSystemKey(config, dsClient, traceID)
	if err != nil {
		return "", true, err
	}
	if config.NeedEncrypt {
		value, err = encryptData(config, value)
		if err != nil {
			return "", false, fmt.Errorf("failed to encrypt value: %v", err)
		}
	}
	runtime.LockOSThread()
	dsClient.kvClient.SetTraceID(traceID)
	if err = localClientLibruntime.SetTenantID(config.TenantID); err != nil {
		runtime.UnlockOSThread()
		return "", false, err
	}
	status := dsClient.kvClient.KVSet(key, value, param)
	runtime.UnlockOSThread()
	if status.IsError() {
		if shouldRetry(status.Code) {
			log.GetLogger().Warnf("uploadWithKey dsClient(nodeIP: %s) is unavailable, code: %d, err: %s,"+
				" retry other clients, traceID: %s", config.NodeIP, status.Code, status.Err, traceID)
			return "", true, status.Err
		}
		log.GetLogger().Warnf("dsClient(nodeIP: %s) is unavailable, code: %d, can't retry err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return "", false, status.Err
	}
	return genKey, false, nil
}

func getDataSystemKey(config *Config, dsClient DsClientImpl,
	traceID string) (string, string, error) {
	if config.NoNeedGenKey {
		return config.KeyPrefix, "", nil
	}

	genKey := dsClient.kvClient.GenerateKey()

	if genKey == "" {
		log.GetLogger().Warnf("failed to generate key, retry other clients, nodeIP: %s, traceId: %s",
			config.NodeIP, traceID)
		return "", "", fmt.Errorf("failed to generate key")
	}

	key := config.KeyPrefix + genKey

	return key, genKey, nil
}

// DownloadArrayRetry - download multiple keys at once
func DownloadArrayRetry(keys []string, config *Config, traceID string) ([][]byte, error) {
	if config.Limit == 0 {
		config.Limit = defaultDownloadLimit
	}
	for {
		byteValues, retry, err := downloadArray(keys, config, traceID)
		if err == nil {
			return byteValues, nil
		}
		if retry {
			config.invalidIP = append(config.invalidIP, config.NodeIP)
			config.NodeIP = ""
			continue
		}
		return nil, err
	}
}

func downloadArray(keys []string, config *Config, traceID string) ([][]byte, bool, error) {
	dsClient, retry, err := getClient(config, traceID)
	if err != nil {
		return nil, retry, err
	}
	if dsClient.kvClient == nil {
		return nil, false, fmt.Errorf("dsclient is nil")
	}
	runtime.LockOSThread()
	dsClient.kvClient.SetTraceID(traceID)
	if err = localClientLibruntime.SetTenantID(config.TenantID); err != nil {
		runtime.UnlockOSThread()
		return nil, false, err
	}
	sizes, queryErr := dsClient.kvClient.KVQuerySize(keys)
	retry, err = checkStatus(queryErr, config, traceID)
	if err != nil {
		runtime.UnlockOSThread()
		return nil, retry, err
	}
	var totalSize uint64
	for _, val := range sizes {
		totalSize += val
	}
	if totalSize > config.Limit {
		runtime.UnlockOSThread()
		log.GetLogger().Errorf("query size: %d exceeded the limit: %d", totalSize, config.Limit)
		return nil, false, ErrValueSizeExceeded
	}
	byteValues, status := dsClient.kvClient.KVGetMulti(keys)
	runtime.UnlockOSThread()
	retry, err = checkStatus(status, config, traceID)
	if err != nil {
		return nil, retry, err
	}
	for i, v := range byteValues {
		if config.NeedEncrypt {
			val, err := decryptData(config, v)
			if err != nil {
				return nil, false, fmt.Errorf("failed to decrypt value: %v", err)
			}
			byteValues[i] = val
			continue
		}
		byteValues[i] = v
	}

	return byteValues, false, nil
}

func checkStatus(status api.ErrorInfo, config *Config, traceID string) (bool, error) {
	if status.IsError() {
		if status.Code == errKeyNotFound {
			return false, ErrKeyNotFound
		}
		if shouldRetry(status.Code) {
			log.GetLogger().Warnf("dsClient(nodeIP: %s) is unavailable, code: %d, err: %s,"+
				" retry other clients, traceID: %s", config.NodeIP, status.Code, status.Err, traceID)
			return true, status.Err
		}
		log.GetLogger().Warnf("dsClient(nodeIP: %s) is unavailable, code: %d, can't retry err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return false, status.Err
	}
	return false, nil
}

// DeleteArrayRetry - delete multiple keys at once; the first output parameter is the keys failed to delete
func DeleteArrayRetry(keys []string, config *Config, traceID string) ([]string, error) {
	for {
		failedKeys, retry, err := deleteArray(keys, config, traceID)
		if err == nil {
			return nil, nil
		}
		if retry {
			config.invalidIP = append(config.invalidIP, config.NodeIP)
			config.NodeIP = ""
			continue
		}
		return failedKeys, err
	}
}

func deleteArray(keys []string, config *Config, traceID string) ([]string, bool, error) {
	dsClient, retry, err := getClient(config, traceID)
	if err != nil {
		return keys, retry, err
	}
	if dsClient.kvClient == nil {
		return nil, false, fmt.Errorf("dsclient is nil")
	}
	runtime.LockOSThread()
	dsClient.kvClient.SetTraceID(traceID)
	if err = localClientLibruntime.SetTenantID(config.TenantID); err != nil {
		runtime.UnlockOSThread()
		return nil, false, err
	}
	failedKeys, status := dsClient.kvClient.KVDelMulti(keys)
	runtime.UnlockOSThread()
	if status.IsError() {
		if shouldRetry(status.Code) {
			log.GetLogger().Warnf("dsClient(nodeIP: %s) is unavailable, code: %d, err: %s,"+
				" retry other clients, traceID: %s", config.NodeIP, status.Code, status.Err, traceID)
			return keys, true, status.Err
		}
		log.GetLogger().Warnf("dsClient(nodeIP: %s) is unavailable, code: %d, can't retry err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return keys, false, status.Err
	} else if len(failedKeys) > 0 {
		return failedKeys, false, fmt.Errorf("some keys failed to delete")
	}
	return nil, false, nil
}

// ObjPut - put objects to datasystem
func ObjPut(req *data.PutRequest, config *Config, traceID string) api.ErrorInfo {
	if localClientLibruntime == nil {
		return api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}
	paramLibruntime := api.PutParam{
		WriteMode:       api.WriteModeEnum(req.WriteMode),
		ConsistencyType: api.ConsistencyTypeEnum(req.ConsistencyType),
		CacheType:       api.CacheTypeEnum(req.CacheType),
	}
	runtime.LockOSThread()
	var status api.ErrorInfo
	err := putRaw(req, config.TenantID, paramLibruntime)
	if err != nil {
		status.Code = err.(api.ErrorInfo).Code
		status.Err = err
	}
	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("ObjPut dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return status
	}
	return status
}

func putRaw(req *data.PutRequest, tenantID string, paramLibruntime api.PutParam) error {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return err
	}
	if len(req.NestedObjectIds) == 0 {
		return localClientLibruntime.PutRaw(req.ObjectId, req.ObjectData, paramLibruntime)
	}
	return localClientLibruntime.PutRaw(req.ObjectId, req.ObjectData, paramLibruntime, req.NestedObjectIds...)
}

// ObjGet - get objects to datasystem
func ObjGet(req *data.GetRequest, config *Config, traceID string) ([][]byte, api.ErrorInfo) {
	if localClientLibruntime == nil {
		return nil, api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}

	var byteVals [][]byte
	var status api.ErrorInfo

	runtime.LockOSThread()
	var errInfo error
	byteVals, errInfo = getRaw(req, config.TenantID)
	log.GetLogger().Debugf("libruntime api get values size:%d", len(byteVals))
	if errInfo != nil {
		status.Code = errInfo.(api.ErrorInfo).Code
		status.Err = errInfo
	}
	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("ObjGet dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return byteVals, status
	}
	return byteVals, status
}

func getRaw(req *data.GetRequest, tenantID string) ([][]byte, error) {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return nil, err
	}
	return localClientLibruntime.GetRaw(req.ObjectIds, int(req.TimeoutMs))
}

// GIncreaseRef - increase global ref
func GIncreaseRef(req *data.IncreaseRefRequest, config *Config, traceID string) ([]string, api.ErrorInfo) {
	if localClientLibruntime == nil {
		return nil, api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}

	var status api.ErrorInfo
	var values []string

	runtime.LockOSThread()
	var errorInfo error
	values, errorInfo = gIncreaseRefRaw(req, config.TenantID)
	if errorInfo != nil {
		status.Code = errorInfo.(api.ErrorInfo).Code
		status.Err = errorInfo
	}

	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return values, status
	}
	return values, status
}

func gIncreaseRefRaw(req *data.IncreaseRefRequest, tenantID string) ([]string, error) {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return nil, err
	}
	return localClientLibruntime.GIncreaseRefRaw(req.ObjectIds, req.RemoteClientId)
}

// GDecreaseRef - decrease global ref
func GDecreaseRef(req *data.DecreaseRefRequest, config *Config, traceID string) ([]string, api.ErrorInfo) {
	if localClientLibruntime == nil {
		return nil, api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}

	var status api.ErrorInfo
	var values []string

	runtime.LockOSThread()
	var err error
	values, err = gDecreaseRefRaw(req, config.TenantID)
	if err != nil {
		status.Code = err.(api.ErrorInfo).Code
		status.Err = err
	}
	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("GDecreaseRef dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return values, status
	}
	return values, status
}

func gDecreaseRefRaw(req *data.DecreaseRefRequest, tenantID string) ([]string, error) {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return nil, err
	}
	return localClientLibruntime.GDecreaseRefRaw(req.ObjectIds, req.RemoteClientId)
}

// Set - set kv to datasystem
func Set(req *data.KvSetRequest, config *Config, traceID string) api.ErrorInfo {
	if localClientLibruntime == nil {
		return api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}
	paramLibruntime := api.SetParam{
		WriteMode: api.WriteModeEnum(req.WriteMode),
		TTLSecond: req.TtlSecond,
		Existence: req.Existence,
		CacheType: api.CacheTypeEnum(req.CacheType),
	}
	log.GetLogger().Debugf("set kv to datasystem, key:%s, param:%v", req.Key, paramLibruntime)
	var status api.ErrorInfo
	runtime.LockOSThread()
	if err := kvSet(req, config.TenantID, paramLibruntime); err != nil {
		status.Code = err.(api.ErrorInfo).Code
		status.Err = err
	}
	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("SetKV dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return status
	}
	return status
}

func kvSet(req *data.KvSetRequest, tenantID string, paramLibruntime api.SetParam) error {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return err
	}
	return localClientLibruntime.KVSet(req.Key, req.Value, paramLibruntime)
}

// MSetTx - set multi kvs to datasystem transactionally
func MSetTx(req *data.KvMSetTxRequest, config *Config, traceID string) api.ErrorInfo {
	if localClientLibruntime == nil {
		return api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}
	paramLibruntime := api.MSetParam{
		WriteMode: api.WriteModeEnum(req.WriteMode),
		TTLSecond: req.TtlSecond,
		Existence: req.Existence,
		CacheType: api.CacheTypeEnum(req.CacheType),
	}
	log.GetLogger().Debugf("set multi kvs to datasystem, param:%v", paramLibruntime)
	var status api.ErrorInfo
	runtime.LockOSThread()
	if err := kvMSetTx(req, config.TenantID, paramLibruntime); err != nil {
		status.Code = err.(api.ErrorInfo).Code
		status.Err = err
	}
	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("MSetTx dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return status
	}
	return status
}

func kvMSetTx(req *data.KvMSetTxRequest, tenantID string, paramLibruntime api.MSetParam) error {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return err
	}
	return localClientLibruntime.KVMSetTx(req.Keys, req.Values, paramLibruntime)
}

// Get - get kv to datasystem
func Get(req *data.KvGetRequest, config *Config, traceID string) ([][]byte, api.ErrorInfo) {
	if localClientLibruntime == nil {
		return nil, api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}
	var status api.ErrorInfo
	var byteValues [][]byte
	runtime.LockOSThread()
	var err error
	byteValues, err = kvGetMulti(req, config.TenantID)
	if err != nil {
		status.Code = err.(api.ErrorInfo).Code
		status.Err = err
	}
	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("Get dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return byteValues, status
	}
	return byteValues, status
}

func kvGetMulti(req *data.KvGetRequest, tenantID string) ([][]byte, error) {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return nil, err
	}
	return localClientLibruntime.KVGetMulti(req.Keys, uint(req.TimeoutMs))
}

// Del - del kv to datasystem
func Del(req *data.KvDelRequest, config *Config, traceID string) ([]string, api.ErrorInfo) {
	if localClientLibruntime == nil {
		return nil, api.ErrorInfo{Code: errRPCUnavailable, Err: fmt.Errorf("dsclient is nil")}
	}
	var values []string
	var status api.ErrorInfo

	runtime.LockOSThread()
	var err error
	values, err = kvDelMulti(req, config.TenantID)
	if err != nil {
		status.Code = err.(api.ErrorInfo).Code
		status.Err = err
	}
	runtime.UnlockOSThread()
	if status.IsError() {
		log.GetLogger().Warnf("Del dsClient(nodeIP: %s) is unavailable, code: %d, err: %s, traceID: %s",
			config.NodeIP, status.Code, status.Err, traceID)
		return values, status
	}
	return values, status
}

func kvDelMulti(req *data.KvDelRequest, tenantID string) ([]string, error) {
	if err := localClientLibruntime.SetTenantID(tenantID); err != nil {
		return nil, err
	}
	return localClientLibruntime.KVDelMulti(req.Keys)
}

// SubscribeStream -
func SubscribeStream(param SubscribeParam, ctx StreamCtx) error {
	var subscriptionConfig api.SubscriptionConfig
	subscriptionConfig.SubscriptionName = param.StreamName
	subscriptionConfig.SubscriptionType = api.Stream
	subscriptionConfig.TraceId = param.TraceId
	runtime.LockOSThread()
	if err := localClientLibruntime.SetTenantID(ctx.GetRequestHeader(constant.HeaderTenantID)); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	consumer, errorInfo := localClientLibruntime.Subscribe(param.StreamName, subscriptionConfig)
	runtime.UnlockOSThread()
	if errorInfo == nil {
		receiveStream(param, consumer, ctx)
		return nil

	}
	log.GetLogger().Warnf("create stream consumer failed, streamName: "+
		"%s, message: %s", param.StreamName, errorInfo.Error())
	return errorInfo
}

func receiveStream(param SubscribeParam, consumer api.StreamConsumer, ctx StreamCtx) {
	ctx.SetResponseHeader("Content-Type", "text/event-stream")
	ctx.SetResponseHeader("Cache-Control", "no-cache")
	ctx.SetResponseHeader("Connection", "keep-alive")
	logger := log.GetLogger().With(zap.Any("streamName", param.StreamName))
	ctx.Stream(func(w io.Writer) bool {
		defer streamFinishedHandler(param, consumer)
		cancelCh := ctx.Done()
		closeCh := make(<-chan bool)
		notify, ok := w.(http.CloseNotifier)
		err := ctx.FlushResult(w, []byte(""))
		if err != nil {
			logger.Warnf("flush stream result failed, error: %s", err.Error())
			return false
		}
		if ok {
			closeCh = notify.CloseNotify()
		}
		startTime := time.Now()
		for {
			select {
			case _, ok = <-cancelCh:
				if !ok {
					logger.Warnf("cancel channel is closed")
				}
				logger.Warnf("subscribe request of stream client is canceled")
				return false
			// This case takes effect only when ginCtx is used.
			// When the client is disconnected, return and close consumer.
			case _, ok = <-closeCh:
				if !ok {
					logger.Warnf("close channel is closed")
				}
				logger.Warnf("http connection of stream client is closed")
				return false
			default:
			}
			elements, errorInfo := receiveElements(consumer, defaultSubscribeTimeoutMs, param.ExpectReceiveNum)
			if errorInfo != nil {
				logger.Warnf("receive stream error: %s,element size: %d", errorInfo.Error(), len(elements))
				break
			}
			if len(elements) == 0 && time.Now().Sub(startTime).Milliseconds() <= int64(param.TimeoutMs) {
				continue
			}
			if len(elements) == 0 {
				logger.Warnf("receive stream failed,element size is zero")
				break
			}
			startTime = time.Now()
			result, streamEnd := processElements(elements, consumer, logger)
			err := ctx.FlushResult(w, result)
			if err != nil {
				logger.Warnf("flush stream result failed, error: %s,bytes: %s", err.Error(), result)
				if strings.Contains(err.Error(), "connection closed") ||
					strings.Contains(err.Error(), "connection was forcibly closed by the remote host") {
					return true
				}
			}
			if streamEnd {
				logger.Infof("receive end element,close consumer")
				return false
			}
		}
		return false
	})
}

func streamFinishedHandler(param SubscribeParam, consumer api.StreamConsumer) {
	logger := log.GetLogger().With(zap.Any("streamName", param.StreamName))
	if err := consumer.Close(); err != nil {
		logger.Errorf("failed to close consumer %s", err.Error())
	}
	logger.Infof("consumer close success")
	if param.Callback != nil {
		param.Callback()
	}
}

func processElements(elements []api.Element, consumer api.StreamConsumer, logger api.FormatLogger) ([]byte, bool) {
	var (
		result    []byte
		streamEnd bool
	)
	for i := range elements {
		sh := reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(elements[i].Ptr)),
			Len:  int(elements[i].Size),
			Cap:  int(elements[i].Size),
		}
		bytes := *(*[]byte)(unsafe.Pointer(&sh))
		logger.Debugf("receive stream with element: %s, size, %d", string(bytes), elements[i].Size)
		consumer.Ack(elements[i].Id)
		if elements[i].Size == StreamEndElementSize &&
			string(bytes) == StreamEndElement {
			streamEnd = true
			break
		}
		result = append(result, bytes...)
	}
	result = append(result, '\n')
	return result, streamEnd
}

func receiveElements(consumer api.StreamConsumer, timeoutMs uint32, expectReceiveNum int32) ([]api.Element, error) {
	if expectReceiveNum <= 0 {
		return consumer.Receive(timeoutMs)
	} else {
		expectNum := uint32(expectReceiveNum)
		return consumer.ReceiveExpectNum(expectNum, timeoutMs)
	}
}

func encryptData(cfg *Config, data []byte) ([]byte, error) {
	var res []byte
	return res, nil
}

func decryptData(cfg *Config, data []byte) ([]byte, error) {
	var res []byte
	return res, nil
}
