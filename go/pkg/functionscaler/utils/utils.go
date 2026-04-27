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

// Package utils -
package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool"
	"yuanrong.org/kernel/pkg/common/uuid"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	ipPortLens           = 2
	defaultConcurrentNum = 100
	// DefaultMapSize is default size of map
	DefaultMapSize = 16
	// DefaultSliceSize is default size of slice
	DefaultSliceSize     = 16
	defaultCPUValue      = 0
	defaultMemValue      = 0
	resourceCPUName      = "CPU"
	resourceMemoryName   = "Memory"
	validInnerFuncKeyLen = 3
	// ValidFuncKeyLen is the valid length of funcKey
	ValidFuncKeyLen = 3
	// FuncKeyDelimiter is the delimiter for parsing inner funcKey from funcKey
	FuncKeyDelimiter = "/"
	// InnerFuncKeyDelimiter is the delimiter for parsing funcName from inner funcKey
	InnerFuncKeyDelimiter  = "-"
	funcNameIndexInFuncKey = 2
	defaultCreateTimeout   = time.Duration(30) * time.Second
	maxSessionLength       = 63
	namespaceLabelKey      = "POD_NAMESPACE"
	deploymentNameLabelKey = "POD_DEPLOYMENT_NAME"
	podNameLabelKey        = "POD_NAME"
	// CrKeySep is a CrName separator of CrKey, crKey example: default:12268ce-478c-8ae7-3bd7d6c6033-hello-agent-latest
	CrKeySep            = ":"
	crKeyElementsCounts = 2
	namespaceIndex      = 0
	crNameIndex         = 1
)

var (
	labelRegexp = regexp.MustCompile(`-invoke-label-(.*?)-ephemeral-storage-`)
)

// RecoverSessionCallback -
type RecoverSessionCallback func(sessInfo *types.SessionInfo, instance *types.Instance)

// GetNpuInstanceType -
func GetNpuInstanceType(delegateContainer string) (bool, string) {
	config := &types.DelegateContainerConfig{}
	err := json.Unmarshal([]byte(delegateContainer), config)
	if err != nil {
		return false, ""
	}

	for _, v := range config.Env {
		if v.Name == types.SystemNodeInstanceType {
			return true, v.Value
		}
	}
	return false, ""
}

// GetNpuTypeAndInstanceTypeFromStr 函数用于从字符串中获取 NPU 类型和实例类型
func GetNpuTypeAndInstanceTypeFromStr(customResource string, customResourcesSpec string) (string, string) {
	customResourceMap := make(map[string]int64)
	customResourceSpecMap := make(map[string]interface{})

	err1 := json.Unmarshal([]byte(customResource), &customResourceMap)
	err2 := json.Unmarshal([]byte(customResourcesSpec), &customResourceSpecMap)
	if err1 != nil {
		return "", ""
	}

	if err2 != nil {
		// pass
	}
	return GetNpuTypeAndInstanceType(customResourceMap, customResourceSpecMap)
}

// GetNpuTypeAndInstanceType 函数用于获取 NPU 类型和实例类型
func GetNpuTypeAndInstanceType(customRes map[string]int64, customResSpec map[string]interface{}) (string, string) {
	v, ok := customRes[types.AscendResourceD910B]
	if !ok || v <= 0 {
		return "", ""
	}
	d910bType, ok := customResSpec[types.AscendResourceD910BInstanceType]
	if !ok {
		return types.AscendResourceD910B, "376T"
	}
	d910bTypeStr, ok := d910bType.(string)
	if !ok {
		return types.AscendResourceD910B, "376T"
	}
	return types.AscendResourceD910B, d910bTypeStr
}

// ConvertInstanceResource will convert instance resource
func ConvertInstanceResource(res commonTypes.Resources) *resspeckey.ResourceSpecification {
	resSpec := &resspeckey.ResourceSpecification{
		CustomResources: make(map[string]int64, constant.DefaultMapSize),
	}
	for k, v := range res.Resources {
		if k == constant.ResourceCPUName {
			resSpec.CPU = int64(v.Scalar.Value)
			continue
		}
		if k == constant.ResourceMemoryName {
			resSpec.Memory = int64(v.Scalar.Value)
			continue
		}
		if k == constant.ResourceEphemeralStorage {
			resSpec.EphemeralStorage = int(v.Scalar.Value)
			continue
		}
		resSpec.CustomResources[k] = int64(v.Scalar.Value)
	}
	return resSpec
}

// AppendInstanceTypeToInstanceResource -
func AppendInstanceTypeToInstanceResource(resSpec *resspeckey.ResourceSpecification, npuInstanceType string) {
	if resSpec == nil {
		return
	}
	if resSpec.CustomResourcesSpec == nil {
		resSpec.CustomResourcesSpec = make(map[string]interface{})
	}
	resSpec.CustomResourcesSpec["instanceType"] = npuInstanceType
}

// IsFaaSManager checks if a innerFuncKey is FaaSManager, innerFuncKey example: 0-system-faasmanager
func IsFaaSManager(innerFuncKey string) bool {
	items := strings.Split(innerFuncKey, InnerFuncKeyDelimiter)
	if len(items) != validInnerFuncKeyLen {
		return false
	}
	return items[funcNameIndexInFuncKey] == types.FaasManagerFuncName
}

var systemFuncNames = map[string]struct{}{
	types.FaasManagerFuncName:    {},
	types.FaasSchedulerFuncName:  {},
	types.FaasFrontendFuncName:   {},
	types.FaasControllerFuncName: {},
}

// IsFaaSInstance -
func IsFaaSInstance(funcKey string) bool {
	items := strings.Split(funcKey, FuncKeyDelimiter)
	if len(items) != ValidFuncKeyLen {
		return false
	}
	innerItems := strings.Split(items[1], InnerFuncKeyDelimiter)
	if len(innerItems) != validInnerFuncKeyLen {
		return false
	}

	if items[0] == "0" {
		_, ok := systemFuncNames[innerItems[funcNameIndexInFuncKey]]
		if ok {
			return false
		}
	}
	return true
}

// GetConcurrentNum gets a valid concurrentNum
func GetConcurrentNum(concurrentNum int) int {
	if concurrentNum == 0 {
		return 1
	}
	return concurrentNum
}

// GenerateTraceID -
func GenerateTraceID() string {
	return uuid.New().String()
}

// IsUnrecoverableError checks if error should not be retried
func IsUnrecoverableError(err error) bool {
	if snErr, ok := err.(snerror.SNError); ok {
		if snerror.IsUserError(snErr) {
			return true
		}
		if snErr.Code() == statuscode.UserFuncEntryNotFoundErrCode || snErr.Code() == statuscode.KernelEtcdWriteFailedCode ||
			snErr.Code() == statuscode.WiseCloudNuwaColdStartErrCode || snErr.Code() == statuscode.StsConfigErrCode {
			return true
		}
	}
	return false
}

// IsNoNeedToRePullError checks if error pod should not be repull
func IsNoNeedToRePullError(err error) bool {
	if snErr, ok := err.(snerror.SNError); ok {
		return snErr.Code() == statuscode.StsConfigErrCode
	}
	return false
}

// GenerateStsSecretName -
func GenerateStsSecretName(etcdFuncKey string) string {
	// keep same as CaaS, prevent secret resource leaks
	return strings.ToLower(fmt.Sprintf("%s-sts", urnutils.CrNameByKey(etcdFuncKey)))
}

// AddNodeSelector -
func AddNodeSelector(nodeSelectorMap map[string]string, schedulingOptions *types.SchedulingOptions,
	resSpec *resspeckey.ResourceSpecification) {
	if resSpec == nil {
		return
	}
	if resSpec.CustomResources == nil || len(resSpec.CustomResources) == 0 {
		if nodeSelectorMap != nil && len(nodeSelectorMap) != 0 {
			nodeSelectorVal, err := json.Marshal(nodeSelectorMap)
			if err != nil {
				log.GetLogger().Warnf("failed to marshal "+
					"nodeSelectorMap %v, err %s", nodeSelectorMap, err.Error())
				return
			}
			schedulingOptions.Extension[utils.NodeSelectorKey] = string(nodeSelectorVal)
		}
	}
}

// AddAffinityCPU -
func AddAffinityCPU(crName string, schedulingOptions *types.SchedulingOptions,
	resSpec *resspeckey.ResourceSpecification, affinityType api.AffinityType) {
	if resSpec == nil {
		return
	}
	if resSpec.CustomResources == nil || len(resSpec.CustomResources) == 0 {
		schedulingOptions.Affinity = commonUtils.CreatePodAffinity(crName, "", affinityType)
	}
}

// IsResSpecEmpty -
func IsResSpecEmpty(resSpec *resspeckey.ResourceSpecification) bool {
	if resSpec == nil {
		return true
	}
	return resSpec.CPU == 0 && resSpec.Memory == 0 && len(resSpec.CustomResources) == 0
}

// GetCreateTimeout -
func GetCreateTimeout(funcSpec *types.FunctionSpecification) time.Duration {
	createTimeout := time.Duration(funcSpec.ExtendedMetaData.Initializer.Timeout) * time.Second
	if funcSpec.FuncMetaData.Runtime == types.CustomContainerRuntimeType {
		createTimeout += constant.CustomImageExtraTimeout * time.Second
	} else {
		createTimeout += (constant.KernelScheduleTimeout + constant.CommonExtraTimeout) * time.Second
	}
	if createTimeout < defaultCreateTimeout {
		createTimeout = defaultCreateTimeout
	}
	return createTimeout
}

// GetRequestTimeout -
func GetRequestTimeout(funcSpec *types.FunctionSpecification) time.Duration {
	if funcSpec.InstanceMetaData.ScalePolicy == types.InstanceScalePolicyStaticFunction {
		return time.Duration(funcSpec.FuncMetaData.Timeout) * time.Second
	} else {
		return GetCreateTimeout(funcSpec)
	}
}

// GenRandomString will generate random string with given length
func GenRandomString(n int) string {
	randBytes := make([]byte, n/2)
	if _, err := rand.Read(randBytes); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", randBytes)
}

// GenFuncKeyWithRes will generate funcKeyWithRes
func GenFuncKeyWithRes(funcKey, resKey string) string {
	return fmt.Sprintf("%s-%s", funcKey, resKey)
}

// GetLeaseInterval returns the interval of lease
func GetLeaseInterval() time.Duration {
	leaseInterval := time.Duration(config.GlobalConfig.LeaseSpan) * time.Millisecond
	if leaseInterval < types.MinLeaseInterval {
		leaseInterval = types.MinLeaseInterval
	}
	return leaseInterval
}

func parseIPAndPort(address string) (string, string) {
	var (
		IP, Port string
	)
	Address := strings.Split(address, ":")
	if len(Address) == ipPortLens {
		IP = Address[0]
		Port = Address[1]
	}
	return IP, Port
}

// BuildInstanceFromInsSpec builds instance from instanceSpecification
func BuildInstanceFromInsSpec(insSpec *commonTypes.InstanceSpecification,
	funcSpec *types.FunctionSpecification) *types.Instance {
	instanceIP, instancePort := parseIPAndPort(insSpec.RuntimeAddress)
	// FunctionProxyID is not "IP:Port" like string, need to find another way to get nodeIP and nodePort
	nodeIP, nodePort := parseIPAndPort(insSpec.FunctionProxyID)
	instanceName := insSpec.CreateOptions[types.InstanceNameNote]
	var (
		instanceType  types.InstanceType
		isPermanent   bool
		schedulerID   string
		concurrentNum int
		err           error
	)
	instanceNote, ok := insSpec.CreateOptions[types.InstanceTypeNote]
	if ok {
		instanceType = types.InstanceType(instanceNote)
	} else {
		instanceType = types.InstanceTypeUnknown
	}
	resSpecKey, err := resspeckey.GetResKeyFromStr(insSpec.CreateOptions[types.ResourceSpecNote])
	if err != nil {
		log.GetLogger().Errorf("failed to GetResKeyFromStr from %s for instance %s",
			insSpec.CreateOptions[types.ResourceSpecNote], insSpec.InstanceID)
	}
	schedulerNote := insSpec.CreateOptions[types.SchedulerIDNote]
	if strings.HasSuffix(schedulerNote, types.PermanentInstance) {
		isPermanent = true
		schedulerID = strings.TrimSuffix(schedulerNote, types.PermanentInstance)
	} else if strings.HasSuffix(schedulerNote, types.TemporaryInstance) {
		schedulerID = strings.TrimSuffix(schedulerNote, types.TemporaryInstance)
	} else {
		schedulerID = schedulerNote
	}
	concurrentNote := insSpec.CreateOptions[types.ConcurrentNumKey]
	if len(concurrentNote) != 0 {
		concurrentNum, err = strconv.Atoi(concurrentNote)
		if err != nil {
			log.GetLogger().Errorf("failed to parse concurrentNum from %s for instance %s", concurrentNote,
				insSpec.InstanceID)
			concurrentNum = defaultConcurrentNum
		}
	}
	metricsLabels := make([]string, 0)
	if funcSpec != nil {
		metricsLabels = wisecloudtool.GetMetricLabels(&funcSpec.FuncMetaData, resSpecKey.InvokeLabel,
			insSpec.Extensions.PodNamespace, insSpec.Extensions.PodDeploymentName, insSpec.Extensions.PodName)
	}
	return &types.Instance{
		InstanceStatus: insSpec.InstanceStatus,
		InstanceType:   instanceType,
		ResKey:         resSpecKey,
		InstanceID:     insSpec.InstanceID,
		InstanceName:   instanceName,
		ParentID:       insSpec.ParentID,
		FuncSig:        insSpec.CreateOptions[types.FunctionSign],
		FuncKey:        insSpec.Function,
		// this instance maybe an old one with previous concurrentNum, need to optimize this logic in future
		ConcurrentNum:     concurrentNum,
		CreateSchedulerID: schedulerID,
		Permanent:         isPermanent,
		NodeIP:            nodeIP,
		NodePort:          nodePort,
		InstanceIP:        instanceIP,
		InstancePort:      instancePort,
		MetricLabelValues: metricsLabels,
		PodID:             insSpec.Extensions.PodNamespace + ":" + insSpec.Extensions.PodName,
		PodDeploymentName: insSpec.Extensions.PodDeploymentName,
	}
}

// CheckInstanceSessionValid checks if InstanceSessionConfig is valid
func CheckInstanceSessionValid(insSess commonTypes.InstanceSessionConfig) bool {
	if len(insSess.SessionID) == 0 || len(insSess.SessionID) > maxSessionLength {
		return false
	}
	if insSess.SessionTTL < 0 {
		return false
	}
	if insSess.Concurrency == 0 || insSess.Concurrency < -1 {
		return false
	}
	return true
}

// GetInvokeLabelFromResKey get invoke label from res key
func GetInvokeLabelFromResKey(s string) string {
	if !strings.Contains(s, "-invoke-label-") {
		return ""
	}
	matches := labelRegexp.FindStringSubmatch(s)
	// 第一个匹配项是整个匹配的字符串，第二个匹配项是我们想要的
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// IntMax returns the maximum of the params
func IntMax(a, b int) int {
	if b > a {
		return b
	}
	return a
}

// ParseFromCrKey Parse namespace and crName from crKey
func ParseFromCrKey(crKey string) (string, string, error) {
	elements := strings.Split(crKey, CrKeySep)
	crLen := len(elements)
	if crLen != crKeyElementsCounts {
		return "", "", fmt.Errorf("failed to parse crName from: %s, invalid length: %d", crKey, crLen)
	}
	return elements[namespaceIndex], elements[crNameIndex], nil
}
