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

// Package utils for common functions
package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/uuid"
)

const (
	// OriginDefaultTimeout is 900
	OriginDefaultTimeout = 900
	// maxTimeout is 100 days
	maxTimeout        = 100 * 24 * 3600
	bytesToMb         = 1024 * 1024
	uint64ArrayLength = 8
	uint32Len         = 4
	// DirMode dir mode
	DirMode = 0700
	// FileMode file mode
	FileMode = 0600
	readSize = 32 * 1024
	// ObsMaxRetry obs max retry times 0
	ObsMaxRetry = 0
	// ObsDefaultTimeout 30 seconds
	ObsDefaultTimeout = 30
	// ObsDefaultConnectTimeout 10 seconds
	ObsDefaultConnectTimeout = 5
	// LayerListSep define the LayerList separation character
	LayerListSep      = "-#-"
	instanceIDLength  = 2
	dnsPairLength     = 2
	hostFilePath      = "/etc/hosts"
	defaultMessageLen = 256
)

const (
	minimumMemoryUnit      = 128
	minimumCPUUnit         = 100
	minimumReservedCPUUnit = 200
)

const (
	tenantValueIndex        = 6
	funcNameValueIndex      = 8
	versionValueIndex       = 10
	instanceIDValueIndex    = 13
	moduleSchedulerKeyLen   = 7
	functionSchedulerKeyLen = 7
	snInstanceKeyLen        = 14
	functionNameIndex       = 6
	leaseIndex              = 7
	defaultVersion          = "latest"
	defaultTenant           = "0"
	defaultFunctionName     = "faas-scheduler"
)

type hostFileInfo struct {
	Sha256  string
	Content []byte
	Mutex   sync.Mutex
}

// HostFile /etc/hosts file info
var HostFile hostFileInfo

// SetClusterNameEnv -
func SetClusterNameEnv(clusterName string) error {
	if err := os.Setenv(constant.ClusterNameEnvKey, clusterName); err != nil {
		return fmt.Errorf("failed to set env of %s, err: %s", constant.ClusterNameEnvKey, err.Error())
	}
	return nil
}

// CalculateCPUByMemory CPU and memory calculation methods presented by fg: cpu=memory/128*100+200
func CalculateCPUByMemory(memory int) int {
	return memory/minimumMemoryUnit*minimumCPUUnit + minimumReservedCPUUnit
}

var azEnv = parseAzEnv()

func parseAzEnv() string {
	az := os.Getenv(constant.ZoneKey)
	if az == "" {
		az = constant.DefaultAZ
	}
	if len(az) > constant.ZoneNameLen {
		az = az[0 : constant.ZoneNameLen-1]
	}
	return az
}

// AzEnv set defaultaz env
func AzEnv() string {
	return azEnv
}

// GenerateInstanceID -
func GenerateInstanceID(podName string) string {
	return AzEnv() + "-#-" + podName
}

// GetPodNameByInstanceID -
func GetPodNameByInstanceID(instanceID string) string {
	elements := strings.Split(instanceID, LayerListSep)
	if len(elements) < instanceIDLength {
		return ""
	}
	return elements[1]
}

// Domain2IP convert domain to ip
func Domain2IP(endpoint string) (string, error) {
	var host, port string
	var err error
	host = endpoint
	if strings.Contains(endpoint, ":") {
		host, port, err = net.SplitHostPort(endpoint)
		if err != nil {
			return "", err
		}
	}
	if net.ParseIP(host) != nil {
		return endpoint, nil
	}
	ips, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	if port == "" {
		return ips[0], nil
	}
	return net.JoinHostPort(ips[0], port), nil
}

// DeepCopy will generate a new copy of original collection type
// currently this function is not recursive so elements will not be deep copied
func DeepCopy(origin interface{}) interface{} {
	oriTyp := reflect.TypeOf(origin)
	oriVal := reflect.ValueOf(origin)
	switch oriTyp.Kind() {
	case reflect.Slice:
		elemType := oriTyp.Elem()
		length := oriVal.Len()
		capacity := oriVal.Cap()
		newObj := reflect.MakeSlice(reflect.SliceOf(elemType), length, capacity)
		reflect.Copy(newObj, oriVal)
		return newObj.Interface()
	case reflect.Map:
		newObj := reflect.MakeMapWithSize(oriTyp, len(oriVal.MapKeys()))
		for _, key := range oriVal.MapKeys() {
			value := oriVal.MapIndex(key)
			newObj.SetMapIndex(key, value)
		}
		return newObj.Interface()
	default:
		return nil
	}
}

// ValidateTimeout check timeout
func ValidateTimeout(timeout *int64, defaultTimeout int64) {
	if *timeout <= 0 {
		*timeout = defaultTimeout
		return
	}
	if *timeout > maxTimeout {
		*timeout = maxTimeout
	}
}

// ClearStringMemory -
func ClearStringMemory(s string) {
	if len(s) == 0 {
		return
	}
	bs := *(*[]byte)(unsafe.Pointer(&s))
	ClearByteMemory(bs)
}

// ClearByteMemory -
func ClearByteMemory(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

// Float64ToByte -
func Float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

// ByteToFloat64 -
func ByteToFloat64(bytes []byte) float64 {
	// bounds check to guarantee safety of function Uint64
	if len(bytes) != uint64ArrayLength {
		return 0
	}
	bits := binary.LittleEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

// ExistPath whether path exists
func ExistPath(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// IsInputParameterValid check if input parameter is valid
func IsInputParameterValid(cmdName string) bool {
	if strings.Contains(cmdName, "&") ||
		strings.Contains(cmdName, "|") ||
		strings.Contains(cmdName, ";") ||
		strings.Contains(cmdName, "$") ||
		strings.Contains(cmdName, "'") ||
		strings.Contains(cmdName, "`") ||
		strings.Contains(cmdName, "(") ||
		strings.Contains(cmdName, ")") ||
		strings.Contains(cmdName, "\"") {
		return false
	}
	return true
}

// UniqueID get unique ID
func UniqueID() string {
	return uuid.New().String()
}

// ShortUUID return short uuid encode by base64
func ShortUUID() string {
	id := uuid.New()
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(id)))
	base64.StdEncoding.Encode(buf, id[:])
	for i := range buf {
		if buf[i] == '=' || buf[i] == '+' || buf[i] == '/' {
			buf[i] = '-'
		}
	}
	return strings.ToLower(strings.Trim(string(buf), "-"))
}

// WriteFileToPath write file to path
func WriteFileToPath(writePath string, buffer []byte) error {
	baseDir := path.Dir(writePath)
	err := os.MkdirAll(baseDir, DirMode)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(writePath, buffer, FileMode); err != nil {
		return err
	}
	return nil
}

// IsConnRefusedErr  -
func IsConnRefusedErr(err error) bool {
	netErr, ok := err.(net.Error)
	if !ok {
		return false
	}
	opErr, ok := netErr.(*net.OpError)
	if !ok {
		return false
	}
	syscallErr, ok := opErr.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	if errno, ok := syscallErr.Err.(syscall.Errno); ok {
		if errno == syscall.ECONNREFUSED {
			return true
		}
	}
	return false
}

// ContainsConnRefusedErr -
func ContainsConnRefusedErr(err error) bool {
	const connRefusedStr = "connection refused"
	return strings.Contains(err.Error(), connRefusedStr)
}

// DefaultStringEnv return environment variable named by key and return val when not exist
func DefaultStringEnv(key string, val string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return val
}

// ReplaceByDNS update /etc/hosts
func ReplaceByDNS(filePath string, domainNames map[string]string) error {
	lines, err := ReadLines(filePath)
	if err != nil {
		return err
	}
	checkedDNSNames := make(map[string]bool, len(domainNames))
	var hasChange bool
	for i := range lines {
		arr := strings.Fields(lines[i])
		if len(arr) != dnsPairLength {
			continue
		}
		for name, ipAddress := range domainNames {
			if arr[0] == name || arr[1] == name {
				originLine := lines[i]
				lines[i] = ipAddress + " " + name
				checkedDNSNames[name] = true
				if lines[i] != originLine {
					hasChange = true
				}
				break
			}
		}
	}
	for name, ipAddress := range domainNames {
		// domain name is not in hosts file will append to hosts file
		if !checkedDNSNames[name] {
			lines = append(lines, ipAddress+" "+name)
			hasChange = true
		}
	}
	if !hasChange {
		return nil
	}
	HostFile.Mutex.Lock()
	defer HostFile.Mutex.Unlock()
	if err := WriteLines(filePath, lines); err != nil {
		return err
	}
	if err := HostFile.SaveHostFileInfo(); err != nil {
		return err
	}
	return nil
}

func (hostFileInfo) SaveHostFileInfo() error {
	_, sha, err := GetFileHashInfo(hostFilePath)
	if err != nil {
		return err
	}
	HostFile.Sha256 = sha
	content, err := ioutil.ReadFile(hostFilePath)
	if err != nil {
		return err
	}
	HostFile.Content = content
	return nil
}

// ReadLines read the lines of the given file.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// WriteLines  writes the lines to the given file.
func WriteLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// GenStateIDByKey returns stateID by serviceID, functionName and key
func GenStateIDByKey(tenantID, serviceID, funcName, key string) string {
	// if stateKey is empty, stateID is generated by default.
	if len(key) == 0 {
		return uuid.New().String()
	}
	preAllocationSlice := make([]byte, 0, len(tenantID)+len(serviceID)+len(funcName)+len(key))
	preAllocationSlice = append(preAllocationSlice, tenantID...)
	preAllocationSlice = append(preAllocationSlice, serviceID...)
	preAllocationSlice = append(preAllocationSlice, funcName...)
	preAllocationSlice = append(preAllocationSlice, key...)
	stateID := uuid.NewSHA1(uuid.NameSpaceURL, preAllocationSlice)
	return stateID.String()
}

// GetFileHashInfo get file hash info
func GetFileHashInfo(path string) (int64, string, error) {
	var fileSize int64
	realPath, err := filepath.Abs(path)
	if err != nil {
		return 0, "", err
	}
	file, err := os.Open(realPath)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return 0, "", err
	}
	fileSize = stat.Size()
	fileHash := sha256.New()
	if _, err := io.Copy(fileHash, file); err != nil {
		return 0, "", err
	}
	hashValue := hex.EncodeToString(fileHash.Sum(nil))
	return fileSize, hashValue, nil
}

// IsNetworkError judge whether it is a network error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(net.Error)
	if !ok {
		return false
	}
	return true
}

// IsUserError -
func IsUserError(err error) bool {
	newErr, ok := err.(snerror.SNError)
	if !ok {
		return false
	}
	return snerror.IsUserError(newErr)
}

// FnvHashInt a hash function
func FnvHashInt(s string) int {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}

	// for 2 <= base <= 36. The result uses the lower-case letters 'a' to 'z'
	return int(h.Sum32())
}

// FileMD5 calculate the md5 of file
func FileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// IsCAEFunc judge whether it is a CAE function
func IsCAEFunc(businessType string) bool {
	return businessType == constant.BusinessTypeCAE
}

var directFunctions = map[string]struct{}{
	"javax": {},
}

// IsDirectFunc check whether it if a direct function (runtime connect to bus directly)
func IsDirectFunc(language string) bool {
	_, ok := directFunctions[language]
	return ok
}

// IsStringInArray -
func IsStringInArray(str string, arr []string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// GetFunctionInstanceInfoFromEtcdKey parses the instance info from the etcd path
// e.g. /sn/instance/business/yrk/tenant/0/function/xxx/version/lastest/defaultaz/
// job-9e54951c-task-77156757-fb16-4b4a-ad61-6646c7d1c57c-d4ad6c74-0/3f079541-15fc-4009-8c41-50b2b2936772
func GetFunctionInstanceInfoFromEtcdKey(path string) (*types.InstanceInfo, error) {
	elements := strings.Split(path, "/")
	if len(elements) != snInstanceKeyLen {
		return nil, fmt.Errorf("unexpected etcd path format: %s", path)
	}
	return &types.InstanceInfo{
		TenantID:     elements[tenantValueIndex],
		FunctionName: elements[funcNameValueIndex],
		Version:      elements[versionValueIndex],
		InstanceName: elements[instanceIDValueIndex],
		InstanceID:   elements[instanceIDValueIndex],
	}, nil
}

// GetSchedulerInfoFromEtcdKey /sn/faas-scheduler/instances/cluster001/7.xx.xx.25/faas-scheduler-xxxx-8xdjf
func GetSchedulerInfoFromEtcdKey(path string) (*types.InstanceInfo, error) {
	elements := strings.Split(path, "/")
	if len(elements) != moduleSchedulerKeyLen {
		return nil, fmt.Errorf("unexpected etcd path format: %s", path)
	}
	return &types.InstanceInfo{
		TenantID:     defaultTenant,
		FunctionName: defaultFunctionName,
		Version:      defaultVersion,
		InstanceName: elements[functionNameIndex],
	}, nil
}

// GetInstanceNameFromSchedulerLeaseEtcdKey /sn/faas-scheduler/instances/cluster001/7.xx.xx.25/faas-scheduler-xxxx-8xdjf/{leaseId}
func GetInstanceNameFromSchedulerLeaseEtcdKey(path string) (string, bool) {
	elements := strings.Split(path, "/")
	if len(elements) != moduleSchedulerKeyLen+1 {
		return "", false
	}

	return elements[functionNameIndex], true
}

// ParseLeaseFromSchedulerLeaseEtcdKey /sn/faas-scheduler/instances/cluster001/7.xx.xx.25/faas-scheduler-xxxx-8xdjf/{leaseId}
func ParseLeaseFromSchedulerLeaseEtcdKey(path string) string {
	elements := strings.Split(path, "/")
	if len(elements) != moduleSchedulerKeyLen+1 {
		return ""
	}

	return elements[leaseIndex]
}

// GetFunctionSchedulerInfoFromEtcdKey /sn/faas-scheduler/instances/cluster001/7.xx.xx.25/faas-scheduler-xxxx-8xdjf
func GetFunctionSchedulerInfoFromEtcdKey(path string) (*types.InstanceInfo, error) {
	elements := strings.Split(path, "/")
	if len(elements) != functionSchedulerKeyLen {
		return nil, fmt.Errorf("unexpected etcd path format: %s", path)
	}
	return &types.InstanceInfo{
		TenantID:     defaultTenant,
		FunctionName: defaultFunctionName,
		Version:      defaultVersion,
		InstanceName: elements[functionNameIndex],
	}, nil
}

// CheckFaaSSchedulerInstanceFault -
func CheckFaaSSchedulerInstanceFault(status types.InstanceStatus) bool {
	faultInstanceStatusMap := map[constant.InstanceStatus]struct{}{
		constant.KernelInstanceStatusFatal:          {},
		constant.KernelInstanceStatusScheduleFailed: {},
		constant.KernelInstanceStatusEvicting:       {},
		constant.KernelInstanceStatusEvicted:        {},
		constant.KernelInstanceStatusExiting:        {},
		constant.KernelInstanceStatusExited:         {},
	}

	_, ok := faultInstanceStatusMap[constant.InstanceStatus(status.Code)]
	return ok
}

// IsNil checks if an object (could be an interface) is nil
func IsNil(i interface{}) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}

// CalcFileMD5 calculates file MD5
func CalcFileMD5(filepath string) string {
	file, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer file.Close()
	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// ReceiveWithinTimeout first element is the chan, second element is the timeout
func ReceiveWithinTimeout[T any](ch <-chan T, timeout time.Duration) (T, bool) {
	var val T
	select {
	case val, ok := <-ch:
		return val, ok
	case <-time.After(timeout):
		return val, false
	}
}

func MessageTruncation(message string) string {
	if len(message) > defaultMessageLen {
		return message[:defaultMessageLen]
	}
	return message
}

// SafeCloseChannel will close channel in a safe way
func SafeCloseChannel(stopCh chan struct{}) {
	if stopCh == nil {
		return
	}
	select {
	case _, ok := <-stopCh:
		if ok {
			close(stopCh)
		}
	default:
		close(stopCh)
	}
}
