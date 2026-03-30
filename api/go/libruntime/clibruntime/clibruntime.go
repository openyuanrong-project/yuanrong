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

/*
This package clibruntime encapsulates all cgo invocations.
*/
package clibruntime

/*
#cgo CFLAGS: -I../cpplibruntime
#cgo LDFLAGS: -L../../../../build/output/runtime/service/go/bin -lcpplibruntime
#include <stdlib.h>
#include <string.h>
#include "clibruntime.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/uuid"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

// KvClientImpl -
type KvClientImpl struct {
	stateStore *StateStore
}

// StateStore -
type StateStore struct {
	mutex         *sync.RWMutex
	stateStorePtr C.CStateStorePtr
}

func kvClientCheckNil(ptr *KvClientImpl) api.ErrorInfo {
	if ptr == nil {
		return api.ErrorInfo{Code: api.DsClientNilError, Err: errors.New("client is nil")}
	}
	if ptr.stateStore == nil {
		return api.ErrorInfo{Code: api.DsClientNilError, Err: errors.New("stateStore is nil")}
	}
	return api.ErrorInfo{Code: api.Ok, Err: nil}
}

func stateStorePtrCheckNil(ptr *StateStore) api.ErrorInfo {
	if ptr.stateStorePtr == nil {
		return api.ErrorInfo{Code: api.DsClientNilError, Err: errors.New("cStateStorePtr is nil")}
	}
	return api.ErrorInfo{Code: api.Ok, Err: nil}
}

// CreateClient -
func CreateClient(config api.ConnectArguments) (api.KvClient, error) {
	arguments := C.CConnectArguments{
		host:                      CSafeString(config.Host),
		port:                      C.int(config.Port),
		timeoutMs:                 C.int(config.TimeoutMs),
		token:                     CSafeBytes(config.Token),
		tokenLen:                  C.int(len(config.Token)),
		clientPublicKey:           CSafeString(config.ClientPublicKey),
		clientPublicKeyLen:        C.int(len(config.ClientPublicKey)),
		clientPrivateKey:          CSafeBytes(config.ClientPrivateKey),
		clientPrivateKeyLen:       C.int(len(config.ClientPrivateKey)),
		serverPublicKey:           CSafeString(config.ServerPublicKey),
		serverPublicKeyLen:        C.int(len(config.ServerPublicKey)),
		accessKey:                 CSafeString(config.AccessKey),
		accessKeyLen:              C.int(len(config.AccessKey)),
		secretKey:                 CSafeBytes(config.SecretKey),
		secretKeyLen:              C.int(len(config.SecretKey)),
		authClientID:              CSafeString(config.AuthclientID),
		authClientIDLen:           C.int(len(config.AuthclientID)),
		authClientSecret:          CSafeBytes(config.AuthclientSecret),
		authClientSecretLen:       C.int(len(config.AuthclientSecret)),
		authUrl:                   CSafeString(config.AuthURL),
		authUrlLen:                C.int(len(config.AuthURL)),
		tenantID:                  CSafeString(config.TenantID),
		tenantIDLen:               C.int(len(config.TenantID)),
		enableCrossNodeConnection: C.char(btoi(config.EnableCrossNodeConnection)),
	}
	defer freeCArguments(&arguments)
	var s StateStore
	cErr := C.CCreateStateStore(&arguments, &s.stateStorePtr)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, "create kv client: ")
	}
	s.mutex = new(sync.RWMutex)
	return &KvClientImpl{stateStore: &s}, nil
}

func messageFree(cErr C.CErrorInfo) string {
	msg := CSafeGoString(cErr.message)
	CSafeFree(cErr.message)
	return msg
}

func codeNotZeroErr(code int, cErr C.CErrorInfo, str string) error {
	msg := messageFree(cErr)
	return api.ErrorInfo{Code: code, Err: fmt.Errorf(str+"%s", msg)}
}

func codeNotZeroDsErr(code int, cErr C.CErrorInfo, str string) api.ErrorInfo {
	msg := messageFree(cErr)
	return api.ErrorInfo{
		Code: int(cErr.dsStatusCode), Err: api.ErrorInfo{Code: code, Err: fmt.Errorf(str+"%s", msg)},
	}
}

// KVSet -
func (c *KvClientImpl) KVSet(key string, value []byte, param api.SetParam) api.ErrorInfo {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return status
	}
	cKey := CSafeString(key)
	defer CSafeFree(cKey)

	cValue, cValueLen := ByteSliceToCBinaryDataNoCopy(value)
	cBuf := C.CBuffer{
		buffer:              cValue,
		size_buffer:         C.int64_t(cValueLen),
		selfSharedPtrBuffer: nil,
	}
	cParam := cSetParam(param)

	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore)
	if status.IsError() {
		return status
	}

	cErr := C.CSetByStateStore(c.stateStore.stateStorePtr, cKey, cBuf, cParam)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroDsErr(code, cErr, "kv set: ")
	}
	return api.ErrorInfo{Code: api.Ok, Err: nil}
}

// KVSetWithoutKey -
func (c *KvClientImpl) KVSetWithoutKey(value []byte, param api.SetParam) (string, api.ErrorInfo) {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return "", status
	}
	cValue, cValueLen := ByteSliceToCBinaryDataNoCopy(value)
	cBuf := C.CBuffer{
		buffer:              cValue,
		size_buffer:         C.int64_t(cValueLen),
		selfSharedPtrBuffer: nil,
	}
	var cKey *C.char = nil
	defer func() {
		CSafeFree(cKey)
	}()
	var cKeyLen C.int
	cParam := cSetParam(param)

	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore)
	if status.IsError() {
		return "", status
	}

	cErr := C.CSetValueByStateStore(c.stateStore.stateStorePtr, cBuf, cParam, &cKey, &cKeyLen)
	code := int(cErr.code)
	if code != 0 {
		return "", codeNotZeroDsErr(code, cErr, "kv set value: ")
	}
	return CSafeGoStringN(cKey, cKeyLen), api.ErrorInfo{Code: api.Ok, Err: nil}
}

// KVGet -
func (c *KvClientImpl) KVGet(key string, timeoutMs ...uint32) ([]byte, api.ErrorInfo) {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return nil, status
	}
	cTimeoutMs := C.int(0)
	if len(timeoutMs) > 0 {
		cTimeoutMs = C.int(timeoutMs[0])
	}
	var cData C.CBuffer
	cKey := CSafeString(key)
	defer CSafeFree(cKey)

	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore)
	if status.IsError() {
		return nil, status
	}

	cErr := C.CGetByStateStore(c.stateStore.stateStorePtr, cKey, &cData, cTimeoutMs)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroDsErr(code, cErr, "kv get: ")
	}
	defer CSafeFree((*C.char)(cData.buffer))
	return C.GoBytes(cData.buffer, C.int(cData.size_buffer)), api.ErrorInfo{Code: api.Ok, Err: nil}
}

// GetCredential -
func GetCredential() api.Credential {
	cCredential := C.CGetCredential()
	defer CSafeFree(cCredential.ak)
	defer CSafeFree(cCredential.dk)
	defer CSafeFree(cCredential.sk)
	return GoCredential(cCredential)
}

// KVGetMulti -
func (c *KvClientImpl) KVGetMulti(keys []string, timeoutMs ...uint32) ([][]byte, api.ErrorInfo) {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return nil, status
	}
	cTimeoutMs := C.int(0)
	if len(timeoutMs) > 0 {
		cTimeoutMs = C.int(timeoutMs[0])
	}
	cKeys, cKeysLen := CStrings(keys)
	defer freeCStrings(cKeys, cKeysLen)
	cBuffersSlice := make([]C.CBuffer, len(keys))
	cBufferPtr := (*C.CBuffer)(unsafe.Pointer(&cBuffersSlice[0]))

	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore)
	if status.IsError() {
		return nil, status
	}

	cErr := C.CGetArrayByStateStore(c.stateStore.stateStorePtr, cKeys, cKeysLen, cBufferPtr, cTimeoutMs)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroDsErr(code, cErr, "kv get array: ")
	}
	values := make([][]byte, len(keys))
	for i, val := range cBuffersSlice {
		values[i] = C.GoBytes(val.buffer, C.int(val.size_buffer))
		CSafeFree((*C.char)(val.buffer))
	}
	return values, api.ErrorInfo{Code: api.Ok, Err: nil}
}

// KVQuerySize -
func (c *KvClientImpl) KVQuerySize(keys []string) ([]uint64, api.ErrorInfo) {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return nil, status
	}
	cKeys, cKeysLen := CStrings(keys)
	defer freeCStrings(cKeys, cKeysLen)
	sizes := make([]uint64, len(keys))
	cSizes := (*C.uint64_t)(unsafe.Pointer(&sizes[0]))

	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore)
	if status.IsError() {
		return nil, status
	}

	cErr := C.CQuerySizeByStateStore(c.stateStore.stateStorePtr, cKeys, cKeysLen, cSizes)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroDsErr(code, cErr, "kv query size: ")
	}
	return sizes, api.ErrorInfo{Code: api.Ok, Err: nil}
}

// KVDel -
func (c *KvClientImpl) KVDel(key string) api.ErrorInfo {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return status
	}
	cKey := CSafeString(key)
	defer CSafeFree(cKey)

	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore)
	if status.IsError() {
		return status
	}

	cErr := C.CDelByStateStore(c.stateStore.stateStorePtr, cKey)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroDsErr(code, cErr, "kv del: ")
	}
	return api.ErrorInfo{Code: api.Ok, Err: nil}
}

// KVDelMulti -
func (c *KvClientImpl) KVDelMulti(keys []string) ([]string, api.ErrorInfo) {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return keys, status
	}
	cKeys, cKeysLen := CStrings(keys)
	defer freeCStrings(cKeys, cKeysLen)
	var cFailedKeys **C.char
	var cFailedKeysLen C.int

	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore) // double check
	if status.IsError() {
		return keys, status
	}

	cErr := C.CDelArrayByStateStore(c.stateStore.stateStorePtr, cKeys, cKeysLen, &cFailedKeys, &cFailedKeysLen)
	var failedKeys []string
	if int(cFailedKeysLen) > 0 {
		failedKeys = GoStrings(cFailedKeysLen, cFailedKeys)
	}
	code := int(cErr.code)
	if code != 0 {
		return failedKeys, codeNotZeroDsErr(code, cErr, "kv del array: ")
	}
	return failedKeys, api.ErrorInfo{Code: api.Ok, Err: nil}
}

// GenerateKey -
func (c *KvClientImpl) GenerateKey() string {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return ""
	}
	var cKey *C.char = nil
	defer func() {
		CSafeFree(cKey)
	}()
	var cKeyLen C.int
	c.stateStore.mutex.RLock()
	defer c.stateStore.mutex.RUnlock()
	status = stateStorePtrCheckNil(c.stateStore)
	if status.IsError() {
		return ""
	}

	cErr := C.CGenerateKey(c.stateStore.stateStorePtr, &cKey, &cKeyLen)
	if code := int(cErr.code); code != 0 {
		messageFree(cErr)
		return ""
	}
	return CSafeGoStringN(cKey, cKeyLen)
}

// SetTraceID -
func (c *KvClientImpl) SetTraceID(traceID string) {
	cTraceID := CSafeString(traceID)
	defer CSafeFree(cTraceID)
	traceIDLen := C.int(len(traceID))
	cErr := C.CSetTraceId(cTraceID, traceIDLen)
	code := int(cErr.code)
	if code != 0 {
		messageFree(cErr)
	}
}

// DestroyClient -
func (c *KvClientImpl) DestroyClient() {
	status := kvClientCheckNil(c)
	if status.IsError() {
		return
	}
	c.stateStore.mutex.Lock()
	defer c.stateStore.mutex.Unlock()
	C.CDestroyStateStore(c.stateStore.stateStorePtr)
	c.stateStore.stateStorePtr = nil
}

func freeCArguments(arguments *C.CConnectArguments) {
	CSafeFree(arguments.host)
	CSafeFree(arguments.token)
	CSafeFree(arguments.clientPublicKey)
	CSafeFree(arguments.clientPrivateKey)
	CSafeFree(arguments.serverPublicKey)
	CSafeFree(arguments.accessKey)
	CSafeFree(arguments.secretKey)
	CSafeFree(arguments.authClientID)
	CSafeFree(arguments.authClientSecret)
	CSafeFree(arguments.authUrl)
	CSafeFree(arguments.tenantID)
}

// StreamProducerImpl struct represents a producer of streaming data.
type StreamProducerImpl struct {
	producer C.Producer_p
}

// StreamConsumerImpl struct represents a consumer of streaming data.
type StreamConsumerImpl struct {
	consumer C.Consumer_p
}

// CSafeString -
func CSafeString(s string) *C.char {
	if len(s) == 0 {
		return nil
	}
	return C.CString(s)
}

// CSafeFree -
func CSafeFree(s *C.char) {
	if s == nil {
		return
	}
	C.free(unsafe.Pointer(s))
}

// CSafeBytes -
func CSafeBytes(b []byte) *C.char {
	if len(b) == 0 {
		return nil
	}
	return (*C.char)(C.CBytes(b))
}

// CSafeGoBytes -
func CSafeGoBytes(cStr *C.char, length C.int) []byte {
	if cStr == nil {
		return nil
	}
	byteArr := C.GoBytes(unsafe.Pointer(cStr), length)
	return byteArr
}

// CSafeGoString safely converts a *C.char to a Go string.
func CSafeGoString(message *C.char) string {
	if message == nil {
		return ""
	} else {
		return C.GoString(message)
	}
}

// CSafeGoStringN -
func CSafeGoStringN(message *C.char, length C.int) string {
	if message == nil || length == 0 {
		return ""
	} else {
		return C.GoStringN(message, length)
	}
}

// Send sends an element.
// This method can be used to send data to consumers.
func (p *StreamProducerImpl) Send(element api.Element) error {
	cPtr := (*C.uint8_t)(unsafe.Pointer(element.Ptr))
	cErr := C.CProducerSend(p.producer, cPtr, C.uint64_t(element.Size), C.uint64_t(element.Id))
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "stream producer send: ")
	}
	return nil
}

// SendWithTimeout sends an element with a specified timeout.
// If the element cannot be sent within the timeout duration, it will be discarded.
// Parameters:
// - element: the element to be sent.
// - timeoutMs: the duration to wait before discarding the element.
func (p *StreamProducerImpl) SendWithTimeout(element api.Element, timeoutMs int64) error {
	cPtr := (*C.uint8_t)(unsafe.Pointer(element.Ptr))
	cErr := C.CProducerSendWithTimeout(
		p.producer, cPtr, C.uint64_t(element.Size), C.uint64_t(element.Id), C.int64_t(timeoutMs),
	)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "stream producer send with timeout: ")
	}
	return nil
}

// Close signals the producer to stop accepting new data and automatically flushes
// any pending data in the buffer. Once closed, the producer is no longer available.
func (p *StreamProducerImpl) Close() error {
	cErr := C.CProducerClose(p.producer)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "stream producer close: ")
	}
	return nil
}

// Receive retrieves data from the consumer with an optional timeout.
// Parameters:
// - timeoutMs: Maximum time in milliseconds to wait for data before timing out.
// Returns:
// - []api.Element: The received data.
// - error: nil if data was received within the timeout, error otherwise.
func (c *StreamConsumerImpl) Receive(timeoutMs uint32) ([]api.Element, error) {
	var count C.uint64_t
	var pEles *C.CElement
	cErr := C.CConsumerReceive(c.consumer, C.uint32_t(timeoutMs), &pEles, &count)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, "stream consumer receive: ")
	}
	eles := GoEles(pEles, count)
	return eles, nil
}

// ReceiveExpectNum waits to receive the expected number of elements from the consumer,
// either until the timeout is reached or the expected number of elements are received.
// Parameters:
// - expectNum: The expected number of elements to receive.
// - timeoutMs: Maximum time in milliseconds to wait before timing out.
// Returns:
// - []api.Element: The received data.
// - error: nil if data was received within the timeout, error otherwise.
func (c *StreamConsumerImpl) ReceiveExpectNum(expectNum uint32, timeoutMs uint32) ([]api.Element, error) {
	var count C.uint64_t
	var pEles *C.CElement
	cErr := C.CConsumerReceiveExpectNum(c.consumer, C.uint32_t(expectNum), C.uint32_t(timeoutMs), &pEles, &count)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, fmt.Sprintf("stream consumer receive expect num: %d", expectNum))
	}
	eles := GoEles(pEles, count)
	return eles, nil
}

// Ack confirms that the consumer has completed processing the element identified by elementID.
// This function signals to other workers whether the consumer has finished processing the element.
// If all consumers have acknowledged processing the element, it triggers internal memory reclamation
// for the corresponding page.
// Parameters:
// - elementID: The identifier of the element that has been consumed.
func (c *StreamConsumerImpl) Ack(elementId uint64) error {
	cErr := C.CConsumerAck(c.consumer, C.uint64_t(elementId))
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "stream consumer ack: ")
	}
	return nil
}

// Close closes the consumer, unsubscribing it from further data consumption.
// This method also acknowledges any unacknowledged elements on the consumer,
// ensuring that they are marked as processed before shutting down.
func (c *StreamConsumerImpl) Close() error {
	cErr := C.CConsumerClose(c.consumer)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "stream close: ")
	}
	return nil
}

var (
	cfg               config.Config
	getAsyncCallbacks sync.Map
	getEventCallbacks sync.Map
	rawCallbacks      sync.Map
)

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func itob(i int) bool {
	if i > 0 {
		return true
	}
	return false
}

func checkIfRef(t api.ArgType) C.char {
	if t == api.ObjectRef {
		return C.char(1)
	}
	return C.char(0)
}

func apiTypeToCApiType(apiType api.ApiType) C.CApiType {
	switch apiType {
	case api.ActorApi:
		return C.ACTOR
	case api.FaaSApi:
		return C.FAAS
	case api.PosixApi:
		return C.POSIX
	default:
		return C.POSIX
	}
}

// Init Initialization entry, which is used to initialize the data system and function system.
func Init(conf config.Config) error {
	cfg = conf

	cFunctionSystemAddress := C.CString(conf.FunctionSystemAddress)
	defer C.free(unsafe.Pointer(cFunctionSystemAddress))
	cGrpcAddress := C.CString(conf.GrpcAddress)
	defer C.free(unsafe.Pointer(cGrpcAddress))
	cDataSystemAddress := C.CString(conf.DataSystemAddress)
	defer C.free(unsafe.Pointer(cDataSystemAddress))
	cIamAddress := C.CString(conf.IamAddress)
	defer C.free(unsafe.Pointer(cIamAddress))

	cJobId := C.CString(conf.JobID)
	defer C.free(unsafe.Pointer(cJobId))
	cRuntimeId := C.CString(conf.RuntimeID)
	defer C.free(unsafe.Pointer(cRuntimeId))
	cInstanceId := C.CString(conf.InstanceID)
	defer C.free(unsafe.Pointer(cInstanceId))
	cFunctionName := C.CString(conf.FunctionName)
	defer C.free(unsafe.Pointer(cFunctionName))
	cLogLevel := C.CString(conf.LogLevel)
	defer C.free(unsafe.Pointer(cLogLevel))
	cLogDir := C.CString(conf.LogDir)
	defer C.free(unsafe.Pointer(cLogDir))
	cServerName := C.CString(conf.ServerName)
	defer C.free(unsafe.Pointer(cServerName))
	cNs := C.CString(conf.Namespace)
	defer C.free(unsafe.Pointer(cNs))
	cPrivateKeyPath := C.CString(conf.PrivateKeyPath)
	defer C.free(unsafe.Pointer(cPrivateKeyPath))
	cCertificateFilePath := C.CString(conf.CertificateFilePath)
	defer C.free(unsafe.Pointer(cCertificateFilePath))
	cVerifyFilePath := C.CString(conf.VerifyFilePath)
	defer C.free(unsafe.Pointer(cVerifyFilePath))
	cPrivateKeyPaaswd := C.CString(conf.PrivateKeyPaaswd)
	defer C.free(unsafe.Pointer(cPrivateKeyPaaswd))
	cSystemAuthAccessKey := C.CString(conf.SystemAuthAccessKey)
	defer C.free(unsafe.Pointer(cSystemAuthAccessKey))
	cSystemAuthSecretKey := C.CString(conf.SystemAuthSecretKey)
	defer C.free(unsafe.Pointer(cSystemAuthSecretKey))
	cSystemAuthSecretKeySize := C.int(len(conf.SystemAuthSecretKey))
	cSystemAuthDataKey := C.CString(conf.SystemAuthDataKey)
	defer C.free(unsafe.Pointer(cSystemAuthDataKey))
	cSystemAuthDataKeySize := C.int(len(conf.SystemAuthDataKey))
	cEncryptPrivateKeyPasswd := C.CString(conf.EncryptPrivateKeyPasswd)
	defer C.free(unsafe.Pointer(cEncryptPrivateKeyPasswd))
	cPrimaryKeyStoreFile := C.CString(conf.PrimaryKeyStoreFile)
	defer C.free(unsafe.Pointer(cPrimaryKeyStoreFile))
	cStandbyKeyStoreFile := C.CString(conf.StandbyKeyStoreFile)
	defer C.free(unsafe.Pointer(cStandbyKeyStoreFile))
	cRuntimePublicKeyContextPath := C.CString(conf.RuntimePublicKeyContextPath)
	defer C.free(unsafe.Pointer(cRuntimePublicKeyContextPath))
	cRuntimePrivateKeyContextPath := C.CString(conf.RuntimePrivateKeyContextPath)
	defer C.free(unsafe.Pointer(cRuntimePrivateKeyContextPath))
	cDsPublicKeyContextPath := C.CString(conf.DsPublicKeyContextPath)
	defer C.free(unsafe.Pointer(cDsPublicKeyContextPath))
	cMaxConcurrencyCreateNum := C.int(conf.MaxConcurrencyCreateNum)

	cFunctionId := C.CString(conf.FunctionId)
	defer C.free(unsafe.Pointer(cFunctionId))
	cConf := C.CLibruntimeConfig{
		functionSystemAddress:        cFunctionSystemAddress,
		grpcAddress:                  cGrpcAddress,
		dataSystemAddress:            cDataSystemAddress,
		iamAddress:                   cIamAddress,
		jobId:                        cJobId,
		runtimeId:                    cRuntimeId,
		instanceId:                   cInstanceId,
		functionName:                 cFunctionName,
		logLevel:                     cLogLevel,
		logDir:                       cLogDir,
		functionId:                   cFunctionId,
		apiType:                      apiTypeToCApiType(conf.Api),
		inCluster:                    C.char(btoi(conf.InCluster)),
		isDriver:                     C.char(btoi(conf.IsDriver)),
		enableMTLS:                   C.char(btoi(conf.EnableMTLS)),
		privateKeyPath:               cPrivateKeyPath,
		certificateFilePath:          cCertificateFilePath,
		verifyFilePath:               cVerifyFilePath,
		privateKeyPaaswd:             cPrivateKeyPaaswd,
		systemAuthAccessKey:          cSystemAuthAccessKey,
		systemAuthSecretKey:          cSystemAuthSecretKey,
		systemAuthSecretKeySize:      cSystemAuthSecretKeySize,
		systemAuthDataKey:            cSystemAuthDataKey,
		systemAuthDataKeySize:        cSystemAuthDataKeySize,
		encryptPrivateKeyPasswd:      cEncryptPrivateKeyPasswd,
		primaryKeyStoreFile:          cPrimaryKeyStoreFile,
		standbyKeyStoreFile:          cStandbyKeyStoreFile,
		enableDsEncrypt:              C.char(btoi(conf.EnableDsEncrypt)),
		runtimePublicKeyContextPath:  cRuntimePublicKeyContextPath,
		runtimePrivateKeyContextPath: cRuntimePrivateKeyContextPath,
		dsPublicKeyContextPath:       cDsPublicKeyContextPath,
		maxConcurrencyCreateNum:      cMaxConcurrencyCreateNum,
		enableSigaction:              C.char(btoi(conf.EnableSigaction)),
		enableEvent:                  C.char(btoi(conf.EnableEvent)),
	}
	cErr := C.CInit(&cConf)
	code := int(cErr.code)
	if code != 0 {
		return fmt.Errorf("failed to init libruntime, code: %d, message: %s", code, messageFree(cErr))
	}
	return nil
}

// ReceiveRequestLoop begins loop processing the received request.
func ReceiveRequestLoop() {
	C.CReceiveRequestLoop()
}

// ExecShutdownHandler exec shutdown handler.
func ExecShutdownHandler(signum int) {
	cSignum := C.int(signum)
	C.CExecShutdownHandler(cSignum)
}

// CreateInstance Golang posix function.
func CreateInstance(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	cFuncMeta := cFunctionMeta(funcMeta)
	defer freeCFunctionMeta(cFuncMeta)

	cArgs, cArgsLen := cArgs(args)
	defer freeCArgs(cArgs, cArgsLen)

	cInvokeOpt := cInvokeOptions(invokeOpt)
	defer freeCInvokeOptions(cInvokeOpt)

	var cInstanceID *C.char
	cErr := C.CCreateInstance(cFuncMeta, cArgs, cArgsLen, cInvokeOpt, &cInstanceID)
	code := int(cErr.code)
	if code != 0 {
		return "", codeNotZeroErr(code, cErr, "")
	}
	defer C.free(unsafe.Pointer(cInstanceID))

	objectID := CSafeGoString(cInstanceID)
	wait := make(chan error, 1)
	WaitAsync(
		objectID, func(result []byte, err error) {
			wait <- err
		},
	)
	defer func() {
		if _, err := GDecreaseRef([]string{objectID}); err != nil {
			fmt.Printf("failed to decrease object ref,err: %s", err.Error())
		}
	}()

	var createErr error
	timer := time.NewTimer(time.Duration(invokeOpt.Timeout) * time.Second)
	select {
	case <-timer.C:
		createErr = api.ErrorInfo{Code: 3002, Err: fmt.Errorf("create instance timeout")}
	case err, ok := <-wait:
		if !ok {
			createErr = api.ErrorInfo{Code: 3002, Err: fmt.Errorf("failed to create instance")}
		} else {
			createErr = err
		}
	}
	var instanceID string
	cRealInstanceID := C.CGetRealInstanceId(cInstanceID, C.int(invokeOpt.Timeout))
	instanceID = CSafeGoString(cRealInstanceID)
	C.free(unsafe.Pointer(cRealInstanceID))
	if instanceID == "" || instanceID == objectID {
		if createErr == nil {
			return "", api.ErrorInfo{Code: 1003, Err: fmt.Errorf("real instance id not exist, get failed")}
		} else {
			return "", createErr // avoid sending an fake insId(i.e. objectId) to scheduler
		}
	}
	return instanceID, createErr
}

// InvokeByInstanceId Golang posix function
func InvokeByInstanceId(
	funcMeta api.FunctionMeta, instanceID string, args []api.Arg, invokeOpt api.InvokeOptions,
) (string, error) {
	cFuncMeta := cFunctionMeta(funcMeta)
	defer freeCFunctionMeta(cFuncMeta)

	cInstanceID := C.CString(instanceID)
	defer C.free(unsafe.Pointer(cInstanceID))

	cArgs, cArgsLen := cArgs(args)
	defer freeCArgs(cArgs, cArgsLen)

	cInvokeOpt := cInvokeOptions(invokeOpt)
	defer freeCInvokeOptions(cInvokeOpt)

	var cRetObjID *C.char
	cErr := C.CInvokeByInstanceId(cFuncMeta, cInstanceID, cArgs, cArgsLen, cInvokeOpt, &cRetObjID)
	code := int(cErr.code)
	if code != 0 {
		return "", codeNotZeroErr(code, cErr, "invoke by instance id: ")
	}
	retObjID := CSafeGoString(cRetObjID)
	C.free(unsafe.Pointer(cRetObjID))
	return retObjID, nil
}

// InvokeByFunctionName Supports system functions and faas.
func InvokeByFunctionName(funcMeta api.FunctionMeta, args []api.Arg, invokeOpt api.InvokeOptions) (string, error) {
	cFuncMeta := cFunctionMeta(funcMeta)

	defer freeCFunctionMeta(cFuncMeta)

	cArgs, cArgsLen := cArgs(args)
	defer freeCArgs(cArgs, cArgsLen)

	cInvokeOpt := cInvokeOptions(invokeOpt)
	defer freeCInvokeOptions(cInvokeOpt)

	var cRetObjID *C.char
	cErr := C.CInvokeByFunctionName(cFuncMeta, cArgs, cArgsLen, cInvokeOpt, &cRetObjID)
	code := int(cErr.code)
	if code != 0 {
		return "", codeNotZeroErr(code, cErr, "")
	}
	retObjID := CSafeGoString(cRetObjID)
	C.free(unsafe.Pointer(cRetObjID))
	return retObjID, nil
}

// AcquireInstance Supports system functions and faas.
func AcquireInstance(stateID string, funcMeta api.FunctionMeta, acquireOpt api.InvokeOptions) (api.InstanceAllocation,
	error) {
	cFuncMeta := cFunctionMeta(funcMeta)
	defer freeCFunctionMeta(cFuncMeta)

	cInvokeOpt := cAcquireOptions(acquireOpt)
	defer freeCInvokeOptions(cInvokeOpt)

	cStateID := C.CString(stateID)
	defer C.free(unsafe.Pointer(cStateID))
	instanceAllocation := new(C.CInstanceAllocation)
	defer freeCInstanceAllocation(instanceAllocation)
	cErr := C.CAcquireInstance(cStateID, cFuncMeta, cInvokeOpt, instanceAllocation)
	code := int(cErr.code)
	if code != 0 {
		return api.InstanceAllocation{}, codeNotZeroErr(code, cErr, "")
	}
	return api.InstanceAllocation{
		FuncSig:       CSafeGoString(instanceAllocation.funcSig),
		FuncKey:       CSafeGoString(instanceAllocation.functionId),
		InstanceID:    CSafeGoString(instanceAllocation.instanceId),
		LeaseID:       CSafeGoString(instanceAllocation.leaseId),
		LeaseInterval: int64(instanceAllocation.tLeaseInterval),
	}, nil
}

// ReleaseInstance release lease
func ReleaseInstance(allocation api.InstanceAllocation, stateID string, abnormal bool, option api.InvokeOptions) {
	cInstanceAllocation := cCInstanceAllocation(allocation)
	cOptions := cInvokeOptions(option)
	defer freeCInstanceAllocation(cInstanceAllocation)
	defer freeCInvokeOptions(cOptions)
	cStateID := C.CString(stateID)
	defer C.free(unsafe.Pointer(cStateID))
	cErr := C.CReleaseInstance(cInstanceAllocation, cStateID, C.char(btoi(abnormal)), cOptions)
	if code := int(cErr.code); code != 0 {
		fmt.Printf("failed to release lease %s  error: %s\n", allocation.LeaseID, messageFree(cErr))
	}
}

// Kill instances
func Kill(instanceID string, signo int, data []byte) error {
	cInstanceID := C.CString(instanceID)
	defer C.free(unsafe.Pointer(cInstanceID))

	cSigno := C.int(signo)

	cData, cDataLen := ByteSliceToCBinaryData(data)
	if cData != nil {
		defer C.free(cData)
	}

	cBuf := C.CBuffer{
		buffer:      cData,
		size_buffer: C.int64_t(cDataLen),
	}
	cErr := C.CKill(cInstanceID, cSigno, cBuf)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "kill instance: ")
	}
	return nil
}

// UpdateSchdulerInfo update libruntime scheduler info
func UpdateSchdulerInfo(schedulerName string, schedulerId string, option string) {
	cSchedulerName := C.CString(schedulerName)
	defer C.free(unsafe.Pointer(cSchedulerName))
	cSchedulerId := C.CString(schedulerId)
	defer C.free(unsafe.Pointer(cSchedulerId))
	cOption := C.CString(option)
	defer C.free(unsafe.Pointer(cOption))

	C.CUpdateSchdulerInfo(cSchedulerName, cSchedulerId, cOption)
}

// GetAsync with a callback
func GetAsync(objectID string, cb api.GetAsyncCallback) {
	getAsyncCallbacks.Store(objectID, cb)
	cObjectID := C.CString(objectID)
	defer C.free(unsafe.Pointer(cObjectID))
	C.CGetAsync(cObjectID, nil)
}

// GetEvent with a callback
func GetEvent(objectID string, cb api.GetEventCallback) {
	getEventCallbacks.Store(objectID, cb)
	cObjectID := C.CString(objectID)
	defer C.free(unsafe.Pointer(cObjectID))
	C.CGetEvent(cObjectID, nil)
}

// DeleteGetEventCallback -
func DeleteGetEventCallback(objectID string) {
	getEventCallbacks.Delete(objectID)
}

// WaitAsync with a callback
func WaitAsync(objectID string, cb api.GetAsyncCallback) {
	getAsyncCallbacks.Store(objectID, cb)
	cObjectID := C.CString(objectID)
	defer C.free(unsafe.Pointer(cObjectID))
	C.CWaitAsync(cObjectID, nil)
}

// CreateStreamProducer creates and returns a new Producer instance
func CreateStreamProducer(streamName string, producerConf api.ProducerConf) (api.StreamProducer, error) {
	cStreamName := C.CString(streamName)
	defer C.free(unsafe.Pointer(cStreamName))
	cProducerConf := cProducerConfig(producerConf)
	defer C.free(unsafe.Pointer(cProducerConf.traceId))
	var streamProducer StreamProducerImpl
	cErr := C.CCreateStreamProducer(cStreamName, cProducerConf, &streamProducer.producer)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, "create stream producer: ")
	}
	return &streamProducer, nil
}

// CreateStreamConsumer creates and returns a new Consumer instance
func CreateStreamConsumer(streamName string, config api.SubscriptionConfig) (api.StreamConsumer, error) {
	cStreamName := C.CString(streamName)
	defer C.free(unsafe.Pointer(cStreamName))
	cSubscriptConf := cSubscriptionConfig(config)
	defer C.free(unsafe.Pointer(cSubscriptConf.subscriptionName))
	defer C.free(unsafe.Pointer(cSubscriptConf.traceId))
	var streamConsumer StreamConsumerImpl
	cErr := C.CCreateStreamConsumer(cStreamName, cSubscriptConf, &streamConsumer.consumer)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, "create stream consumer: ")
	}
	return &streamConsumer, nil
}

// DeleteStream Delete a data flow.
// When the number of global producers and consumers is 0,
// the data flow is no longer used and the metadata related to the data flow is deleted from each worker and the
// master. This function can be invoked on any host node.
func DeleteStream(streamName string) error {
	cStreamName := C.CString(streamName)
	defer C.free(unsafe.Pointer(cStreamName))
	cErr := C.CDeleteStream(cStreamName)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "delete stream: ")
	}
	return nil
}

// QueryGlobalProducersNum Specifies the flow name to query the number of all producers of the flow.
func QueryGlobalProducersNum(streamName string) (uint64, error) {
	cStreamName := C.CString(streamName)
	defer C.free(unsafe.Pointer(cStreamName))

	var num C.uint64_t
	cErr := C.CQueryGlobalProducersNum(cStreamName, &num)
	code := int(cErr.code)
	if code != 0 {
		return 0, codeNotZeroErr(code, cErr, "qurey global producers num: ")
	}
	return uint64(num), nil
}

// QueryGlobalConsumersNum Specifies the flow name to query the number of all consumers of the flow.
func QueryGlobalConsumersNum(streamName string) (uint64, error) {
	cStreamName := C.CString(streamName)
	defer C.free(unsafe.Pointer(cStreamName))

	var num C.uint64_t
	cErr := C.CQueryGlobalConsumersNum(cStreamName, &num)
	code := int(cErr.code)
	if code != 0 {
		return 0, codeNotZeroErr(code, cErr, "qurey global consumers num: ")
	}
	return uint64(num), nil
}

// GoGetAsyncCallback is exported as a C function for calling from C/C++ code.
// The purpose is to execute the get callback function of go.
//
//export GoGetAsyncCallback
func GoGetAsyncCallback(cObjectID *C.char, cBuf C.CBuffer, cErr *C.CErrorInfo, userData unsafe.Pointer) {
	objectID := CSafeGoString((*C.char)(cObjectID))
	cb, ok := getGetAsyncCallback(objectID)
	if !ok {
		return
	}

	code := int(cErr.code)
	if code != 0 {
		cb([]byte{}, codeNotZeroErr(code, *cErr, ""))
		return
	}

	cb(C.GoBytes(cBuf.buffer, C.int(cBuf.size_buffer)), nil)
	C.free(cBuf.buffer)
}

// GoWaitAsyncCallback is exported as a C function for calling from C/C++ code.
// The purpose is to execute the wait callback function of go.
//
//export GoWaitAsyncCallback
func GoWaitAsyncCallback(cObjectID *C.char, cErr *C.CErrorInfo, userData unsafe.Pointer) {
	objectID := CSafeGoString((*C.char)(cObjectID))
	cb, ok := getGetAsyncCallback(objectID)
	if !ok {
		return
	}

	if code := int(cErr.code); code != 0 {
		cb([]byte{}, codeNotZeroErr(code, *cErr, ""))
		return
	}
	cb([]byte{}, nil)
}

func getGetAsyncCallback(objectID string) (api.GetAsyncCallback, bool) {
	value, ok := getAsyncCallbacks.LoadAndDelete(objectID)
	if ok {
		cb, ok := value.(api.GetAsyncCallback)
		return cb, ok
	}
	return nil, false
}

// GoGetEventCallback is exported as a C function for calling from C/C++ code.
// The purpose is to execute the get callback function of go.
//
//export GoGetEventCallback
func GoGetEventCallback(cObjectID *C.char, cBuf C.CBuffer, cErr *C.CErrorInfo, userData unsafe.Pointer) {
	objectID := CSafeGoString((*C.char)(cObjectID))
	cb, ok := getGetEventCallback(objectID)
	if !ok {
		return
	}

	code := int(cErr.code)
	if code != 0 {
		cb([]byte{}, codeNotZeroErr(code, *cErr, ""))
		return
	}

	cb(C.GoBytes(cBuf.buffer, C.int(cBuf.size_buffer)), nil)
	C.free(cBuf.buffer)
}

func getGetEventCallback(objectID string) (api.GetEventCallback, bool) {
	value, ok := getEventCallbacks.Load(objectID)
	if ok {
		cb, ok := value.(api.GetEventCallback)
		return cb, ok
	}
	return nil, false
}

// RawCallback define the raw callback function type.
type RawCallback func(result []byte, err error)

// GoRawCallback is exported as a C function for calling from C/C++ code.
// The purpose is to execute the raw callback function of go.
//
//export GoRawCallback
func GoRawCallback(cKey *C.char, cErr C.CErrorInfo, cResultRaw C.CBuffer) {
	key := CSafeGoString((*C.char)(cKey))
	cb, ok := getRawCallback(key)
	if !ok {
		return
	}

	code := int(cErr.code)
	if code != 0 {
		cb([]byte{}, codeNotZeroErr(code, cErr, "raw callback error: "))
		return
	}

	cb(C.GoBytes(cResultRaw.buffer, C.int(cResultRaw.size_buffer)), nil)
}

func getRawCallback(key string) (RawCallback, bool) {
	value, ok := rawCallbacks.LoadAndDelete(key)
	if ok {
		cb, ok := value.(RawCallback)
		return cb, ok
	}
	return nil, false
}

// CreateInstanceRaw Raw interface provided for the frontend.
func CreateInstanceRaw(createReqRaw []byte) ([]byte, error) {
	createReqRawPtr, createReqRawLen := ByteSliceToCBinaryDataNoCopy(createReqRaw)
	cCreateReqRaw := C.CBuffer{
		buffer:              createReqRawPtr,
		size_buffer:         C.int64_t(createReqRawLen),
		selfSharedPtrBuffer: nil,
	}

	errChan := make(chan error, 1)
	key := uuid.New().String()
	var result []byte
	var rawCallback RawCallback = func(resultRaw []byte, err error) {
		if err == nil {
			result = resultRaw
		}
		errChan <- err
	}
	rawCallbacks.Store(key, rawCallback)

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.CCreateInstanceRaw(cCreateReqRaw, cKey)

	resultErr, ok := <-errChan
	if !ok {
		return []byte{}, api.ErrorInfo{Code: -1, Err: fmt.Errorf("create instance raw: channel closed")}
	}
	if resultErr != nil {
		return []byte{}, resultErr
	}

	return result, nil
}

// InvokeByInstanceIdRaw Raw interface provided for the frontend.
func InvokeByInstanceIdRaw(invokeReqRaw []byte) ([]byte, error) {
	invokeReqRawPtr, invokeReqRawLen := ByteSliceToCBinaryDataNoCopy(invokeReqRaw)
	cInvokeReqRaw := C.CBuffer{
		buffer:              invokeReqRawPtr,
		size_buffer:         C.int64_t(invokeReqRawLen),
		selfSharedPtrBuffer: nil,
	}

	errChan := make(chan error, 1)
	var result []byte
	key := uuid.New().String()
	var callback RawCallback = func(resultRaw []byte, err error) {
		if err == nil {
			result = resultRaw
		}
		errChan <- err
	}
	rawCallbacks.Store(key, callback)

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.CInvokeByInstanceIdRaw(cInvokeReqRaw, cKey)

	resultErr, ok := <-errChan
	if !ok {
		return []byte{}, api.ErrorInfo{Code: -1, Err: fmt.Errorf("invoke raw: channel closed")}
	}
	if resultErr != nil {
		return []byte{}, resultErr
	}

	return result, nil
}

// KillRaw Raw interface provided for the frontend.
func KillRaw(killReqRaw []byte) ([]byte, error) {
	killReqRawPtr, killReqRawLen := ByteSliceToCBinaryDataNoCopy(killReqRaw)
	cKillReqRaw := C.CBuffer{
		buffer:              killReqRawPtr,
		size_buffer:         C.int64_t(killReqRawLen),
		selfSharedPtrBuffer: nil,
	}

	errChan := make(chan error, 1)
	var result []byte
	key := uuid.New().String()
	var rawCallback RawCallback = func(resultRaw []byte, e error) {
		if e == nil {
			result = resultRaw
		}
		errChan <- e
	}
	rawCallbacks.Store(key, rawCallback)
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.CKillRaw(cKillReqRaw, cKey)

	resultErr, ok := <-errChan
	if !ok {
		return []byte{}, api.ErrorInfo{Code: -1, Err: fmt.Errorf("kill raw: channel closed")}
	}
	if resultErr != nil {
		return []byte{}, resultErr
	}

	return result, nil
}

// Exit send exit request.
func Exit(code int, message string) {
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	C.CExit(C.int(code), cMessage)
}

// Finalize func for go sdk
// This API is used to release resources, such as created function instances and data objects,
// to prevent residual resources.
func Finalize() {
	C.CFinalize()
}

// KVSet save binary data to the data system.
func KVSet(key string, value []byte, param api.SetParam) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	cValue, cValueLen := ByteSliceToCBinaryDataNoCopy(value)
	cBuf := C.CBuffer{
		buffer:              cValue,
		size_buffer:         C.int64_t(cValueLen),
		selfSharedPtrBuffer: nil,
	}
	cParam := cSetParam(param)
	cErr := C.CKVWrite(cKey, cBuf, cParam)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "kv set: ")
	}
	return nil
}

// KVMSetTx save binary datas to the data system.
func KVMSetTx(keys []string, values [][]byte, param api.MSetParam) error {
	cKeys, cKeysLen := CStrings(keys)
	defer freeCStrings(cKeys, cKeysLen)

	cBuffersSlice := make([]C.CBuffer, len(values))
	for idx, val := range values {
		cValue, cValueLen := ByteSliceToCBinaryDataNoCopy(val)
		cBuf := C.CBuffer{
			buffer:              cValue,
			size_buffer:         C.int64_t(cValueLen),
			selfSharedPtrBuffer: nil,
		}
		cBuffersSlice[idx] = cBuf
	}
	cBufferPtr := (*C.CBuffer)(unsafe.Pointer(&cBuffersSlice[0]))
	cParam := cMSetParam(param)
	cErr := C.CKVMSetTx(cKeys, cKeysLen, cBufferPtr, cParam)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "kv multi set tx: ")
	}
	return nil
}

// KVGet get binary data from data system.
func KVGet(key string, timeoutms uint) ([]byte, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	var cData C.CBuffer
	cErr := C.CKVRead(cKey, C.int(timeoutms), &cData)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, "kv get: ")
	}
	value := C.GoBytes(cData.buffer, C.int(cData.size_buffer))
	C.free(cData.buffer)
	return value, nil
}

// KVGetMulti get multi binary data from data system.
func KVGetMulti(keys []string, timeoutms uint) ([][]byte, error) {
	cKeys, cKeysLen := CStrings(keys)
	defer freeCStrings(cKeys, cKeysLen)

	cBuffersSlice := make([]C.CBuffer, len(keys))
	cBufferPtr := (*C.CBuffer)(unsafe.Pointer(&cBuffersSlice[0]))
	cErr := C.CKVMultiRead(cKeys, cKeysLen, C.int(timeoutms), 1, cBufferPtr)
	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, "kv get multi: ")
	}

	values := make([][]byte, len(keys))
	for idx, val := range cBuffersSlice {
		values[idx] = C.GoBytes(val.buffer, C.int(val.size_buffer))
		C.free(val.buffer)
	}
	return values, nil
}

// KVDel del data from data system.
func KVDel(key string) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cErr := C.CKVDel(cKey)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "kv del: ")
	}
	return nil
}

// KVDelMulti del multi data from data system.
func KVDelMulti(keys []string) ([]string, error) {
	cKeys, cKeysLen := CStrings(keys)
	defer freeCStrings(cKeys, cKeysLen)

	var cFailedKeys **C.char
	var cFailedKeysLen C.int
	cErr := C.CKMultiVDel(cKeys, cKeysLen, &cFailedKeys, &cFailedKeysLen)
	var failedKeys []string
	if int(cFailedKeysLen) > 0 {
		failedKeys = GoStrings(cFailedKeysLen, cFailedKeys)
	}
	code := int(cErr.code)
	if code != 0 {
		return failedKeys, codeNotZeroErr(code, cErr, "kv del multi: ")
	}
	return failedKeys, nil
}

// SetTenantID -
func SetTenantID(tenantId string) error {
	cTenantId := CSafeString(tenantId)
	defer CSafeFree(cTenantId)
	tenantIdLen := C.int(len(tenantId))
	cErr := C.CSetTenantId(cTenantId, tenantIdLen)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "set tenantID failed: ")
	}
	return nil
}

// PutCommon put obj data to data system,
// Parameters:
//   - objectID: object id.
//   - value: object binary data.
//   - param: put param.
//   - nestedObjectIDs: Indicates that the current object will be used by objects in nestedIds.
//     If the object in nestedIds is not deleted,
//     the current object will be held by the system until all objects in nestedIds are deleted.
//
// Returns:
// - error: nil if data put success, error otherwise.
func PutCommon(objectID string, value []byte, param api.PutParam, isRaw bool, nestedObjectIDs ...string) error {
	cObjID := C.CString(objectID)
	defer C.free(unsafe.Pointer(cObjID))
	cValuePtr, cValueLen := ByteSliceToCBinaryDataNoCopy(value)
	cData := C.CBuffer{
		buffer:              cValuePtr,
		size_buffer:         C.int64_t(cValueLen),
		selfSharedPtrBuffer: nil,
	}
	cNestedIDs, cNestedIDsLen := CStrings(nestedObjectIDs)
	defer freeCStrings(cNestedIDs, cNestedIDsLen)

	var cErr C.CErrorInfo
	cIsRaw := C.char(0)
	if isRaw {
		cIsRaw = C.char(1)
	}
	createParam := cCreateParam(param)
	cErr = C.CPutCommon(cObjID, cData, cNestedIDs, cNestedIDsLen, cIsRaw, createParam)

	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "put: ")
	}
	return nil
}

// Put put obj data to data system.
func Put(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	return PutCommon(objectID, value, param, false, nestedObjectIDs...)
}

// PutRaw put obj data to data system.
func PutRaw(objectID string, value []byte, param api.PutParam, nestedObjectIDs ...string) error {
	return PutCommon(objectID, value, param, true, nestedObjectIDs...)
}

// GetCommon get objs from data system.
func GetCommon(objectIDs []string, timeoutMs int, isRaw bool) ([][]byte, error) {
	cObjIDs, cObjIDsLen := CStrings(objectIDs)
	defer freeCStrings(cObjIDs, cObjIDsLen)

	cBuffersSlice := make([]C.CBuffer, len(objectIDs))
	cBufferPtr := (*C.CBuffer)(unsafe.Pointer(&cBuffersSlice[0]))

	var cErr C.CErrorInfo
	cIsRaw := C.char(0)
	if isRaw {
		cIsRaw = C.char(1)
	}
	cErr = C.CGetMultiCommon(cObjIDs, cObjIDsLen, C.int(timeoutMs), 1, cBufferPtr, cIsRaw)

	code := int(cErr.code)
	if code != 0 {
		return nil, codeNotZeroErr(code, cErr, "")
	}

	values := make([][]byte, len(objectIDs))
	for idx, val := range cBuffersSlice {
		values[idx] = C.GoBytes(val.buffer, C.int(val.size_buffer))
		C.free(val.buffer)
	}
	return values, nil
}

// Get to get objs from data system.
func Get(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return GetCommon(objectIDs, timeoutMs, false)
}

// GetRaw get objs from data system.
func GetRaw(objectIDs []string, timeoutMs int) ([][]byte, error) {
	return GetCommon(objectIDs, timeoutMs, true)
}

// Wait until result return or timeout
func Wait(objectIDs []string, waitNum uint64, timeout int) (
	readyObjectIds, unReadyObjectIds []string, errors map[string]error,
) {
	cObjIDs, cObjIDsLen := CStrings(objectIDs)
	defer freeCStrings(cObjIDs, cObjIDsLen)

	waitResult := C.CWaitResult{
		readyIds:        nil,
		size_readyIds:   0,
		unreadyIds:      nil,
		size_unreadyIds: 0,
		errorIds:        nil,
		size_errorIds:   0,
	}
	C.CWait(cObjIDs, cObjIDsLen, C.int(waitNum), C.int(timeout), &waitResult)

	if int(waitResult.size_readyIds) > 0 {
		readyObjectIds = GoStrings(waitResult.size_readyIds, waitResult.readyIds)
	}

	if int(waitResult.size_unreadyIds) > 0 {
		unReadyObjectIds = GoStrings(waitResult.size_unreadyIds, waitResult.unreadyIds)
	}

	if int(waitResult.size_errorIds) > 0 {
		errors = make(map[string]error)
		errorIds := unsafe.Slice(waitResult.errorIds, int(waitResult.size_errorIds))
		defer C.free(unsafe.Pointer(waitResult.errorIds))
		for _, errorId := range errorIds {
			var errCode C.int
			var errorMessage *C.char
			var objectId *C.char
			var stackTracesInfo C.CStackTracesInfo
			C.CParseCErrorObjectPointer(errorId, &errCode, &errorMessage, &objectId, &stackTracesInfo)
			code := int(errCode)
			msg := CSafeGoString(errorMessage)
			C.free(unsafe.Pointer(errorMessage))
			id := CSafeGoString(objectId)
			C.free(unsafe.Pointer(objectId))
			sinfo := goStackTraces(stackTracesInfo)
			errors[id] = api.NewErrInfo(code, msg, sinfo)
			C.free(unsafe.Pointer(errorId))
		}
	}
	return
}

// GIncreaseRefCommon increase object reference count
func GIncreaseRefCommon(objectIDs []string, isRaw bool, remoteClientID ...string) ([]string, error) {
	cObjIDs, cObjIDsLen := CStrings(objectIDs)
	defer freeCStrings(cObjIDs, cObjIDsLen)

	var cRemoteID *C.char = nil
	if len(remoteClientID) > 0 {
		cRemoteID = C.CString(remoteClientID[0])
		defer C.free(unsafe.Pointer(cRemoteID))
	}
	var cFailedIDs **C.char
	var cFailedIDsLen C.int
	cIsRaw := C.char(0)
	if isRaw {
		cIsRaw = C.char(1)
	}

	var cErr C.CErrorInfo
	cErr = C.CIncreaseReferenceCommon(cObjIDs, cObjIDsLen, cRemoteID, &cFailedIDs, &cFailedIDsLen, cIsRaw)

	var failedIDs []string
	if int(cFailedIDsLen) > 0 {
		failedIDs = GoStrings(cFailedIDsLen, cFailedIDs)
	}
	code := int(cErr.code)
	if code != 0 {
		return failedIDs, codeNotZeroErr(code, cErr, "global increase ref: ")
	}
	return failedIDs, nil
}

// GIncreaseRef increase object reference count
func GIncreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return GIncreaseRefCommon(objectIDs, false, remoteClientID...)
}

// GIncreaseRefRaw increase object reference count
func GIncreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return GIncreaseRefCommon(objectIDs, true, remoteClientID...)
}

// GDecreaseRefCommon decrease object reference count
func GDecreaseRefCommon(objectIDs []string, isRaw bool, remoteClientID ...string) ([]string, error) {
	cObjIDs, cObjIDsLen := CStrings(objectIDs)
	defer freeCStrings(cObjIDs, cObjIDsLen)

	var cRemoteID *C.char = nil
	if len(remoteClientID) > 0 {
		cRemoteID = C.CString(remoteClientID[0])
		defer C.free(unsafe.Pointer(cRemoteID))
	}
	var cFailedIDs **C.char
	var cFailedIDsLen C.int

	var cErr C.CErrorInfo

	cIsRaw := C.char(0)
	if isRaw {
		cIsRaw = C.char(1)
	}
	cErr = C.CDecreaseReferenceCommon(cObjIDs, cObjIDsLen, cRemoteID, &cFailedIDs, &cFailedIDsLen, cIsRaw)

	var failedIDs []string
	if int(cFailedIDsLen) > 0 {
		failedIDs = GoStrings(cFailedIDsLen, cFailedIDs)
	}
	code := int(cErr.code)
	if code != 0 {
		return failedIDs, codeNotZeroErr(code, cErr, "global decrease ref: ")
	}
	return failedIDs, nil
}

// GDecreaseRef decrease object reference count
func GDecreaseRef(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return GDecreaseRefCommon(objectIDs, false, remoteClientID...)
}

// GDecreaseRefRaw decrease object reference count
func GDecreaseRefRaw(objectIDs []string, remoteClientID ...string) ([]string, error) {
	return GDecreaseRefCommon(objectIDs, true, remoteClientID...)
}

// ReleaseGRefs release object refs by remote client id
func ReleaseGRefs(remoteClientID string) error {
	var cRemoteID *C.char = nil
	cRemoteID = C.CString(remoteClientID)
	defer C.free(unsafe.Pointer(cRemoteID))
	cErr := C.CReleaseGRefs(cRemoteID)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "global decrease ref: ")
	}
	return nil
}

// AllocReturnObject Creates an object and applies for a memory block.
// Computing operations can be performed on the memory block.
// will return a 'Buffer' that will be used to manipulate the memory
func AllocReturnObject(do *config.DataObject, size uint, nestedIds []string, totalNativeBufferSize *uint) error {
	cDataObject := C.CDataObject{
		id:                C.CString(do.ID),
		selfSharedPtr:     do.CSharedPtr,
		nestedObjIds:      nil,
		size_nestedObjIds: 0,
	}
	defer CSafeFree(cDataObject.id)
	cNestedIDs, cNestedIDsLen := CStrings(nestedIds)
	defer freeCStrings(cNestedIDs, cNestedIDsLen)

	var cTotalNativeBufferSize C.uint64_t = C.uint64_t(*totalNativeBufferSize)
	cErr := C.CAllocReturnObject(&cDataObject, C.int(size), cNestedIDs, cNestedIDsLen, &cTotalNativeBufferSize)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "alloc return object: ")
	}
	do.Buffer = config.DataBuffer{
		Ptr:             cDataObject.buffer.buffer,
		Size:            int(cDataObject.buffer.size_buffer),
		SharedPtrBuffer: cDataObject.buffer.selfSharedPtrBuffer,
	}
	*totalNativeBufferSize = uint(cTotalNativeBufferSize)
	return nil
}

// SetReturnObject if return by message, set return object
func SetReturnObject(do *config.DataObject, size uint) {
	cDataObject := C.CDataObject{
		id:                C.CString(do.ID),
		selfSharedPtr:     do.CSharedPtr,
		nestedObjIds:      nil,
		size_nestedObjIds: 0,
	}
	defer CSafeFree(cDataObject.id)
	C.CSetReturnObject(&cDataObject, C.int(size))
	do.Buffer = config.DataBuffer{
		Ptr:             cDataObject.buffer.buffer,
		Size:            int(cDataObject.buffer.size_buffer),
		SharedPtrBuffer: cDataObject.buffer.selfSharedPtrBuffer,
	}
}

// WriterLatch Obtains the write lock of the buffer object.
func WriterLatch(do *config.DataObject) error {
	cBuf := C.CBuffer{
		buffer:              do.Buffer.Ptr,
		size_buffer:         C.int64_t(do.Buffer.Size),
		selfSharedPtrBuffer: do.Buffer.SharedPtrBuffer,
	}
	cErr := C.CWriterLatch(&cBuf)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "buffer writer latch: ")
	}
	return nil
}

// MemoryCopy Writes data to a buffer object.
func MemoryCopy(do *config.DataObject, src []byte) error {
	cBuf := C.CBuffer{
		buffer:              do.Buffer.Ptr,
		size_buffer:         C.int64_t(do.Buffer.Size),
		selfSharedPtrBuffer: do.Buffer.SharedPtrBuffer,
	}
	if len(src) == 0 {
		return nil
	}
	cSrcPtr, cSrcLen := ByteSliceToCBinaryDataNoCopy(src)
	cErr := C.CMemoryCopy(&cBuf, cSrcPtr, C.uint64_t(cSrcLen))
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "buffer memory copy: ")
	}
	return nil
}

// Seal Publish the object and seal it. Sealed objects cannot be modified again.
func Seal(do *config.DataObject) error {
	cBuf := C.CBuffer{
		buffer:              do.Buffer.Ptr,
		size_buffer:         C.int64_t(do.Buffer.Size),
		selfSharedPtrBuffer: do.Buffer.SharedPtrBuffer,
	}
	cErr := C.CSeal(&cBuf)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "buffer seal: ")
	}
	return nil
}

// WriterUnlatch release the write lock of the buffer object.
func WriterUnlatch(do *config.DataObject) error {
	cBuf := C.CBuffer{
		buffer:              do.Buffer.Ptr,
		size_buffer:         C.int64_t(do.Buffer.Size),
		selfSharedPtrBuffer: do.Buffer.SharedPtrBuffer,
	}
	cErr := C.CWriterUnlatch(&cBuf)
	code := int(cErr.code)
	if code != 0 {
		return codeNotZeroErr(code, cErr, "buffer writer unlatch: ")
	}
	return nil
}

func cSetParam(param api.SetParam) C.CSetParam {
	return C.CSetParam{
		writeMode: cWriteMode(param.WriteMode),
		ttlSecond: C.uint32_t(param.TTLSecond),
		existence: cExistenceOpt(param.Existence),
		cacheType: cCacheType(param.CacheType),
	}
}

func cMSetParam(param api.MSetParam) C.CMSetParam {
	return C.CMSetParam{
		writeMode: cWriteMode(param.WriteMode),
		ttlSecond: C.uint32_t(param.TTLSecond),
		existence: cExistenceOpt(param.Existence),
		cacheType: cCacheType(param.CacheType),
	}
}

func cCreateParam(param api.PutParam) C.CCreateParam {
	return C.CCreateParam{
		writeMode:       cWriteMode(param.WriteMode),
		consistencyType: cConsistencyType(param.ConsistencyType),
		cacheType:       cCacheType(param.CacheType),
	}
}

func cWriteMode(wm api.WriteModeEnum) C.CWriteMode {
	switch wm {
	case api.NoneL2Cache:
		return C.NONE_L2_CACHE
	case api.WriteThroughL2Cache:
		return C.WRITE_THROUGH_L2_CACHE
	case api.WriteBackL2Cache:
		return C.WRITE_BACK_L2_CACHE
	case api.NoneL2CacheEvict:
		return C.NONE_L2_CACHE_EVICT
	default:
		return C.NONE_L2_CACHE
	}
}

func cExistenceOpt(o int32) C.CExistenceOpt {
	if o == 0 {
		return C.NONE
	}
	return C.NX
}

func cCacheType(ct api.CacheTypeEnum) C.CCacheType {
	switch ct {
	case api.MEMORY:
		return C.MEMORY
	case api.DISK:
		return C.DISK
	default:
		return C.MEMORY
	}
}

func cConsistencyType(ct api.ConsistencyTypeEnum) C.CConsistencyType {
	switch ct {
	case api.PRAM:
		return C.PRAM
	case api.CAUSAL:
		return C.CAUSAL
	default:
		return C.PRAM
	}
}

func cStringOptional(str *string) (*C.char, C.char) {
	if str == nil {
		return nil, 0
	}
	return C.CString(*str), 1
}

func cFunctionMeta(funcMeta api.FunctionMeta) *C.CFunctionMeta {
	cName, hasName := cStringOptional(funcMeta.Name)
	cNamespace, hasNamespace := cStringOptional(funcMeta.Namespace)
	cFuncMeta := C.CFunctionMeta{
		appName:      C.CString(funcMeta.AppName),
		moduleName:   nil,
		funcName:     C.CString(funcMeta.FuncName),
		className:    nil,
		functionId:   C.CString(funcMeta.FuncID),
		languageType: C.int(funcMeta.Language),
		signature:    C.CString(funcMeta.Sig),
		poolLabel:    C.CString(funcMeta.PoolLabel),
		apiType:      C.CApiType(funcMeta.Api),
		hasName:      hasName,
		name:         cName,
		hasNs:        hasNamespace,
		ns:           cNamespace,
		codeId:       nil,
	}
	return &cFuncMeta
}

func freeCFunctionMeta(cFuncMeta *C.CFunctionMeta) {
	CSafeFree(cFuncMeta.appName)
	CSafeFree(cFuncMeta.moduleName)
	CSafeFree(cFuncMeta.funcName)
	CSafeFree(cFuncMeta.className)
	CSafeFree(cFuncMeta.functionId)
	CSafeFree(cFuncMeta.signature)
	CSafeFree(cFuncMeta.poolLabel)
	CSafeFree(cFuncMeta.name)
	CSafeFree(cFuncMeta.ns)
	CSafeFree(cFuncMeta.codeId)
}

func cArgs(args []api.Arg) (*C.CInvokeArg, C.int) {
	var cArgs *C.CInvokeArg = nil
	argsLen := len(args)
	if argsLen != 0 {
		cArgs = (*C.CInvokeArg)(C.malloc(C.size_t(argsLen) * C.sizeof_CInvokeArg))
		cArgsSlice := unsafe.Slice(cArgs, argsLen)
		for idx, val := range args {
			ptr, length := ByteSliceToCBinaryDataNoCopy(val.Data)
			nestedPtr, nestedLen := CStrings(val.NestedObjectIDs)
			cTenantId := C.CString(val.TenantID)
			cArgsSlice[idx] = C.CInvokeArg{
				buf:                ptr,
				size_buf:           C.int64_t(length),
				isRef:              checkIfRef(val.Type),
				objId:              nil,
				nestedObjects:      nestedPtr,
				size_nestedObjects: nestedLen,
				tenantId:           cTenantId,
			}
		}
	}
	return cArgs, C.int(argsLen)
}

func freeCArgs(cArgs *C.CInvokeArg, cArgsLen C.int) {
	argsLen := int(cArgsLen)
	if argsLen == 0 {
		return
	}
	cArgsSlice := unsafe.Slice(cArgs, argsLen)
	for idx, _ := range cArgsSlice {
		freeCStrings(cArgsSlice[idx].nestedObjects, cArgsSlice[idx].size_nestedObjects)
		C.free(unsafe.Pointer(cArgsSlice[idx].tenantId))
	}
	C.free(unsafe.Pointer(cArgs))
}

func cAcquireOptions(acquireOpt api.InvokeOptions) *C.CInvokeOptions {
	cSchedInstIDs, cSchedInstIDsLen := CStrings(acquireOpt.SchedulerInstanceIDs)
	cAcquireOpt := C.CInvokeOptions{
		schedulerFunctionId:       CSafeString(acquireOpt.SchedulerFunctionID),
		schedulerInstanceIds:      cSchedInstIDs,
		size_schedulerInstanceIds: cSchedInstIDsLen,
		traceId:                   CSafeString(acquireOpt.TraceID),
		timeout:                   C.int(acquireOpt.Timeout),
		acquireTimeout:            C.int(acquireOpt.AcquireTimeout),
		trafficLimited:            C.char(btoi(acquireOpt.TrafficLimited)),
	}
	return &cAcquireOpt
}

func cCInstanceAllocation(instanceAllocation api.InstanceAllocation) *C.CInstanceAllocation {
	cInstanceAlloc := C.CInstanceAllocation{
		functionId:     C.CString(instanceAllocation.FuncKey),
		funcSig:        C.CString(instanceAllocation.FuncSig),
		instanceId:     C.CString(instanceAllocation.InstanceID),
		leaseId:        C.CString(instanceAllocation.LeaseID),
		tLeaseInterval: C.int(instanceAllocation.LeaseInterval),
	}
	return &cInstanceAlloc
}

func freeCInstanceAllocation(cInstanceAllocation *C.CInstanceAllocation) {
	CSafeFree(cInstanceAllocation.funcSig)
	CSafeFree(cInstanceAllocation.functionId)
	CSafeFree(cInstanceAllocation.instanceId)
	CSafeFree(cInstanceAllocation.leaseId)
}

func cInvokeOptions(invokeOpt api.InvokeOptions) *C.CInvokeOptions {
	cRes, cResLen := cCustomResources(invokeOpt.CustomResources)
	cExts, cExtsLen := cCustomExtensions(invokeOpt.CustomExtensions)
	cCreate, cCreateLen := cCreateOpt(invokeOpt.CreateOpt)
	cLabels, cLabelsLen := CStrings(invokeOpt.Labels)
	cSchedAffs, cSchedAffsLen := cScheduleAffinities(invokeOpt.ScheduleAffinities)
	cCodePaths, cCodePathsLen := CStrings(invokeOpt.CodePaths)
	cSchedInstIDs, cSchedInstIDsLen := CStrings(invokeOpt.SchedulerInstanceIDs)
	cIvkLabel, cIvkLabelLen := cInvokeLabels(invokeOpt.InvokeLabels)
	cInvokeOpt := C.CInvokeOptions{
		cpu:                       C.int(invokeOpt.Cpu),
		memory:                    C.int(invokeOpt.Memory),
		customResources:           cRes,
		size_customResources:      cResLen,
		customExtensions:          cExts,
		size_customExtensions:     cExtsLen,
		createOpt:                 cCreate,
		size_createOpt:            cCreateLen,
		labels:                    cLabels,
		size_labels:               cLabelsLen,
		schedAffinities:           cSchedAffs,
		size_schedAffinities:      cSchedAffsLen,
		codePaths:                 cCodePaths,
		size_codePaths:            cCodePathsLen,
		schedulerFunctionId:       C.CString(invokeOpt.SchedulerFunctionID),
		schedulerInstanceIds:      cSchedInstIDs,
		size_schedulerInstanceIds: cSchedInstIDsLen,
		traceId:                   C.CString(invokeOpt.TraceID),
		timeout:                   C.int(invokeOpt.Timeout),
		acquireTimeout:            C.int(invokeOpt.AcquireTimeout),
		RetryTimes:                C.int(invokeOpt.RetryTimes),
		RecoverRetryTimes:         C.int(invokeOpt.RecoverRetryTimes),
		invokeLabels:              cIvkLabel,
		size_invokeLabels:         cIvkLabelLen,
		scheduleTimeoutMs:         C.int64_t(invokeOpt.ScheduleTimeoutMs),
		forceInvoke:               C.char(btoi(invokeOpt.ForceInvoke)),
	}
	if invokeOpt.InstanceSession != nil {
		cCInstanceSession := (*C.CInstanceSession)(C.malloc(C.sizeof_CInstanceSession))
		cCInstanceSession.sessionId = C.CString(invokeOpt.InstanceSession.SessionID)
		cCInstanceSession.sessionTtl = C.int(invokeOpt.InstanceSession.SessionTTL)
		cCInstanceSession.concurrency = C.int(invokeOpt.InstanceSession.Concurrency)
		cInvokeOpt.instanceSession = cCInstanceSession
	}

	return &cInvokeOpt
}

func freeCInvokeOptions(cInvokeOpt *C.CInvokeOptions) {
	freeCCustomResources(cInvokeOpt.customResources, cInvokeOpt.size_customResources)
	freeCCustomExtensions(cInvokeOpt.customExtensions, cInvokeOpt.size_customExtensions)
	freeCCreateOpt(cInvokeOpt.createOpt, cInvokeOpt.size_createOpt)
	freeCStrings(cInvokeOpt.labels, cInvokeOpt.size_labels)
	freeCScheduleAffinities(cInvokeOpt.schedAffinities, cInvokeOpt.size_schedAffinities)
	freeCStrings(cInvokeOpt.codePaths, cInvokeOpt.size_codePaths)
	CSafeFree(cInvokeOpt.schedulerFunctionId)
	freeCStrings(cInvokeOpt.schedulerInstanceIds, cInvokeOpt.size_schedulerInstanceIds)
	CSafeFree(cInvokeOpt.traceId)
	freeCInvokeLabels(cInvokeOpt.invokeLabels, cInvokeOpt.size_invokeLabels)
	if unsafe.Pointer(cInvokeOpt.instanceSession) != nil {
		CSafeFree(cInvokeOpt.instanceSession.sessionId)
		C.free(unsafe.Pointer(cInvokeOpt.instanceSession))
	}
}

func cScheduleAffinities(schedAffinities []api.Affinity) (*C.CAffinity, C.int) {
	length := len(schedAffinities)
	if length == 0 {
		return nil, 0
	}
	cSchedAffs := (*C.CAffinity)(C.malloc(C.size_t(length) * C.sizeof_CAffinity))
	cSchedAffsSlice := unsafe.Slice(cSchedAffs, length)
	for idx, val := range schedAffinities {
		cSchedAffsSlice[idx].affKind = cAffinityKind(val.Kind)
		cSchedAffsSlice[idx].affType = cAffinityType(val.Affinity)
		cSchedAffsSlice[idx].preferredPrio = C.char(btoi(val.PreferredPriority))
		cSchedAffsSlice[idx].preferredAntiOtherLabels = C.char(btoi(val.PreferredAntiOtherLabels))
		cLabelOps, cLabelOpsLen := cLabelOperators(val.LabelOps)
		cSchedAffsSlice[idx].labelOps = cLabelOps
		cSchedAffsSlice[idx].size_labelOps = cLabelOpsLen
	}
	return cSchedAffs, (C.int)(length)
}

func freeCScheduleAffinities(cSchedAffs *C.CAffinity, cLen C.int) {
	length := int(cLen)
	if length == 0 {
		return
	}
	cSchedAffsSlice := unsafe.Slice(cSchedAffs, length)
	for idx, _ := range cSchedAffsSlice {
		freeCLabelOperators(cSchedAffsSlice[idx].labelOps, cSchedAffsSlice[idx].size_labelOps)
	}
	C.free(unsafe.Pointer(cSchedAffs))
}

func cLabelOperators(labelOps []api.LabelOperator) (*C.CLabelOperator, C.int) {
	length := len(labelOps)
	if length == 0 {
		return nil, 0
	}
	cLabelOps := (*C.CLabelOperator)(C.malloc(C.size_t(length) * C.sizeof_CLabelOperator))
	cLabelOpsSlice := unsafe.Slice(cLabelOps, length)
	for idx, val := range labelOps {
		cLabelOpsSlice[idx].opType = cLabelOpType(val.Type)
		if val.LabelKey != "" {
			cLabelOpsSlice[idx].labelKey = C.CString(val.LabelKey)
		} else {
			cLabelOpsSlice[idx].labelKey = nil
		}
		cLabelVals, cLabelValsLen := CStrings(val.LabelValues)
		cLabelOpsSlice[idx].labelValues = cLabelVals
		cLabelOpsSlice[idx].size_labelValues = cLabelValsLen
	}
	return cLabelOps, (C.int)(length)
}

func freeCLabelOperators(cLabelOps *C.CLabelOperator, cLen C.int) {
	length := int(cLen)
	if length == 0 {
		return
	}
	cLabelOpsSlice := unsafe.Slice(cLabelOps, length)
	for idx := range length {
		CSafeFree(cLabelOpsSlice[idx].labelKey)
		freeCStrings(cLabelOpsSlice[idx].labelValues, cLabelOpsSlice[idx].size_labelValues)
	}
	C.free(unsafe.Pointer(cLabelOps))
}

func cLabelOpType(opType api.OperatorType) C.CLabelOpType {
	switch opType {
	case api.LabelOpIn:
		return C.IN
	case api.LabelOpNotIn:
		return C.NOT_IN
	case api.LabelOpExists:
		return C.EXISTS
	case api.LabelOpNotExists:
		return C.NOT_EXISTS
	default:
		return C.IN
	}
}

func cAffinityKind(kind api.AffinityKindType) C.CAffinityKind {
	switch kind {
	case api.AffinityKindResource:
		return C.RESOURCE
	case api.AffinityKindInstance:
		return C.INSTANCE
	default:
		return C.INSTANCE
	}
}

func cAffinityType(affType api.AffinityType) C.CAffinityType {
	switch affType {
	case api.PreferredAffinity:
		return C.PREFERRED
	case api.PreferredAntiAffinity:
		return C.PREFERRED_ANTI
	case api.RequiredAffinity:
		return C.REQUIRED
	case api.RequiredAntiAffinity:
		return C.REQUIRED_ANTI
	default:
		return C.PREFERRED
	}
}

func cCustomResources(customResources map[string]float64) (*C.CCustomResource, C.int) {
	length := len(customResources)
	if length == 0 {
		return nil, 0
	}
	cCustomRscs := (*C.CCustomResource)(C.malloc(C.size_t(length) * C.sizeof_CCustomResource))
	cCustomRscsSlice := unsafe.Slice(cCustomRscs, length)
	idx := 0
	for k, v := range customResources {
		cCustomRscsSlice[idx] = C.CCustomResource{
			name:   C.CString(k),
			scalar: C.float(v),
		}
		idx++
	}
	return cCustomRscs, C.int(length)
}

func freeCCustomResources(cCustomRscs *C.CCustomResource, cLen C.int) {
	length := int(cLen)
	if length == 0 {
		return
	}
	cCustomRscsSlice := unsafe.Slice(cCustomRscs, length)
	for idx := range length {
		CSafeFree(cCustomRscsSlice[idx].name)
	}
	C.free(unsafe.Pointer(cCustomRscs))
}

func cCustomExtensions(customExtensions map[string]string) (*C.CCustomExtension, C.int) {
	length := len(customExtensions)
	if length == 0 {
		return nil, 0
	}
	cCustomExts := (*C.CCustomExtension)(C.malloc(C.size_t(length) * C.sizeof_CCustomExtension))
	cCustomExtsSlice := unsafe.Slice(cCustomExts, length)
	idx := 0
	for k, v := range customExtensions {
		cCustomExtsSlice[idx] = C.CCustomExtension{
			key:   C.CString(k),
			value: C.CString(v),
		}
		idx++
	}
	return cCustomExts, C.int(length)
}

func freeCCustomExtensions(cCustomExts *C.CCustomExtension, cLen C.int) {
	length := int(cLen)
	if length == 0 {
		return
	}
	cCustomExtsSlice := unsafe.Slice(cCustomExts, length)
	for idx := range length {
		CSafeFree(cCustomExtsSlice[idx].key)
		CSafeFree(cCustomExtsSlice[idx].value)
	}
	C.free(unsafe.Pointer(cCustomExts))
}

func cInvokeLabels(customInvokeLabels map[string]string) (*C.CInvokeLabels, C.int) {
	length := len(customInvokeLabels)
	if length == 0 {
		return nil, 0
	}
	cCustomILs := (*C.CInvokeLabels)(C.malloc(C.size_t(length) * C.sizeof_CInvokeLabels))
	cCustomILsSlice := unsafe.Slice(cCustomILs, length)
	idx := 0
	for k, v := range customInvokeLabels {
		cCustomILsSlice[idx] = C.CInvokeLabels{
			key:   C.CString(k),
			value: C.CString(v),
		}
		idx++
	}
	return cCustomILs, C.int(length)
}

func freeCInvokeLabels(cCustomILs *C.CInvokeLabels, cLen C.int) {
	length := int(cLen)
	if length == 0 {
		return
	}
	cCustomILsSlice := unsafe.Slice(cCustomILs, length)
	for idx := range length {
		CSafeFree(cCustomILsSlice[idx].key)
		CSafeFree(cCustomILsSlice[idx].value)
	}
	C.free(unsafe.Pointer(cCustomILs))
}

func cCreateOpt(createOpt map[string]string) (*C.CCreateOpt, C.int) {
	length := len(createOpt)
	if length == 0 {
		return nil, 0
	}
	cCreate := (*C.CCreateOpt)(C.malloc(C.size_t(length) * C.sizeof_CCreateOpt))
	cCreateSlice := unsafe.Slice(cCreate, length)
	idx := 0
	for k, v := range createOpt {
		cCreateSlice[idx] = C.CCreateOpt{
			key:   C.CString(k),
			value: C.CString(v),
		}
		idx++
	}
	return cCreate, C.int(length)
}

func freeCCreateOpt(cCreate *C.CCreateOpt, cLen C.int) {
	length := int(cLen)
	if length == 0 {
		return
	}
	cCreateSlice := unsafe.Slice(cCreate, length)
	for idx := range length {
		CSafeFree(cCreateSlice[idx].key)
		CSafeFree(cCreateSlice[idx].value)
	}
	C.free(unsafe.Pointer(cCreate))
}

func cProducerConfig(producerConf api.ProducerConf) *C.CProducerConfig {
	cProducerConf := C.CProducerConfig{
		delayFlushTime: C.int64_t(producerConf.DelayFlushTime),
		pageSize:       C.int64_t(producerConf.PageSize),
		maxStreamSize:  C.uint64_t(producerConf.MaxStreamSize),
		traceId:        C.CString(producerConf.TraceId),
	}
	return &cProducerConf
}

func cSubscriptionConfig(subscriptionConf api.SubscriptionConfig) *C.CSubscriptionConfig {
	cSubscriptionConf := C.CSubscriptionConfig{
		subscriptionName: C.CString(subscriptionConf.SubscriptionName),
		traceId:          C.CString(subscriptionConf.TraceId),
	}
	return &cSubscriptionConf
}

// ByteSliceToCBinaryData Copy go byte slice to c binary data
func ByteSliceToCBinaryData(data []byte) (unsafe.Pointer, int) {
	if len(data) == 0 {
		return nil, 0
	}
	return C.CBytes(data), len(data)
}

// ByteSliceToCBinaryDataNoCopy convert go byte slice to c binary data with no copy
func ByteSliceToCBinaryDataNoCopy(data []byte) (unsafe.Pointer, int) {
	if len(data) == 0 {
		return nil, 0
	}
	return unsafe.Pointer(&data[0]), len(data)
}

// StringToCBinaryDataNoCopy -
func StringToCBinaryDataNoCopy(data string) (unsafe.Pointer, int) {
	if len(data) == 0 {
		return nil, 0
	}
	p := (*reflect.StringHeader)(unsafe.Pointer(&data))
	return unsafe.Pointer(p.Data), len(data)
}

// GoStrings Copy **C.char to go []string and free **C.char
func GoStrings(count C.int, values **C.char) []string {
	length := int(count)
	if length == 0 {
		return make([]string, 0)
	}
	defer C.free(unsafe.Pointer(values))
	charSlice := unsafe.Slice(values, length)
	strSlice := make([]string, length)
	for idx, val := range charSlice {
		strSlice[idx] = CSafeGoString(val)
		defer C.free(unsafe.Pointer(val))
	}
	return strSlice
}

// GoStringsWithoutFree Copy **C.char to go []string without freeing **C.char
func GoStringsWithoutFree(count C.int, values **C.char) []string {
	length := int(count)
	if length == 0 {
		return make([]string, 0)
	}
	charSlice := unsafe.Slice(values, length)
	strSlice := make([]string, length)
	for idx, val := range charSlice {
		strSlice[idx] = CSafeGoString(val)
	}
	return strSlice
}

// CStrings Copy go []string to **C.char
func CStrings(strs []string) (**C.char, C.int) {
	strsLen := len(strs)
	if strsLen == 0 {
		return nil, 0
	}
	cStrs := (**C.char)(C.malloc(C.size_t(strsLen) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	cStrsSlice := unsafe.Slice(cStrs, strsLen)
	for idx, val := range strs {
		cStrsSlice[idx] = C.CString(val)
	}
	return cStrs, C.int(strsLen)
}

func freeCStrings(cStrs **C.char, cStrsLen C.int) {
	strsLen := int(cStrsLen)
	if strsLen == 0 {
		return
	}
	cStrsSlice := unsafe.Slice(cStrs, strsLen)
	for idx, _ := range cStrsSlice {
		C.free(unsafe.Pointer(cStrsSlice[idx]))
	}
	C.free(unsafe.Pointer(cStrs))
}

func errtoCerr(e error) *C.CErrorInfo {
	if e == nil {
		ce := (*C.CErrorInfo)(C.malloc(C.sizeof_CErrorInfo))
		ce.code = 0
		ce.message = nil
		ce.size_stackTracesInfo = 0
		return ce
	}
	return getCErrwithStackTrace(e)
}

func goStackTraces(stackTracesInfo C.CStackTracesInfo) api.StackTracesInfo {
	stackTracesLen := int(stackTracesInfo.size_stackTraces)
	stacksSlice := unsafe.Slice(stackTracesInfo.stackTraces, stackTracesLen)
	goStackTraces := make([]api.StackTrace, stackTracesLen)
	for k, v := range stacksSlice {
		goStackTraces[k].ClassName = CSafeGoString(v.className)
		goStackTraces[k].MethodName = CSafeGoString(v.methodName)
		goStackTraces[k].FileName = CSafeGoString(v.fileName)
		goStackTraces[k].LineNumber = int64(v.lineNumber)
		C.free(unsafe.Pointer(v.className))
		C.free(unsafe.Pointer(v.methodName))
		C.free(unsafe.Pointer(v.fileName))
		extensionLen := int(v.size_extensions)
		if extensionLen <= 0 {
			continue
		}
		goStackTraces[k].ExtensionInfo = make(map[string]string, extensionLen)
		extensionsSlice := unsafe.Slice(v.extensions, extensionLen)
		for _, val := range extensionsSlice {
			key := CSafeGoString(val.key)
			if key == "" {
				continue
			}
			goStackTraces[k].ExtensionInfo[key] = CSafeGoString(val.value)
			C.free(unsafe.Pointer(val.key))
			C.free(unsafe.Pointer(val.value))
		}
		C.free(unsafe.Pointer(v.extensions))
	}
	info := api.StackTracesInfo{
		Code:        int(stackTracesInfo.code),
		MCode:       int(stackTracesInfo.mcode),
		Message:     CSafeGoString(stackTracesInfo.message),
		StackTraces: goStackTraces,
	}
	C.free(unsafe.Pointer(stackTracesInfo.message))
	return info
}

func getCErrwithStackTrace(e error) *C.CErrorInfo {
	ce := (*C.CErrorInfo)(C.malloc(C.sizeof_CErrorInfo))
	ce.code = 2002
	ce.message = C.CString(e.Error())
	errInfo := api.TurnErrInfo(e)
	if len(errInfo.StackTracesInfo.StackTraces) == 0 {
		ce.size_stackTracesInfo = 0
		return ce
	}
	stsInfo := []api.StackTracesInfo{errInfo.StackTracesInfo}
	stackTracesInfoLen := len(stsInfo)
	cStackTracesInfo := (*C.CStackTracesInfo)(C.malloc(C.size_t(stackTracesInfoLen) * C.sizeof_CStackTracesInfo))
	cStackTracesInfoSlice := unsafe.Slice(cStackTracesInfo, stackTracesInfoLen)
	for idx, val := range stsInfo {
		length := len(val.StackTraces)
		if length == 0 {
			continue
		}
		cStackTraces := (*C.CStackTrace)(C.malloc(C.size_t(length) * C.sizeof_CStackTrace))
		cStackTracesSlice := unsafe.Slice(cStackTraces, length)
		for n, stack := range val.StackTraces {
			cStackTracesSlice[n].className = C.CString(stack.ClassName)
			cStackTracesSlice[n].methodName = C.CString(stack.MethodName)
			cStackTracesSlice[n].fileName = C.CString(stack.FileName)
			cStackTracesSlice[n].lineNumber = C.int64_t(stack.LineNumber)
			cExtensions, cExtensionsLen := cCustomExtensions(stack.ExtensionInfo)
			cStackTracesSlice[n].extensions = cExtensions
			cStackTracesSlice[n].size_extensions = cExtensionsLen
		}
		cStackTracesInfoSlice[idx].stackTraces = &cStackTracesSlice[0]
		cStackTracesInfoSlice[idx].size_stackTraces = C.int(length)
		cStackTracesInfoSlice[idx].message = C.CString(val.Message)
	}
	ce.stackTracesInfo = &cStackTracesInfoSlice[0]
	ce.size_stackTracesInfo = C.int(stackTracesInfoLen)
	return ce
}

// GoCredential transform *C.CCredential to api.Credential
func GoCredential(cCredential C.CCredential) api.Credential {
	return api.Credential{
		AccessKey: CSafeGoString(cCredential.ak),
		SecretKey: CSafeGoBytes(cCredential.sk, cCredential.sizeSk),
		DataKey:   CSafeGoBytes(cCredential.dk, cCredential.sizeDk),
	}
}

// GoFunctionMeta transform *C.CFunctionMeta to api.FunctionMeta
func GoFunctionMeta(funcMeta *C.CFunctionMeta) api.FunctionMeta {
	return api.FunctionMeta{
		AppName:  CSafeGoString(funcMeta.appName),
		FuncName: CSafeGoString(funcMeta.funcName),
		FuncID:   CSafeGoString(funcMeta.functionId),
	}
}

// GoInvokeType transform C.CInvokeType to config.InvokeType
func GoInvokeType(invokeType C.CInvokeType) config.InvokeType {
	return config.InvokeType(invokeType)
}

// GoArgs transform *C.CArg to []api.Arg
func GoArgs(args *C.CArg, argsSize C.int) []api.Arg {
	length := int(argsSize)
	argsSlice := unsafe.Slice(args, length)
	goArgs := make([]api.Arg, length)
	for idx, val := range argsSlice {
		goArgs[idx].Type = api.Value
		goArgs[idx].Data = unsafe.Slice((*byte)(unsafe.Pointer(val.data)), int(val.size))
	}
	return goArgs
}

// GoDataObject transform *C.CDataObject to []config.DataObject
func GoDataObject(returnObjs *C.CDataObject, returnObjsSize C.int) []config.DataObject {
	length := int(returnObjsSize)
	goSlice := make([]config.DataObject, length, length)
	if length == 0 {
		return goSlice
	}
	returnObjsSlice := unsafe.Slice(returnObjs, length)
	for idx, val := range returnObjsSlice {
		goSlice[idx].ID = CSafeGoString(val.id)
		goSlice[idx].Buffer = config.DataBuffer{
			Ptr:  val.buffer.buffer,
			Size: int(val.buffer.size_buffer),
		}
		goSlice[idx].CSharedPtr = val.selfSharedPtr
	}
	return goSlice
}

// GoLoadFunctions is exported as a C function for calling from C/C++ code.
// The purpose is to execute the load callback function of go.
//
//export GoLoadFunctions
func GoLoadFunctions(codePaths **C.char, codePathsSize C.int) *C.CErrorInfo {
	if int(codePathsSize) == 0 {
		return errtoCerr(fmt.Errorf("codePaths empty"))
	}

	err := cfg.Hooks.LoadFunctionCb(GoStringsWithoutFree(codePathsSize, codePaths))
	cErr := errtoCerr(err)
	if err != nil {
		fmt.Println("failed to load function ", err.Error())
	} else {
		fmt.Println("succeed to load function ")
	}
	return cErr
}

// GoFunctionExecution is exported as a C function for calling from C/C++ code.
// The purpose is to execute the function execution callback function of go.
//
//export GoFunctionExecution
func GoFunctionExecution(
	funcMeta *C.CFunctionMeta, invokeType C.CInvokeType, args *C.CArg, argsSize C.int, returnObjs *C.CDataObject,
	returnObjsSize C.int,
) *C.CErrorInfo {
	goFuncMeta := GoFunctionMeta(funcMeta)
	goInvokeType := GoInvokeType(invokeType)
	goArgs := GoArgs(args, argsSize)
	goRetObjs := GoDataObject(returnObjs, returnObjsSize)
	err := cfg.Hooks.FunctionExecutionCb(goFuncMeta, goInvokeType, goArgs, goRetObjs)
	return errtoCerr(err)
}

// GoCheckpoint is exported as a C function for calling from C/C++ code.
// The purpose is to execute the checkpoint callback function of go.
//
//export GoCheckpoint
func GoCheckpoint(checkpointID *C.char, buffer *C.CBuffer) *C.CErrorInfo {
	goCkptId := CSafeGoString(checkpointID)
	data, err := cfg.Hooks.CheckpointCb(goCkptId)
	buf, length := ByteSliceToCBinaryData(data)
	buffer.buffer = buf
	buffer.size_buffer = C.int64_t(length)
	return errtoCerr(err)
}

// GoRecover is exported as a C function for calling from C/C++ code.
// The purpose is to execute the recover callback function of go.
//
//export GoRecover
func GoRecover(buffer *C.CBuffer) *C.CErrorInfo {
	data := C.GoBytes(buffer.buffer, C.int(buffer.size_buffer))
	err := cfg.Hooks.RecoverCb(data)
	return errtoCerr(err)
}

// GoShutdown is exported as a C function for calling from C/C++ code.
// The purpose is to execute the shutdown callback function of go.
//
//export GoShutdown
func GoShutdown(gracePeriodSec C.uint64_t) *C.CErrorInfo {
	err := cfg.Hooks.ShutdownCb(uint64(gracePeriodSec))
	return errtoCerr(err)
}

// GoSignal is exported as a C function for calling from C/C++ code.
// The purpose is to execute the signal callback function of go.
//
//export GoSignal
func GoSignal(sigNo C.int, payload *C.CBuffer) *C.CErrorInfo {
	payloadSlice := C.GoBytes(payload.buffer, C.int(payload.size_buffer))
	err := cfg.Hooks.SignalCb(int(sigNo), payloadSlice)
	return errtoCerr(err)
}

// GoHealthCheck is exported as a C function for calling from C/C++ code.
// The purpose is to execute the health check callback function of go.
//
//export GoHealthCheck
func GoHealthCheck() C.CHealthCheckCode {
	if cfg.Hooks.HealthCheckCb != nil {
		status, _ := cfg.Hooks.HealthCheckCb()
		if status == api.Healthy {
			return C.HEALTHY
		} else if status == api.HealthCheckFailed {
			return C.HEALTH_CHECK_FAILED
		} else if status == api.SubHealth {
			return C.SUB_HEALTH
		}
	}
	return C.HEALTHY
}

// GoHasHealthCheck is to used for check if health check handler is valid
//
//export GoHasHealthCheck
func GoHasHealthCheck() C.char {
	if cfg.Hooks.HealthCheckCb != nil {
		return C.char(1)
	}
	return C.char(0)
}

// GoFunctionExecutionPoolSubmit is used to submit task to golang routines pool from libruntime
//
//export GoFunctionExecutionPoolSubmit
func GoFunctionExecutionPoolSubmit(f unsafe.Pointer) {
	cfg.FunctionExectionPool.Submit(
		func() {
			C.CFunctionExecution(f)
		},
	)
}

// GoEles -
func GoEles(pEles *C.CElement, num C.uint64_t) []api.Element {
	length := uint64(num)
	elesSlice := unsafe.Slice(pEles, length)
	defer C.free(unsafe.Pointer(pEles))
	goEles := make([]api.Element, length)
	for idx, val := range elesSlice {
		// The content in the ptr will be released after ack.
		goEles[idx].Ptr = (*uint8)(unsafe.Pointer(val.ptr))
		goEles[idx].Size = uint64(val.size)
		goEles[idx].Id = uint64(val.id)
	}
	return goEles
}

// IsHealth -
func IsHealth() bool {
	res := int(C.CIsHealth())
	return itob(res)
}

// IsDsHealth -
func IsDsHealth() bool {
	res := int(C.CIsDsHealth())
	return itob(res)
}
