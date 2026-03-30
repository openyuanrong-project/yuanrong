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

package instancepool

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/statuscode"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/signalmanager"
	"yuanrong.org/kernel/pkg/functionscaler/sts"
	"yuanrong.org/kernel/pkg/functionscaler/tenantquota"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// 1 + 2 + 4 + 8 + 16 + 32 + 64 + 128 + 256 + 300 + 300... = 3811秒
const (
	retryDuration = 1 * time.Second // 初始等待时间
	retryFactor   = 2               // 倍数因子（每次翻4倍）
	retryJitter   = 0.1             // 随机抖动系数
	retryTime     = 20
	retryCap      = 300 * time.Second // 最大等待时间上限
)

var (
	globalSdkClient  api.LibruntimeAPI
	coldStartBackoff = wait.Backoff{
		Duration: retryDuration,
		Factor:   retryFactor,
		Jitter:   retryJitter,
		Steps:    retryTime,
		Cap:      retryCap,
	}
	disableSWR = strings.TrimSpace(os.Getenv("DISABLE_SWR_ADDON")) == "true"
)

// SetGlobalSdkClient -
func SetGlobalSdkClient(sdkClient api.LibruntimeAPI) {
	globalSdkClient = sdkClient
}

func killInstanceAndIgnoreNotFoundError(instanceId string) error {
	err := globalSdkClient.Kill(instanceId, killSignalVal, []byte{})
	if err != nil {
		if strings.Contains(err.Error(), "instance not found") {
			return nil
		}
	}
	return err
}

func dealWithVPCError(originErr error) error {
	if errCode := statuscode.VpcCode(originErr.Error()); errCode != 0 {
		return snerror.New(errCode, statuscode.VpcErMsg(errCode))
	}
	return originErr
}

func createInstanceForKernel(request createInstanceRequest) (instance *types.Instance, createErr error) {
	logger := log.GetLogger().With(zap.Any("funcKey", request.funcSpec.FuncKey))
	logger.Infof("start to create instance, traceID: %s", request.traceID)
	createBeginTime := time.Now()
	resSpec := request.resKey.ToResSpec()
	createOpt, args, err := generateOptionAndArgsForCreate(request, resSpec)
	if err != nil {
		createErr = err
		return
	}
	if request.funcSpec.ExtendedMetaData.VpcConfig != nil {
		vpcNatConfigs, vpcErr := createPATService(request.traceID, request.funcSpec, request.faasManagerInfo,
			request.funcSpec.ExtendedMetaData, request.funcSpec.ExtendedMetaData.VpcConfig)
		if vpcErr != nil {
			createErr = dealWithVPCError(vpcErr)
			return
		}
		setCreateOptionForVPC(createOpt, vpcNatConfigs)
	}
	defer func() {
		if createErr != nil && config.GlobalConfig.TenantInsNumLimitEnable {
			tenantID := urnutils.GetTenantFromFuncKey(instance.FuncKey)
			tenantquota.DecreaseTenantInstance(tenantID, instance.InstanceType == types.InstanceTypeReserved)
		}
	}()

	var instanceID string
	schedulingOptions := prepareSchedulingOptions(request.funcSpec, resSpec)
	funcMeta := api.FunctionMeta{
		FuncID: getExecutorFuncKey(request.funcSpec),
		Api:    commonUtils.GetAPIType(request.funcSpec.FuncMetaData.BusinessType),
	}
	invokeOpts := createInvokeOptions(request.funcSpec, schedulingOptions, createOpt, request.poolLabel)
	logger.Debugf("invoke opts cpu is %v, mem is %v\n", invokeOpts.Cpu, invokeOpts.Memory)
	delete(invokeOpts.CustomResources, resourcesCPU)
	delete(invokeOpts.CustomResources, resourcesMemory)
	instanceID, createErr = globalSdkClient.CreateInstance(funcMeta, args, invokeOpts)
	if createErr != nil {
		logger.Errorf("failed to create instance, error info is %#v", createErr)
		createErr = generateSnErrorFromKernelError(createErr)
		return
	}
	instance = buildInstance(instanceID, request)
	logger.Infof("succeed to create instance %s, cost: %s, instance: %+v", instanceID, time.Since(createBeginTime),
		instance)
	return
}

func deleteInstanceForKernel(funcSpec *types.FunctionSpecification, faasManagerInfo faasManagerInfo,
	instance *types.Instance) error {
	log.GetLogger().Debugf("start to delete instance %s for function %s", instance.InstanceID, funcSpec.FuncKey)
	// maybe we should wait for delete response
	var err error
	err = killInstanceAndIgnoreNotFoundError(instance.InstanceID)
	if err != nil {
		log.GetLogger().Errorf("failed to delete instance %s for function %s first, start retry",
			instance.InstanceID, funcSpec.FuncKey)
		go deleteInstanceForKernelRetry(instance.InstanceID)
		return err
	}
	if config.GlobalConfig.TenantInsNumLimitEnable {
		tenantID := urnutils.GetTenantFromFuncKey(instance.FuncKey)
		tenantquota.DecreaseTenantInstance(tenantID, instance.InstanceType == types.InstanceTypeReserved)
	}
	log.GetLogger().Infof("succeed to delete instance %s for function %s", instance.InstanceID, funcSpec.FuncKey)
	return nil
}

func deleteInstanceForKernelRetry(instanceID string) {
	var err error
	backoffErr := wait.ExponentialBackoff(
		coldStartBackoff, func() (bool, error) {
			err = killInstanceAndIgnoreNotFoundError(instanceID)
			if err != nil {
				log.GetLogger().Errorf("failed to delete instance %s, retry", instanceID)
				return false, nil
			}
			return true, nil
		})
	if backoffErr != nil {
		log.GetLogger().Errorf("[MAJOR] failed to delete instance %s, err is %s", backoffErr.Error())
	}
	if err != nil {
		log.GetLogger().Errorf("[MAJOR] failed to delete instance %s, err is %s", err.Error())
	}
}

func deleteInstanceByIDForKernel(instanceID, funcKey string) error {
	log.GetLogger().Debugf("start to delete instance %s for function %s",
		instanceID, funcKey)
	if err := killInstanceAndIgnoreNotFoundError(instanceID); err != nil {
		log.GetLogger().Errorf("failed to delete instance %s, err: %v, retry", instanceID, err)
		go deleteInstanceForKernelRetry(instanceID)
	}
	return nil
}

func getExecutorFuncKey(funcSpec *types.FunctionSpecification) string {
	if funcSpec.FuncMetaData.BusinessType == constant.BusinessTypeServe {
		return serveExecutor
	}
	switch {
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "python3.6"):
		return fmt.Sprintf(executorFormat, "Python3.6")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "python3.7"):
		return fmt.Sprintf(executorFormat, "Python3.7")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "python3.8"):
		return fmt.Sprintf(executorFormat, "Python3.8")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "python3.9"):
		return fmt.Sprintf(executorFormat, "Python3.9")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "python3.10"):
		return fmt.Sprintf(executorFormat, "Python3.10")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "python3.11"):
		return fmt.Sprintf(executorFormat, "Python3.11")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "go"),
		strings.Contains(funcSpec.FuncMetaData.Runtime, "http"),
		strings.Contains(funcSpec.FuncMetaData.Runtime, "custom image"):
		return fmt.Sprintf(executorFormat, "Go1.x")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "java8"):
		return fmt.Sprintf(executorFormat, "Java8")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "java11"):
		return fmt.Sprintf(executorFormat, "Java11")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "java17"):
		return fmt.Sprintf(executorFormat, "Java17")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, "java21"):
		return fmt.Sprintf(executorFormat, "Java21")
	case strings.Contains(funcSpec.FuncMetaData.Runtime, constant.PosixCustomRuntimeType):
		return fmt.Sprintf(executorFormat, "PosixCustom")
	default:
		return funcSpec.FuncKey
	}
}

func generateOptionAndArgsForCreate(request createInstanceRequest, resSpec *resspeckey.ResourceSpecification) (
	map[string]string, []api.Arg, error) {
	createOpt, err := prepareCreateOptions(request, resSpec)
	if err != nil || createOpt == nil {
		return nil, nil, err
	}
	err = errors.New("failed to create instance, failed to generate option and argument")
	args := prepareCreateArguments(request)
	if args == nil {
		return nil, nil, err
	}
	return createOpt, args, nil
}

func prepareSchedulingOptions(funcSpec *types.FunctionSpecification,
	resSpec *resspeckey.ResourceSpecification) *types.SchedulingOptions {
	schedulingOptions := &types.SchedulingOptions{}
	schedulingOptions.Resources = generateResources(resSpec)
	if config.GlobalConfig.DeployMode == constant.DeployModeProcesses {
		schedulingOptions.Extension = commonUtils.CreateCustomExtensions(schedulingOptions.Extension,
			commonUtils.SharedPolicyValue)
	} else {
		schedulingOptions.Extension = commonUtils.CreateCustomExtensions(schedulingOptions.Extension,
			commonUtils.MonopolyPolicyValue)
	}
	if len(funcSpec.InstanceMetaData.PoolID) != 0 {
		schedulingOptions.Extension[constant.AffinityPoolIDKey] = funcSpec.InstanceMetaData.PoolID
	}
	nodeSelectorMap := config.GlobalConfig.NodeSelector
	if funcSpec.NameSpace != "" && funcSpec.NameSpace != constant.DefaultNameSpace {
		nodeSelectorMap = make(map[string]string)
		for k, v := range config.GlobalConfig.NodeSelector {
			nodeSelectorMap[k] = v
		}
		nodeSelectorMap[constant.NodeNameSpaceSelectorKey] = funcSpec.NameSpace
		utils.AddNodeSelector(nodeSelectorMap, schedulingOptions, resSpec)
		log.GetLogger().Infof("succeed to add nodeSelector "+
			"node-namespace=%s for func %s", funcSpec.NameSpace, funcSpec.FuncKey)
	} else {
		utils.AddNodeSelector(config.GlobalConfig.NodeSelector, schedulingOptions, resSpec)
	}
	log.GetLogger().Infof("generate scheduling options %v "+
		"for function %s", schedulingOptions, funcSpec.FuncKey)
	return schedulingOptions
}

func createInvokeOptions(funcSpec *types.FunctionSpecification, schedulingOptions *types.SchedulingOptions,
	createOpt map[string]string, poolLabel string) api.InvokeOptions {
	codeEntrys := []string{funcSpec.ExtendedMetaData.Initializer.Handler, funcSpec.FuncMetaData.Handler}
	if funcSpec.ExtendedMetaData.PreStop.Handler != "" {
		codeEntrys = append(codeEntrys, funcSpec.ExtendedMetaData.PreStop.Handler)
	}
	invokeOpts := api.InvokeOptions{
		Cpu:                int(schedulingOptions.Resources[resourcesCPU]),
		Memory:             int(schedulingOptions.Resources[resourcesMemory]),
		CustomResources:    schedulingOptions.Resources,
		CustomExtensions:   schedulingOptions.Extension,
		CreateOpt:          createOpt,
		Priority:           int(schedulingOptions.Priority),
		ScheduleAffinities: generateScheduleAffinity(schedulingOptions.Affinity, poolLabel),
		Labels:             []string{"faas"},
		CodePaths:          codeEntrys,
		Timeout:            int(utils.GetCreateTimeout(funcSpec).Seconds()),
		ScheduleTimeoutMs:  constant.KernelScheduleTimeout * time.Second.Milliseconds(),
	}
	return invokeOpts
}

func generateScheduleAffinity(scheduleAffinity []api.Affinity, label string) []api.Affinity {
	if label == "" {
		return scheduleAffinity
	}
	labels := strings.Split(label, ",")
	for _, poolLabel := range labels {
		if strings.TrimSpace(poolLabel) == constant.UnUseAntiOtherLabelsKey {
			continue
		}
		affinity := api.Affinity{
			Kind:                     api.AffinityKindResource,
			Affinity:                 api.PreferredAffinity,
			PreferredPriority:        true,
			PreferredAntiOtherLabels: !strings.Contains(label, constant.UnUseAntiOtherLabelsKey),
			LabelOps: []api.LabelOperator{
				{
					Type:        api.LabelOpExists,
					LabelKey:    strings.TrimSpace(poolLabel),
					LabelValues: nil,
				},
			},
		}
		scheduleAffinity = append(scheduleAffinity, affinity)
	}
	return scheduleAffinity
}

func createPATService(traceID string, funcSpec *types.FunctionSpecification, faasManagerInfo faasManagerInfo,
	extMetaData commonTypes.ExtendedMetaData, vpcConfig *commonTypes.VpcConfig) ([]commonTypes.NATConfigure, error) {
	createErr := errors.New("failed to create pat service")
	faasManagerFuncKey := faasManagerInfo.funcKey
	faasManagerInstanceID := faasManagerInfo.instanceID
	if len(faasManagerFuncKey) == 0 || len(faasManagerInstanceID) == 0 {
		log.GetLogger().Errorf("no faas manager instance info")
		return nil, createErr
	}
	if traceID == "" {
		traceID = utils.GenerateTraceID()
	}
	createCh := make(chan error, 1)
	args := prepareCreatePATServiceArguments(traceID, funcSpec.NameSpace, extMetaData, vpcConfig)
	var responseData []byte
	funcMeta := api.FunctionMeta{FuncID: faasManagerFuncKey, Api: api.FaaSApi}
	invokeOpts := api.InvokeOptions{TraceID: traceID}
	objID, invokeErr := globalSdkClient.InvokeByInstanceId(funcMeta, faasManagerInstanceID, args, invokeOpts)
	if invokeErr != nil {
		log.GetLogger().Errorf("failed to send create pat service request for vpcID %s, "+
			"traceID: %s, function: %s, error: %s", vpcConfig.VpcID, traceID, funcSpec.FuncKey, invokeErr.Error())
		return nil, createErr
	}
	globalSdkClient.GetAsync(objID, func(result []byte, err error) {
		responseData = result
		createCh <- err
	})
	timer := time.NewTimer(faasManagerRequestTimeout)
	select {
	case resultErr, ok := <-createCh:
		if !ok {
			log.GetLogger().Errorf("result channel of PATService request is closed, traceID: %s", traceID)
			return nil, createErr
		}
		timer.Stop()
		if resultErr != nil {
			log.GetLogger().Errorf("failed to create pat service with vpcID %s for function %s, "+
				"faas-manager instanceID: %s, traceID: %s, error: %s", vpcConfig.VpcID,
				funcSpec.FuncKey, faasManagerInstanceID, traceID, resultErr.Error())
			return nil, resultErr
		}
	case <-timer.C:
		log.GetLogger().Errorf("time out waiting for PATService creation of function %s, traceID: %s",
			funcSpec.FuncKey, traceID)
		return nil, createErr
	}
	return parseResponse(responseData, traceID, createErr)
}

func parseResponse(responseData []byte, traceID string, createErr error) ([]commonTypes.NATConfigure, error) {
	response := &patSvcCreateResponse{}
	err := json.Unmarshal(responseData, response)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal PATService create response, traceID: %s, error: %s",
			traceID, err.Error())
		return nil, createErr
	}
	if response.Code != constant.InsReqSuccessCode {
		log.GetLogger().Errorf("failed to create pat service with response code: %s, msg: %s, traceID: %s",
			response.Code, response.Message, traceID)
		return nil, createErr
	}
	log.GetLogger().Infof("succeed to create pat with response %s", string(responseData))
	patConfigs := &commonTypes.PatResponseInfo{}
	err = json.Unmarshal([]byte(response.Message), patConfigs)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal network config from response, traceID: %s, error: %s",
			traceID, err.Error())
		return nil, createErr
	}
	return patConfigs.PatPods, nil
}

func generateResources(resSpec *resspeckey.ResourceSpecification) map[string]float64 {
	resourcesMap := make(map[string]float64)
	if resSpec == nil {
		return resourcesMap
	}
	resourcesMap[resourcesCPU] = float64(resSpec.CPU)
	resourcesMap[resourcesMemory] = float64(resSpec.Memory)
	if resSpec.CustomResources != nil {
		for key, value := range resSpec.CustomResources {
			resourcesMap[key] = float64(value)
		}
	}
	return resourcesMap
}

func setCreateOptionForDelegateBootstrap(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.FuncMetaData.Runtime == constant.PosixCustomRuntimeType {
		createOpt[constant.DelegateBootstrapKey] = funcSpec.FuncMetaData.Handler
	}
	return nil
}

func setCreateOptionForInvokeTimeout(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.FuncMetaData.Timeout > 0 {
		createOpt[constant.FaasInvokeTimeout] = strconv.FormatInt(funcSpec.FuncMetaData.Timeout, base)
	}
	return nil
}

// CreateOption contains params for runtime not for user code
func prepareCreateOptions(request createInstanceRequest, resSpec *resspeckey.ResourceSpecification) (map[string]string,
	error) {
	createOpt := make(map[string]string, constant.DefaultMapSize)
	setCreateOptionForFuncSpec(request.funcSpec, createOpt)
	setFunctions := []func(*types.FunctionSpecification, map[string]string) error{
		setCreateOptionForDownloadData,
		setCreateOptionForDelegateMount, setCreateOptionForUserAgencyAndEnv, setCreateOptionForDelegateContainer,
		setCreateOptionForFileBeat, setCreateOptionForHostAliases, setCreateOptionForRASP, setCreateOptionForOtel,
		setCustomPodSeccompProfile,
		setFunctionAgentInitContainer, setCreateOptionForInitContainerEnv, setCreateOptionForLifeCycleDetached,
		setCreateOptionForDelegateBootstrap, setCreateOptionForPostStartExec, setCreateOptionForInvokeTimeout,
	}
	for _, f := range setFunctions {
		if err := f(request.funcSpec, createOpt); err != nil {
			return nil, err
		}
	}
	if err := setCreateOptionForNuwaRuntimeInfo(request.nuwaRuntimeInfo, createOpt); err != nil {
		return nil, err
	}
	if err := setCreateOptionForName(request.instanceName, request.callerPodName, createOpt); err != nil {
		return nil, err
	}
	if err := setCreateOptionForLabel(request.instanceType, request.funcSpec, resSpec, createOpt); err != nil {
		return nil, err
	}
	if err := setCreateOptionForNote(request.instanceType, request.funcSpec, resSpec, createOpt); err != nil {
		return nil, err
	}
	if err := setCreateOptionForAscendNPU(request.funcSpec, resSpec, createOpt); err != nil {
		return nil, err
	}
	setCreateOptionForImagePullSecrets(request.funcSpec, createOpt)
	return createOpt, nil
}

func setCreateOptionForImagePullSecrets(funcSpec *types.FunctionSpecification, createOpt map[string]string) {
	if disableSWR || funcSpec.FuncMetaData.BusinessType != constant.BusinessTypeAgent ||
		createOpt == nil || len(funcSpec.ExtendedMetaData.ImagePullConfig.Secrets) == 0 {
		return
	}
	createOpt[constant.ImagePullSecrets] = strings.Join(funcSpec.ExtendedMetaData.ImagePullConfig.Secrets, ",")
}

func setCreateOptionForPostStartExec(funcSpec *types.FunctionSpecification,
	createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.FuncMetaData.BusinessType == constant.BusinessTypeServe &&
		len(funcSpec.ExtendedMetaData.ServeDeploySchema.Applications) != 0 &&
		len(funcSpec.ExtendedMetaData.ServeDeploySchema.Applications[constant.ApplicationIndex].
			RuntimeEnv.Pip) != 0 {
		installCommand := fmt.Sprintf("%s %s && %s", constant.PipInstallPrefix,
			strings.Join(funcSpec.ExtendedMetaData.ServeDeploySchema.
				Applications[constant.ApplicationIndex].RuntimeEnv.Pip, " "), constant.PipCheckSuffix)
		createOpt[constant.PostStartExec] = installCommand
	}
	return nil
}

func setCreateOptionForDownloadData(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.FuncMetaData.BusinessType == constant.BusinessTypeServe &&
		len(funcSpec.ExtendedMetaData.ServeDeploySchema.Applications) != 0 {
		codeMetaData := commonTypes.CodeMetaData{
			LocalMetaData: commonTypes.LocalMetaData{
				StorageType: constant.WorkingDirType,
				CodePath: funcSpec.ExtendedMetaData.ServeDeploySchema.
					Applications[constant.ApplicationIndex].RuntimeEnv.WorkingDir,
			},
		}
		delegateDownloadData, err := json.Marshal(codeMetaData)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal delegate download data error %s", err.Error())
			return err
		}
		createOpt[constant.DelegateDownloadKey] = string(delegateDownloadData)
		return nil
	}
	if !reflect.DeepEqual(funcSpec.S3MetaData, commonTypes.S3MetaData{}) {
		funcSpec.CodeMetaData = s3MetaDataConvert2CodeMetaData(funcSpec.S3MetaData)
	}
	if funcSpec.CodeMetaData.Sha512 == "" && funcSpec.FuncMetaData.CodeSha512 != "" {
		funcSpec.CodeMetaData.Sha512 = funcSpec.FuncMetaData.CodeSha512
	}
	if funcSpec.FuncMetaData.Runtime != constant.CustomContainerRuntimeType {
		delegateDownloadData, err := json.Marshal(funcSpec.CodeMetaData)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal delegate download data error %s", err.Error())
			return err
		}
		createOpt[constant.DelegateDownloadKey] = string(delegateDownloadData)
	}
	if len(funcSpec.FuncMetaData.Layers) != 0 {
		delegateLayerDownloadData, err := json.Marshal(funcSpec.FuncMetaData.Layers)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal delegate layer download data error %s", err.Error())
			return err
		}
		createOpt[constant.DelegateLayerDownloadKey] = string(delegateLayerDownloadData)
		log.GetLogger().Infof("generate delegate download config %s for "+
			"function %s", string(delegateLayerDownloadData), funcSpec.FuncKey)
	}
	return nil
}

func s3MetaDataConvert2CodeMetaData(s3MetaData commonTypes.S3MetaData) commonTypes.CodeMetaData {
	return commonTypes.CodeMetaData{
		Sha512:        "",
		LocalMetaData: commonTypes.LocalMetaData{StorageType: "s3"},
		S3MetaData:    s3MetaData,
	}
}

func setCreateOptionForLifeCycleDetached(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return nil
	}
	createOpt[constant.InstanceLifeCycle] = constant.InstanceLifeCycleDetached
	return nil
}

func setCreateOptionForDelegateMount(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.ExtendedMetaData.FuncMountConfig != nil &&
		len(funcSpec.ExtendedMetaData.FuncMountConfig.FuncMounts) != 0 {
		bytesData, err := json.Marshal(funcSpec.ExtendedMetaData.FuncMountConfig)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal func mount config error: %s", err.Error())
			return err
		}
		createOpt[constant.DelegateMountKey] = string(bytesData)
		log.GetLogger().Infof("generate delegate mount config %s for function %s", string(bytesData), funcSpec.FuncKey)
	}
	return nil
}

func setCreateOptionForUserAgencyAndEnv(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.FuncMetaData.BusinessType == constant.BusinessTypeServe &&
		len(funcSpec.ExtendedMetaData.ServeDeploySchema.Applications) != 0 &&
		len(funcSpec.ExtendedMetaData.ServeDeploySchema.Applications[constant.ApplicationIndex].
			RuntimeEnv.EnvVars) != 0 {
		envVar, err := json.Marshal(funcSpec.ExtendedMetaData.ServeDeploySchema.
			Applications[constant.ApplicationIndex].RuntimeEnv.EnvVars)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal env var for %s", funcSpec.FuncKey)
			return err
		}
		createOpt[constant.DelegateEnvVar] = string(envVar)
		return nil
	}
	userAgency := funcSpec.ExtendedMetaData.UserAgency
	encryptMap := map[string]string{
		"secretKey": userAgency.SecretKey,
		"accessKey": userAgency.AccessKey, "authToken": userAgency.Token,
		"securityAk": userAgency.SecurityAk, "securitySk": userAgency.SecuritySk,
		"securityToken":       userAgency.SecurityToken,
		"environment":         funcSpec.EnvMetaData.Environment,
		"encrypted_user_data": funcSpec.EnvMetaData.EncryptedUserData,
		"envKey":              funcSpec.EnvMetaData.EnvKey,
		"cryptoAlgorithm":     funcSpec.EnvMetaData.CryptoAlgorithm,
	}
	encryptData, err := json.Marshal(encryptMap)
	if err != nil {
		log.GetLogger().Errorf("encryptData json marshal failed, err:%s", err.Error())
		return err
	}
	createOpt[constant.DelegateEncryptKey] = string(encryptData)
	log.GetLogger().Infof("generate delegate encrypt config %s for function %s", string(encryptData), funcSpec.FuncKey)

	if funcSpec.ExtendedMetaData.EnableAgentSession {
		agentSessionEnv := map[string]string{"ENABLE_AGENT_SESSION": strconv.FormatBool(funcSpec.ExtendedMetaData.EnableAgentSession)}
		agentSessionEnvBytes, err := json.Marshal(agentSessionEnv)
		if err != nil {
			log.GetLogger().Errorf("failed to marshal enable_agent_session env for %s", funcSpec.FuncKey)
			return err
		}
		createOpt[constant.DelegateEnvVar] = string(agentSessionEnvBytes)
		log.GetLogger().Infof("generate delegate env var config %s for function %s", string(agentSessionEnvBytes), funcSpec.FuncKey)
	}

	return nil
}

func getCustomContainerImagePullPolicy() v1.PullPolicy {
	customContainerImagePullPolicy := os.Getenv("CUSTOM_CONTAINER_IMAGE_PULL_POLICY")
	switch v1.PullPolicy(customContainerImagePullPolicy) {
	case v1.PullIfNotPresent, v1.PullAlways, v1.PullNever:
		return v1.PullPolicy(customContainerImagePullPolicy)
	default:
		return v1.PullIfNotPresent
	}
}

func getCustomContainerServiceAccountName(agentID string) string {
	if agentID != "" {
		return "sa-" + agentID
	}
	return "default"
}

func setCreateOptionForDelegateContainer(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if reflect.DeepEqual(funcSpec.ExtendedMetaData.CustomContainerConfig, commonTypes.CustomContainerConfig{}) {
		return nil
	}

	delegateContainerConfig := types.DelegateContainerConfig{
		Image:                  funcSpec.ExtendedMetaData.CustomContainerConfig.Image,
		ImagePullPolicy:        getCustomContainerImagePullPolicy(),
		Command:                funcSpec.ExtendedMetaData.CustomContainerConfig.Command,
		Args:                   funcSpec.ExtendedMetaData.CustomContainerConfig.Args,
		UID:                    funcSpec.ExtendedMetaData.CustomContainerConfig.UID,
		GID:                    funcSpec.ExtendedMetaData.CustomContainerConfig.GID,
		CustomGracefulShutdown: funcSpec.ExtendedMetaData.CustomGracefulShutdown,
		Env:                    initCustomContainerEnv(funcSpec),
		ServiceAccountName:     getCustomContainerServiceAccountName(funcSpec.FuncMetaData.AgentID),
	}
	vb := newVolumeBuilder()
	ve := newEnvBuilder()
	setVolumeForDelegateContainer(funcSpec, vb)
	setEnvForDelegateContainer(funcSpec, ve)
	delegateContainerConfig.VolumeMounts = vb.mounts[containerDelegate]
	delegateContainerConfig.Env = append(delegateContainerConfig.Env, ve.envs[containerDelegate]...)
	delegateRuntimeManager, err := json.Marshal(map[string]interface{}{
		"env": ve.envs[containerDelegate],
	})
	if err != nil {
		log.GetLogger().Errorf("failed to marshal runtime manager config, error %s", err.Error())
		return err
	}
	createOpt[constant.DelegateRuntimeManagerTag] = string(delegateRuntimeManager)
	configData, err := json.Marshal(delegateContainerConfig)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal delegate container config, error %s", err.Error())
		return err
	}
	createOpt[constant.DelegateContainerKey] = string(configData)
	if funcSpec.ExtendedMetaData.CustomGracefulShutdown.MaxShutdownTimeout > 0 {
		createOpt[types.GracefulShutdownTime] = strconv.Itoa(funcSpec.ExtendedMetaData.CustomGracefulShutdown.MaxShutdownTimeout)
	} else {
		createOpt[types.GracefulShutdownTime] = strconv.Itoa(types.MaxShutdownTimeout)
	}

	log.GetLogger().Infof("generate delegate container config %s for function %s", string(configData), funcSpec.FuncKey)
	setVolumeForAgentAndRuntime(funcSpec, vb)
	volumeMountData, err := json.Marshal(vb.mounts[containerRuntimeManager])
	if err != nil {
		log.GetLogger().Errorf("failed to marshal runtime manager volume mount, error:%s", err.Error())
		return err
	}
	createOpt[constant.DelegateVolumeMountKey] = string(volumeMountData)
	agentVolumeMountData, err := json.Marshal(vb.mounts[containerFunctionAgent])
	if err != nil {
		log.GetLogger().Errorf("failed to marshal function agent volume mount, error:%s", err.Error())
		return err
	}
	createOpt[constant.DelegateAgentVolumeMountKey] = string(agentVolumeMountData)
	volumesData, err := json.Marshal(vb.volumes)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal runtime manager volume, error:%s", err.Error())
		return err
	}
	createOpt[constant.DelegateVolumesKey] = string(volumesData)
	log.GetLogger().Infof("generate delegate volume: %s volume mount: %s", string(volumesData),
		string(volumeMountData))
	return nil
}

func setVolumeForAgentAndRuntime(funcSpec *types.FunctionSpecification, vb *volumeBuilder) {
	(&cgroupMemory{}).configVolume(vb)
	(&dockerSocket{}).configVolume(vb)
	(&dockerRootDir{}).configVolume(vb)
	if config.GlobalConfig.Scenario == types.ScenarioWiseCloud {
		if funcSpec.StsMetaData.EnableSts {
			log.GetLogger().Infof("start to build sts delegate for function %s", funcSpec.FuncKey)
			podRequest := types.PodRequest{
				FunSvcID:  funcSpec.FuncSecretName,
				NameSpace: constant.DefaultNameSpace,
			}
			(&stsSecret{
				param: funcSpec,
				req:   podRequest,
			}).configVolume(vb)
		}
		(&faasAgentSts{
			crName: urnutils.CrNameByURN(funcSpec.FuncMetaData.FunctionVersionURN),
		}).configVolume(vb)
	}
}

func setCreateOptionForHostAliases(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if createOpt == nil || len(config.GlobalConfig.HostAliases) == 0 {
		return nil
	}
	hostAliases := generateCustomHostAliases(config.GlobalConfig.HostAliases)
	hostAliasesData, err := json.Marshal(hostAliases)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal host aliases data, error:%s", err.Error())
		return err
	}
	createOpt[constant.DelegateHostAliases] = string(hostAliasesData)
	log.GetLogger().Infof("generate delegate host aliases: %+v "+
		"for function: %s", string(hostAliasesData), funcSpec.FuncKey)
	return nil
}

func setCreateOptionForRASP(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
		return nil
	}
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	add, createErr := initContainerAdd(funcSpec)
	if createErr != nil {
		log.GetLogger().Errorf(fmt.Sprintf("create init container error, %s", createErr.Error()))
	}
	createOpt[constant.DelegateInitContainers] = string(add)
	return nil
}

func setCreateOptionForOtel(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if !funcSpec.ExtendedMetaData.UserOtelConfig.Enable {
		return nil
	}
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	add, createErr := initContainerAdd(funcSpec)
	if createErr != nil {
		log.GetLogger().Errorf(fmt.Sprintf("create init container error, %s", createErr.Error()))
		return createErr
	}
	createOpt[constant.DelegateInitContainers] = string(add)
	return nil
}

func setCreateOptionForFileBeat(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
		return nil
	}
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	add, createErr := sideCarAdd(funcSpec)
	if createErr != nil {
		log.GetLogger().Errorf(fmt.Sprintf("create sideCar error, %s", createErr.Error()))
		return createErr
	}
	createOpt[constant.DelegateContainerSideCars] = string(add)
	return nil
}

func setCustomPodSeccompProfile(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
		return nil
	}
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.FuncMetaData.Runtime != types.CustomContainerRuntimeType {
		return nil
	}
	seccompProfile := &v1.SeccompProfile{Type: v1.SeccompProfileTypeRuntimeDefault}
	seccompProfileData, err := json.Marshal(seccompProfile)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal seccompProfile data, error:%s", err.Error())
		return err
	}
	createOpt[constant.DelegatePodSeccompProfile] = string(seccompProfileData)
	log.GetLogger().Infof("generate delegate seccompProfile: %+v "+
		"for function: %s", string(seccompProfileData), funcSpec.FuncKey)
	return nil
}

func setFunctionAgentInitContainer(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	if config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
		return nil
	}
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if funcSpec.FuncMetaData.Runtime != types.CustomContainerRuntimeType {
		return nil
	}
	vb := newVolumeBuilder()
	vb.addVolumeMount(containerFunctionAgentInit, v1.VolumeMount{
		Name:        biLogsVolume,
		MountPath:   biLogVolumeMountPath,
		SubPathExpr: biLogVolumeMountSubPathExpr,
	})
	volumeMountData, err := json.Marshal(vb.mounts[containerFunctionAgentInit])
	if err != nil {
		log.GetLogger().Errorf("failed to marshal function agent init volumeMount data, error:%s", err.Error())
		return err
	}
	createOpt[constant.DelegateInitVolumeMounts] = string(volumeMountData)
	log.GetLogger().Infof("generate function agent init volumeMount: %+v "+
		"for function: %s", string(volumeMountData), funcSpec.FuncKey)
	return nil
}

func setCreateOptionForInitContainerEnv(funcSpec *types.FunctionSpecification, createOpt map[string]string) error {
	logger := log.GetLogger().With(zap.Any("function", funcSpec.FuncKey))
	npuType, _ := utils.GetNpuTypeAndInstanceTypeFromStr(funcSpec.ResourceMetaData.CustomResources,
		funcSpec.ResourceMetaData.CustomResourcesSpec)
	var initContainerEnvs []v1.EnvVar
	if npuType == types.AscendResourceD910B {
		initContainerEnvs = append(initContainerEnvs, v1.EnvVar{
			Name:  "GENERATE_RANKTABLE_FILE",
			Value: "true",
		})
	}
	if funcSpec.ExtendedMetaData.UserOtelConfig.Enable {
		otelEnvs, err := getFunctionAgentInitOtelEnv(funcSpec)
		if err != nil {
			return err
		}
		initContainerEnvs = append(initContainerEnvs, otelEnvs...)
	}
	initContainerEnvsRaw, err := json.Marshal(initContainerEnvs)
	if err != nil {
		logger.Errorf("failed to marshal function agent init envs data, error:%s", err.Error())
		return err
	}
	if initContainerEnvsRaw != nil {
		createOpt[constant.DelegateInitEnv] = string(initContainerEnvsRaw)
		logger.Infof("generate function agent init envs: %+v ", string(initContainerEnvsRaw))
	}
	return nil
}

func setCreateOptionForLabel(instanceType types.InstanceType, funcSpec *types.FunctionSpecification,
	resSpec *resspeckey.ResourceSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	labels, err := getPodLabel(funcSpec, resSpec, instanceType)
	if err != nil {
		log.GetLogger().Errorf("get pod labels failed, err:%s", err.Error())
		return err
	}
	log.GetLogger().Infof("pod labels is: %s", string(labels))
	if resSpec != nil && resSpec.InvokeLabel != "" {
		createOpt[types.InstanceLabelNode] = resSpec.InvokeLabel
	}
	createOpt[constant.DelegatePodLabels] = string(labels)
	podInitLabels := map[string]string{
		podLabelSecurityGroup: strings.Split(funcSpec.FuncKey, "/")[0],
	}
	initLabels, err := json.Marshal(podInitLabels)
	if err != nil {
		log.GetLogger().Errorf("pod init labels json marshal failed, err:%s", err.Error())
		return err
	}
	createOpt[constant.DelegatePodInitLabels] = string(initLabels)
	return nil
}

func getPodLabel(funcSpec *types.FunctionSpecification, resSpec *resspeckey.ResourceSpecification,
	instanceType types.InstanceType) ([]byte, error) {
	version := funcSpec.FuncMetaData.Version
	// $ is an illegal character in k8s label
	if strings.HasPrefix(version, "$") {
		version = strings.TrimPrefix(version, "$")
	}
	podLabels := map[string]string{
		podLabelInstanceType: string(instanceType),
		podLabelFuncName:     funcSpec.FuncMetaData.FuncName,
		podLabelIsPoolPod:    "false",
		podLabelServiceID:    funcSpec.FuncMetaData.Service,
		podLabelTenantID:     strings.Split(funcSpec.FuncKey, "/")[0],
		podLabelVersion:      version,
	}
	if resSpec != nil {
		resSpecString := fmt.Sprintf("%d-%d-fusion", resSpec.CPU, resSpec.Memory)
		podLabels[podLabelStandard] = resSpecString
	}
	if resSpec != nil && resSpec.InvokeLabel != "" {
		podLabels[types.HeaderInstanceLabel] = resSpec.InvokeLabel
	}
	labels, err := json.Marshal(podLabels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func setCreateOptionForNote(instanceType types.InstanceType, funcSpec *types.FunctionSpecification,
	resSpec *resspeckey.ResourceSpecification, createOpt map[string]string) error {
	if createOpt == nil {
		return nil
	}
	resSpecData, err := json.Marshal(resSpec)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal resourceSpecification error %s", err.Error())
		return err
	}
	createOpt[types.ResourceSpecNote] = string(resSpecData)
	createOpt[types.FunctionKeyNote] = funcSpec.FuncKey
	if funcSpec.FuncMetaData.BusinessType == constant.BusinessTypeServe {
		createOpt[constant.BusinessTypeTypeNote] = constant.BusinessTypeServe
	}
	if selfregister.GlobalSchedulerProxy.IsFuncOwner(funcSpec.FuncKey) {
		createOpt[types.SchedulerIDNote] = selfregister.SelfInstanceID + types.TemporaryInstance
	} else {
		createOpt[types.SchedulerIDNote] = selfregister.SelfInstanceID + types.PermanentInstance
	}
	createOpt[types.InstanceTypeNote] = string(instanceType)
	createOpt[types.TenantID] = strings.Split(funcSpec.FuncKey, "/")[0]
	return nil
}

func buildDelegateNodeAffinity(xpuNodeLabel types.XpuNodeLabel) *v1.NodeAffinity {
	return &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
			{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      xpuNodeLabel.NodeLabelKey,
						Operator: "In",
						Values:   xpuNodeLabel.NodeLabelValues,
					},
				},
			},
		}},
	}
}

func setCreateOptionForAscendNPU(funcSpec *types.FunctionSpecification, resSpec *resspeckey.ResourceSpecification,
	createOpt map[string]string) error {
	if createOpt == nil || resSpec == nil {
		return nil
	}
	npuCores := 0
	logger := log.GetLogger().With(zap.Any("funcKey", funcSpec.FuncKey))

	customResourcesSpecStr := funcSpec.ResourceMetaData.CustomResourcesSpec
	customResourceStr := funcSpec.ResourceMetaData.CustomResources
	npuType, npuInstanceType := utils.GetNpuTypeAndInstanceTypeFromStr(customResourceStr, customResourcesSpecStr)
	if npuType == "" {
		return nil
	}
	for _, v := range config.GlobalConfig.XpuNodeLabels {
		if v.XpuType == npuType && v.InstanceType == npuInstanceType {
			nodeAffinity := buildDelegateNodeAffinity(v)
			jsonBytes, err := json.Marshal(nodeAffinity)
			if err != nil {
				return err
			}
			createOpt[constant.DelegateNodeAffinity] = string(jsonBytes)
			createOpt[constant.DelegateNodeAffinityPolicy] = constant.DelegateNodeAffinityPolicyAggregation
			break
		}
	}

	for k, v := range resSpec.CustomResources {
		if strings.Contains(k, types.AscendResourcePrefix) {
			npuType = k
			npuCores = int(v)
		}
	}
	if npuCores == 0 {
		return nil
	}
	npuMask := uint(0)
	for i := 0; i < npuCores; i++ {
		npuMask += 1 << i
	}
	data, err := json.Marshal(map[string]string{cceAscendAnnotation: fmt.Sprintf("%s=0x%x", npuType, npuMask)})
	if err != nil {
		logger.Errorf("failed to marshal ascend pod annotation, error %s", err.Error())
		return err
	}
	createOpt[constant.DelegatePodAnnotations] = string(data)
	return nil
}

func setCreateOptionForName(instanceName, callerPodName string, createOpt map[string]string) error {
	if createOpt == nil {
		return nil
	}
	if len(instanceName) != 0 {
		createOpt[types.InstanceNameNote] = instanceName
	}
	if len(callerPodName) != 0 {
		labelsData, exist := createOpt[constant.DelegatePodLabels]
		data := &map[string]string{}
		if exist {
			err := json.Unmarshal([]byte(labelsData), data)
			if err != nil {
				log.GetLogger().Errorf("Unmarshal instance labels error")
				return err
			}
		}
		(*data)["callerPodName"] = callerPodName
		labels, err := json.Marshal(data)
		if err != nil {
			log.GetLogger().Errorf("pod labels json marshal failed, err:%s", err.Error())
			return err
		}
		if createOpt != nil {
			createOpt[constant.DelegatePodLabels] = string(labels)
		}
	}
	return nil
}

func prepareCreateArguments(request createInstanceRequest) []api.Arg {
	funcSpecCopy := &types.FunctionSpecification{}
	commonUtils.DeepCopyObj(request.funcSpec, funcSpecCopy)
	funcSpecCopy.ResourceMetaData.CPU = request.resKey.CPU
	funcSpecCopy.ResourceMetaData.Memory = request.resKey.Memory
	funcSpecData, err := json.Marshal(funcSpecCopy)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal create params error %s", err.Error())
		return nil
	}
	createParamsData, err := prepareCreateParamsData(funcSpecCopy, request.resKey)
	if err != nil {
		log.GetLogger().Errorf("failed to prepare creatParamsData error %s", err.Error())
		return nil
	}
	schedulerData, err := signalmanager.PrepareSchedulerArg()
	if err != nil {
		log.GetLogger().Errorf("failed to prepare scheduler params error %s", err.Error())
		return nil
	}
	args := []api.Arg{
		{
			Type: api.Value,
			Data: funcSpecData,
		},
		{
			Type: api.Value,
			Data: createParamsData,
		},
		{
			Type: api.Value,
			Data: schedulerData,
		},
		{
			Type: api.Value,
			Data: request.createEvent,
		},
	}
	if request.funcSpec.FuncMetaData.Runtime == types.CustomContainerRuntimeType {
		customUserData, err := prepareCustomUserArg(funcSpecCopy)
		if err != nil {
			log.GetLogger().Errorf("failed to prepare custom user data error %s", err.Error())
			return nil
		}
		args = append(args, api.Arg{
			Type: api.Value,
			Data: customUserData,
		})
	}
	return args
}

func prepareCreateParamsData(funcSpec *types.FunctionSpecification, resKey resspeckey.ResSpecKey) ([]byte, error) {
	var createParamsData []byte
	var err error
	if strings.Contains(funcSpec.FuncMetaData.Runtime, types.HTTPRuntimeType) ||
		strings.Contains(funcSpec.FuncMetaData.Runtime, types.CustomContainerRuntimeType) {
		createParams := CreateParams{
			InstanceLabel: resKey.InvokeLabel,
			HTTPCreateParams: HTTPCreateParams{
				Port:      types.HTTPFuncPort,
				CallRoute: types.HTTPCallRoute,
			},
		}
		createParamsData, err = json.Marshal(createParams)
	} else {
		userInitEntry := funcSpec.ExtendedMetaData.Initializer.Handler
		userCallEntry := funcSpec.FuncMetaData.Handler
		if len(userCallEntry) == 0 {
			log.GetLogger().Warnf("user call entry for function %s is empty", funcSpec.FuncKey)
		}
		createParams := CreateParams{
			InstanceLabel: resKey.InvokeLabel,
			EventCreateParams: EventCreateParams{
				UserInitEntry: userInitEntry,
				UserCallEntry: userCallEntry,
			},
		}
		createParamsData, err = json.Marshal(createParams)
	}
	if err != nil {
		log.GetLogger().Errorf("failed to marshal create params error %s", err.Error())
		return nil, err
	}
	return createParamsData, nil
}

func prepareCustomUserArg(funcSpec *types.FunctionSpecification) ([]byte, error) {
	faasExecutorStsServerConfig := sts.GetStsServerConfig(funcSpec)
	localAuth := localauth.AuthConfig{
		AKey:     config.GlobalConfig.LocalAuth.AKey,
		SKey:     config.GlobalConfig.LocalAuth.SKey,
		Duration: config.GlobalConfig.LocalAuth.Duration,
	}
	customUserArgInfo := &types.CustomUserArgs{
		AlarmConfig:       config.GlobalConfig.AlarmConfig,
		ClusterName:       config.GlobalConfig.ClusterName,
		DiskMonitorEnable: config.GlobalConfig.DiskMonitorEnable,
		LocalAuth:         localAuth,
		StsServerConfig:   faasExecutorStsServerConfig,
	}
	customUserArgData, err := json.Marshal(customUserArgInfo)
	if err != nil {
		return nil, err
	}
	return customUserArgData, nil
}

func hasD910b(resSpec *resspeckey.ResourceSpecification) bool {
	if resSpec == nil {
		return false
	}
	d910bstr, _ := utils.GetNpuTypeAndInstanceType(resSpec.CustomResources, resSpec.CustomResourcesSpec)
	if d910bstr == types.AscendResourceD910B {
		return true
	}
	return false
}

func setEphemeralStorage(defaultES, configES float64, resourcesMap map[string]float64) {
	if resourcesMap == nil {
		resourcesMap = make(map[string]float64)
	}
	if configES == 0 {
		resourcesMap[resourcesEphemeralStorage] = defaultES
		return
	}
	resourcesMap[resourcesEphemeralStorage] = configES
}

func setCreateOptionForFuncSpec(funcSpec *types.FunctionSpecification, createOpt map[string]string) {
	if createOpt == nil {
		return
	}
	createOpt[types.ConcurrentNumKey] = strconv.Itoa(funcSpec.InstanceMetaData.ConcurrentNum)
	createOpt[constant.DelegateDirectoryInfo] = defaultDelegateDirectoryInfo
	if funcSpec.InstanceMetaData.DiskLimit == 0 {
		createOpt[constant.DelegateDirectoryQuota] = strconv.Itoa(defaultDirectoryLimit)
	} else {
		createOpt[constant.DelegateDirectoryQuota] = strconv.FormatInt(funcSpec.InstanceMetaData.DiskLimit, base)
	}

	createOpt[types.InitCallTimeoutKey] = strconv.Itoa(int(funcSpec.ExtendedMetaData.Initializer.Timeout))
	createOpt[types.CallTimeoutKey] = strconv.Itoa(int(funcSpec.FuncMetaData.Timeout))
	createOpt[types.FunctionSign] = funcSpec.FuncMetaSignature
}

func initCustomContainerEnv(funcSpec *types.FunctionSpecification) []v1.EnvVar {
	parsedURN, err := urnutils.GetFunctionInfo(funcSpec.FuncMetaData.FunctionURN)
	if err != nil {
		log.GetLogger().Errorf("getFunctionInfo error: %s", err.Error())
	}

	envs := []v1.EnvVar{
		{Name: invokeTypeEnvName, Value: invokeTypeEnvValue},
		// Bi environment variable
		{Name: biEvnTenantID, Value: parsedURN.TenantID},
		{Name: biEvnFunctionName, Value: getNoPrefixFuncName(funcSpec.FuncMetaData.Name)},
		{Name: biEvnFunctionVersion, Value: funcSpec.FuncMetaData.Version},
		{Name: biEvnRegion, Value: config.GlobalConfig.RegionName},
		{Name: biEvnClusterID, Value: os.Getenv("CLUSTER_ID")},
		{Name: biEvnNodeIP, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"},
		}},
		{Name: biEvnPodName, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"},
		}},
		{Name: podIPEnv, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"},
		}},
		{Name: hostIPEnv, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"},
		}},
		{Name: podNameEnv, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"},
		}},
		{Name: podIDEnv, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.uid"},
		}},
		{Name: podNameEnvNew, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"},
		}},
	}
	for envKey, envValue := range config.GlobalConfig.CustomContainerEnv {
		envs = append(envs, v1.EnvVar{Name: envKey, Value: envValue})
	}
	if funcSpec.ExtendedMetaData.UserOtelConfig.Enable {
		envs = addOtelSystemEnv(envs)
		for envKey, envValue := range funcSpec.ExtendedMetaData.UserOtelConfig.OtelEnv {
			envs = append(envs, v1.EnvVar{Name: envKey, Value: envValue})
		}
	}
	npuType, npuInstanceType := utils.GetNpuTypeAndInstanceTypeFromStr(
		funcSpec.ResourceMetaData.CustomResources, funcSpec.ResourceMetaData.CustomResourcesSpec)
	if npuType != "" {
		envs = append(envs, v1.EnvVar{
			Name: types.SystemNodeInstanceType, Value: npuInstanceType,
		})
	}
	return envs
}

func getNoPrefixFuncName(name string) string {
	lastIndex := strings.LastIndex(name, "@")
	if lastIndex > 0 {
		return name[lastIndex+1:]
	}
	return name
}

func setVolumeForDelegateContainer(funcSpec *types.FunctionSpecification, vb *volumeBuilder) {
	if strings.Contains(funcSpec.ResourceMetaData.CustomResources, types.AscendResourcePrefix) {
		(&ascendConfig{}).configVolume(vb)
		if config.GlobalConfig.EnableNPUDriverMount {
			(&npuDriver{}).configVolume(vb)
		}
	}
	if len(config.GlobalConfig.FunctionConfig) != 0 {
		for _, fc := range config.GlobalConfig.FunctionConfig {
			(&functionDefaultConfig{
				configName: fc.ConfigName,
				mount:      fc.Mount,
			}).configVolume(vb)
		}
	}
	if funcSpec.FuncMetaData.BusinessType == constant.BusinessTypeAgent {
		(&dataSystemSocket{}).configVolume(vb)
		if funcSpec.FuncMetaData.AgentID != "" {
			(&serviceAccountToken{}).configVolume(vb)
		}
		if funcSpec.ExtendedMetaData.UserOtelConfig.Enable {
			(&otelShared{}).configVolume(vb, funcSpec.ExtendedMetaData.UserOtelConfig.InitContainer.ShardDir)
		}
	}
}

func setEnvForDelegateContainer(funcSpec *types.FunctionSpecification, eb *envBuilder) {
	if funcSpec.StsMetaData.EnableSts {
		configEnv(eb, funcSpec.StsMetaData.SensitiveConfigs)
	}
	if strings.Contains(funcSpec.ResourceMetaData.CustomResources, types.AscendResourcePrefix) {
		eb.addEnvVar(containerDelegate, v1.EnvVar{
			Name:  types.AscendRankTableFileEnvKey,
			Value: types.AscendRankTableFileEnvValue,
		})
	}
}

func generateCustomHostAliases(hostAliases []v1.HostAlias) map[string][]string {
	// key: ip, value: hosts
	hostAliasMap := make(map[string][]string, constant.DefaultMapSize)
	for _, hostAlias := range hostAliases {
		if _, exist := hostAliasMap[hostAlias.IP]; !exist {
			hostAliasMap[hostAlias.IP] = make([]string, 0, constant.DefaultHostAliasesSliceSize)
		}
		hostAliasMap[hostAlias.IP] = append(hostAliasMap[hostAlias.IP], hostAlias.Hostnames...)
	}
	return hostAliasMap
}

func setCreateOptionForVPC(createOption map[string]string, natConfigs []commonTypes.NATConfigure) {
	if len(natConfigs) == 0 {
		return
	}
	if createOption == nil {
		createOption = make(map[string]string, utils.DefaultMapSize)
	}
	var networkConfigs []types.NetworkConfig
	delegateNetworkConfig := types.DelegateNetworkConfig{}
	for _, natConfig := range natConfigs {
		networkConfigs = append(networkConfigs, generateNetworkConfig(natConfig))
		patInstance := types.PatInstance{
			Name:      natConfig.PatPodName,
			Namespace: natConfig.Namespace,
		}
		delegateNetworkConfig.PatInstances = append(delegateNetworkConfig.PatInstances, patInstance)
	}

	networkConfigData, err := json.Marshal(networkConfigs)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal network config for pat service %v", natConfigs)
		return
	}
	createOption[types.NetworkConfigKey] = string(networkConfigData)

	delegateNetworkConfigData, err := json.Marshal(delegateNetworkConfig)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal delegate network config for pat service %v", natConfigs)
		return
	}
	createOption[constant.DelegateNetworkConfig] = string(delegateNetworkConfigData)

	log.GetLogger().Infof("succeed to set createOption networkConfig: %s, delegateNetworkConfig: %s",
		createOption[types.NetworkConfigKey], createOption[constant.DelegateNetworkConfig])
}

func generateNetworkConfig(natConfig commonTypes.NATConfigure) types.NetworkConfig {
	netConfig := types.NetworkConfig{
		RouteConfig: types.RouteConfig{
			Gateway: natConfig.PatContainerIP,
			Cidr:    "0.0.0.0/0",
		},
		TunnelConfig: types.TunnelConfig{
			TunnelName: types.FgsTunnelName,
			RemoteIP:   natConfig.PatContainerIP,
			Mode:       types.IPIPMode,
		},
		FirewallConfig: types.FirewallConfig{
			Chain:     "OUTPUT",
			Table:     "filter",
			Operation: "add",
			Target:    natConfig.PatContainerIP,
			Args:      "-j ACCEPT",
		},
		ProberConfig: types.ProberConfig{
			Protocol:         types.ICMPProtocol,
			Address:          natConfig.PatContainerIP,
			Gateway:          natConfig.PatGateway,
			SubMetaDigest:    natConfig.SubMetaDigest,
			Interval:         patProberInterval,
			Timeout:          patProberTimeout,
			FailureThreshold: patProberFailureThreshold,
		},
	}
	return netConfig
}

func setCreateOptionForNuwaRuntimeInfo(nuwaRuntimeInfo *wisecloudTypes.NuwaRuntimeInfo,
	createOpt map[string]string) error {
	if config.GlobalConfig.Scenario != types.ScenarioWiseCloud {
		return nil
	}
	if createOpt == nil {
		return errors.New("createOpt is nil")
	}
	if nuwaRuntimeInfo == nil {
		return errors.New("nuwa runtimeinfo is nil")
	}
	nuwaRuntimeInfoData, err := json.Marshal(nuwaRuntimeInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal nuwa runtime info for %v", nuwaRuntimeInfo)
		return err
	}
	createOpt[constant.DelegateNuwaRuntimeInfo] = string(nuwaRuntimeInfoData)
	return nil
}

func prepareCreatePATServiceArguments(traceID string, namespace string,
	extMetaData commonTypes.ExtendedMetaData, vpcConfig *commonTypes.VpcConfig) []api.Arg {
	patSvcReq := commonTypes.PATServiceRequest{
		ID:             vpcConfig.ID,
		Namespace:      namespace,
		DomainID:       vpcConfig.DomainID,
		ProjectID:      vpcConfig.Namespace,
		VpcID:          vpcConfig.VpcID,
		SubnetID:       vpcConfig.SubnetID,
		TenantCidr:     vpcConfig.TenantCidr,
		HostVMCidr:     vpcConfig.HostVMCidr,
		Gateway:        vpcConfig.Gateway,
		SecurityGroups: vpcConfig.SecurityGroups,
		Xrole:          extMetaData.Role.XRole,
	}
	patSvcReqData, err := json.Marshal(patSvcReq)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal vpc config, error %s", err.Error())
	}
	return []api.Arg{
		{
			Type: api.Value,
			Data: []byte(vpcOpCreatePATService),
		},
		{
			Type: api.Value,
			Data: patSvcReqData,
		},
		{
			Type: api.Value,
			Data: []byte(traceID),
		},
	}
}

func generateSnErrorFromKernelError(kernelErr error) snerror.SNError {
	var kernelErrCode int
	var kernelErrMsg string
	if snErr, ok := kernelErr.(api.ErrorInfo); ok {
		kernelErrCode = snErr.Code
		kernelErrMsg = snErr.Error()
	} else {
		kernelErrCode = statuscode.GetKernelErrorCode(kernelErr.Error())
		kernelErrMsg = statuscode.GetKernelErrorMessage(kernelErr.Error())
	}
	// if user error, return original errorCode
	if kernelErrCode < snerror.UserErrorMax && kernelErrCode >= snerror.UserErrorMin {
		return snerror.New(kernelErrCode, kernelErrMsg)
	}
	// posix errorCode from functionsystem
	if kernelErrCode == constant.KernelUserCodeLoadErrCode {
		return snerror.New(statuscode.UserFuncEntryNotFoundErrCode, kernelErrMsg)
	}
	if kernelErrCode == constant.KernelCreateLimitErrCode {
		return snerror.New(statuscode.CreateLimitErrorCode, kernelErrMsg)
	}
	if kernelErrCode == constant.KernelWriteEtcdCircuitErrCode {
		return snerror.New(statuscode.KernelEtcdWriteFailedCode, kernelErrMsg)
	}
	if kernelErrCode == constant.KernelResourceNotEnoughErrCode {
		return snerror.New(statuscode.KernelResourceNotEnoughErrCode, kernelErrMsg)
	}

	if kernelErrCode == constant.KernelRequestErrBetweenRuntimeAndBus &&
		strings.Contains(kernelErrMsg, "reason(timeout)") {
		return snerror.New(statuscode.UserFuncInitTimeoutCode, "runtime initialization timed out")
	}

	// faas-executor returns two types of error: inner-system-error and user-function-error,
	// which contains message in JSON format
	snErr := snerror.New(statuscode.InternalErrorCode, kernelErrMsg)
	if kernelErrCode == constant.KernelInnerSystemErrCode ||
		kernelErrCode == constant.KernelUserFunctionExceptionErrCode {
		initRsp := &types.ExecutorInitResponse{}
		err := json.Unmarshal([]byte(kernelErrMsg), initRsp)
		if err != nil {
			log.GetLogger().Errorf("json unMarshal initRsp error: %s", err.Error())
			return snErr
		}
		errCode, err := strconv.Atoi(initRsp.ErrorCode)
		if err != nil {
			log.GetLogger().Errorf("initRsp errorCode invalid: %s", err.Error())
			return snErr
		}
		errMsg, err := initRsp.Message.MarshalJSON()
		if err != nil {
			log.GetLogger().Errorf("initRsp message marshal error: %s", err.Error())
			return snErr
		}
		return snerror.New(errCode, string(errMsg))
	}
	return snErr
}

func handlePullTriggerCreate(faasMgrInfo faasManagerInfo, funcSpec *types.FunctionSpecification) {
	if len(faasMgrInfo.funcKey) == 0 || len(faasMgrInfo.instanceID) == 0 {
		log.GetLogger().Errorf("no faas-manager instance info")
		return
	}
	createCh := make(chan error, 1)
	log.GetLogger().Infof("start to prepare arguments %s", funcSpec.FuncKey)
	args := prepareCreatePullTriggerArguments(funcSpec)
	if args == nil {
		return
	}
	traceID := utils.GenerateTraceID()
	log.GetLogger().Infof("start to invoke manager %s, instance %s, traceID: %s", faasMgrInfo.funcKey,
		faasMgrInfo.instanceID, traceID)
	var invokeErr error
	funcMeta := api.FunctionMeta{FuncID: faasMgrInfo.funcKey, Api: api.PosixApi}
	invokeOpts := api.InvokeOptions{TraceID: traceID}
	objID, errorInfo := globalSdkClient.InvokeByInstanceId(funcMeta, faasMgrInfo.instanceID, args, invokeOpts)
	invokeErr = errorInfo
	if invokeErr == nil {
		globalSdkClient.GetAsync(objID, func(result []byte, err error) {
			createCh <- err
		})
	}
	if invokeErr != nil {
		log.GetLogger().Errorf("failed to send create request of vpcPullTrigger %s, traceID: %s, error: %s",
			funcSpec.FuncKey, traceID, invokeErr.Error())
		return
	}
	timer := time.NewTimer(vpcPullTriggerRequestTimeout)
	select {
	case resultErr, ok := <-createCh:
		if !ok {
			log.GetLogger().Errorf("result channel of create pullTrigger request is closed, traceID: %s", traceID)
		}
		timer.Stop()
		if resultErr != nil {
			log.GetLogger().Errorf("failed to create pullTrigger %s, traceID: %s, error %s",
				funcSpec.FuncKey, traceID, resultErr.Error())
			return
		}
	case <-timer.C:
		log.GetLogger().Errorf("time out waiting for pullTrigger create of function %s, traceID %s",
			funcSpec.FuncKey, traceID)
		return
	}
}

func handlePullTriggerDelete(faasMgrInfo faasManagerInfo, funcSpec *types.FunctionSpecification) {
	traceID := utils.GenerateTraceID()
	log.GetLogger().Infof("handling vpc pull trigger %s delete, traceID: %s", funcSpec.FuncKey, traceID)
	if len(faasMgrInfo.funcKey) == 0 || len(faasMgrInfo.instanceID) == 0 {
		log.GetLogger().Errorf("no faas-manager instance info, traceID: %s", traceID)
		return
	}
	args := prepareDeletePullTriggerArguments(funcSpec.FuncKey)
	deleteCh := make(chan error, 1)
	var deleteErr error
	funcMeta := api.FunctionMeta{FuncID: faasMgrInfo.funcKey, Api: api.PosixApi}
	invokeOpts := api.InvokeOptions{TraceID: traceID}
	objID, errorInfo := globalSdkClient.InvokeByInstanceId(funcMeta, faasMgrInfo.instanceID, args, invokeOpts)
	deleteErr = errorInfo
	if deleteErr == nil {
		globalSdkClient.GetAsync(objID, func(result []byte, err error) {
			deleteCh <- err
		})
	}
	if deleteErr != nil {
		log.GetLogger().Errorf("failed to send delete request of vpcPullTrigger %s, traceID: %s, error %s",
			funcSpec.FuncKey, traceID, deleteErr.Error())
		return
	}
	timer := time.NewTimer(vpcPullTriggerRequestTimeout)
	select {
	case resultErr, ok := <-deleteCh:
		if !ok {
			log.GetLogger().Errorf("result channel of delete pullTrigger request is closed, traceID: %s", traceID)
			return
		}
		timer.Stop()
		if resultErr != nil {
			log.GetLogger().Errorf("failed to delete pullTrigger %s, traceID: %s, error %s",
				funcSpec.FuncKey, traceID, resultErr.Error())
			return
		}
	case <-timer.C:
		log.GetLogger().Errorf("time out waiting for pullTrigger delete of function %s, traceID: %s",
			funcSpec.FuncKey, traceID)
		return
	}
}

func prepareCreatePullTriggerArguments(funcSpec *types.FunctionSpecification) []api.Arg {
	if funcSpec.ExtendedMetaData.VpcConfig == nil {
		log.GetLogger().Errorf("failed to get vpc config")
		return nil
	}
	pullTriggerReq := types.PullTriggerRequestInfo{
		PodName:    funcSpec.FuncKey,
		Image:      funcSpec.FuncMetaData.VPCTriggerImage,
		DomainID:   funcSpec.ExtendedMetaData.VpcConfig.DomainID,
		Namespace:  funcSpec.ExtendedMetaData.VpcConfig.Namespace,
		VpcName:    funcSpec.ExtendedMetaData.VpcConfig.VpcName,
		VpcID:      funcSpec.ExtendedMetaData.VpcConfig.VpcID,
		SubnetName: funcSpec.ExtendedMetaData.VpcConfig.SubnetName,
		SubnetID:   funcSpec.ExtendedMetaData.VpcConfig.SubnetID,
		TenantCidr: funcSpec.ExtendedMetaData.VpcConfig.TenantCidr,
		HostVMCidr: funcSpec.ExtendedMetaData.VpcConfig.HostVMCidr,
		Gateway:    funcSpec.ExtendedMetaData.VpcConfig.Gateway,
		Xrole:      funcSpec.ExtendedMetaData.Role.XRole,
		AppXrole:   funcSpec.ExtendedMetaData.Role.AppXRole,
	}
	pullTriggerReqData, err := json.Marshal(pullTriggerReq)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal pullTrigger config, error %s", err.Error())
		return nil
	}
	return []api.Arg{
		{
			Type: api.Value,
			Data: []byte(vpcOpCreatePullTrigger),
		},
		{
			Type: api.Value,
			Data: pullTriggerReqData,
		},
	}
}

func prepareDeletePullTriggerArguments(funcKey string) []api.Arg {
	deleteInfo := types.PullTriggerDeleteInfo{
		PodName: funcKey,
	}
	deleteData, err := json.Marshal(deleteInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal delete info for funcKey %s, error %s", funcKey, err.Error())
	}
	return []api.Arg{
		{
			Type: api.Value,
			Data: []byte(vpcOpDeletePullTrigger),
		},
		{
			Type: api.Value,
			Data: deleteData,
		},
	}
}
