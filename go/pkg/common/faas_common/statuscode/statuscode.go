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

// Package statuscode define status code of Frontend
package statuscode

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

// system error code
const (
	// InnerResponseSuccessCode -
	InnerResponseSuccessCode = 0
	// InternalErrorCode if the value is 331404, try again.
	InternalErrorCode    = 330404
	InternalRetryErrCode = 331404
	InternalErrorMessage = "internal system error"

	// InnerInstanceCircuitCode need retry
	InnerInstanceCircuitCode = 4011

	// BackpressureCode indicate that frontend should choose another proxy/worker and retry
	BackpressureCode = 211429
)

// frontend error code
const (
	// FrontendStatusOk ok code
	FrontendStatusOk = 200200
	// FrontendStatusAccepted -
	FrontendStatusAccepted = 200202
	// FrontendStatusNoContent -
	FrontendStatusNoContent = 200204

	// FrontendStatusBadRequest -
	FrontendStatusBadRequest = 200400
	// FrontendStatusUnAuthorized -
	FrontendStatusUnAuthorized = 200401
	// FrontendStatusForbidden -
	FrontendStatusForbidden = 200403
	// FrontendStatusNotFound -
	FrontendStatusNotFound = 200404
	// FrontendStatusRequestEntityTooLarge -
	FrontendStatusRequestEntityTooLarge = 200413
	// FrontendStatusTooManyRequests -
	FrontendStatusTooManyRequests = 200429

	// FrontendStatusInternalError -
	FrontendStatusInternalError = 200500
	// HTTPStreamNOTEnableError -
	HTTPStreamNOTEnableError = 200600
	// CreateStreamProducerError -
	CreateStreamProducerError = 200601
	// QueryStreamCustomerError -
	QueryStreamCustomerError = 200602
	// SendDataToStreamError -
	SendDataToStreamError = 200603
	// WriteResponseError -
	WriteResponseError = 200604

	// DsUploadFailed - upload to data system failed
	DsUploadFailed = 200701
	// DsDownloadFailed - download from data system failed
	DsDownloadFailed = 200702
	// DsDeleteFailed - delete from data system failed
	DsDeleteFailed = 200703
	// DsKeyNotFound - key not found on data system
	DsKeyNotFound = 200704

	// UserFunctionInvokeError - user function error
	UserFunctionInvokeError = 200705

	// FuncMetaNotFound function meta not found, this error occurs only when the internal service is abnormal.
	FuncMetaNotFound = 150424
	// HeavyLoadCode indicate the server's memory usage reaches threshold
	HeavyLoadCode = 214503
)

// User error code
const (
	// UserFuncEntryNotFoundErrCode -
	UserFuncEntryNotFoundErrCode = 4001
	// UserFuncRunningExceptionErrCode -
	UserFuncRunningExceptionErrCode = 4002
	// StateContentTooLargeErrCode state content is too large
	StateContentTooLargeErrCode = 4003
	// UserFuncRspExceedLimitErrCode response of user function exceeds the platform limit
	UserFuncRspExceedLimitErrCode = 4004
	// UndefinedStateErrCode state is undefined
	UndefinedStateErrCode = 4005
	// HeartBeatFunctionInvalidErrCode heart beat function of user invalid
	HeartBeatFunctionInvalidErrCode = 4006
	// FunctionResultInvalidErrCode user function result is invalid
	FunctionResultInvalidErrCode = 4007
	// InitializeFunctionErrorErrCode user initialize function error
	InitializeFunctionErrorErrCode = 4009
	// UserFuncInvokeTimeout -
	UserFuncInvokeTimeout = 4010
	// FrontendStatusWorkerIoTimeout -
	FrontendStatusWorkerIoTimeout = 4014
	// FrontendStatusTrafficLimitEffective is the error code for traffic limitation
	FrontendStatusTrafficLimitEffective = 4021
	// FrontendStatusLabelUnavailable -
	FrontendStatusLabelUnavailable = 4022
	// FrontendStatusFuncMetaNotFound is error code of function meta not found
	FrontendStatusFuncMetaNotFound = 4024
	// FrontendStatusUnableSpecifyResource unable to specify resource in a scene where no resource specified
	FrontendStatusUnableSpecifyResource = 4026
	// FrontendStatusMaxRequestBodySize -
	FrontendStatusMaxRequestBodySize = 4140
	// UserFuncInitFailCode code of user function initialization failed
	UserFuncInitFailCode = 4201
	// ErrSharedMemoryLimited -
	ErrSharedMemoryLimited = 4202
	// ErrOperateDiskFailed -
	ErrOperateDiskFailed = 4203
	// ErrInsufficientDiskSpace -
	ErrInsufficientDiskSpace = 4204

	// UserFuncInitTimeoutCode code of initialing runtime timed out
	UserFuncInitTimeoutCode = 4211
	// StsConfigErrCode sts config set error code
	StsConfigErrCode = 4036
	// InstanceSessionInvalidErrCode -
	InstanceSessionInvalidErrCode = 4037
	// ErrFinalized -
	ErrFinalized = 9000
	// ErrAllSchedulerUnavailable -
	ErrAllSchedulerUnavailable = 9009
	// InnerUserErrBase -
	InnerUserErrBase = 50_0000
	// InnerRuntimeInitTimeoutCode -
	InnerRuntimeInitTimeoutCode = InnerUserErrBase + UserFuncInitTimeoutCode
)

// proxy internal error codes which suggests to retry in cluster
const (
	// ClientExitErrCode function instance  is exiting (proxy side)
	ClientExitErrCode = 211503

	// WorkerExitErrCode function instance is exiting (worker side)
	WorkerExitErrCode = 211504

	// UserFuncIsUpdatedCode -
	UserFuncIsUpdatedCode = 211411
	// SendReqErrCode call request sending error
	SendReqErrCode = 211406
)

// executor error code
const (
	// ExecutorErrCodeInitFail -
	ExecutorErrCodeInitFail = 6001
)

// The kernel and faaspattern should maintain an appropriate set of error codes.
// Common, such as a unified understanding of whether retry is required.
// In addition, the current transmission involves various character string conversions,
// which increases transcoding and matching barriers and causes high overheads.
// These are important, otherwise it will cause a lot of unclear boundaries and rework :)
const (
	// ErrInstanceNotFound -
	ErrInstanceNotFound = 1003
	// ErrInstanceExitedCode -
	ErrInstanceExitedCode = 1007
	// ErrInstanceCircuitCode -
	ErrInstanceCircuitCode = 1009
	// ErrInstanceEvicted -
	ErrInstanceEvicted = 1013

	// ErrRequestBetweenRuntimeBusCode -
	ErrRequestBetweenRuntimeBusCode = 3001
	// ErrInnerCommunication -
	ErrInnerCommunication = 3002
	// ErrRequestBetweenRuntimeFrontendCode -
	ErrRequestBetweenRuntimeFrontendCode = 3008
	// ErrAcquireTimeoutCode -
	ErrAcquireTimeoutCode = 3009
)

// errors comes from faas scheduler (FG worker manager error)
const (
	// StatusInternalServerError status internal server error
	StatusInternalServerError = 150500
	// VIPClusterOverloadCode cluster has no available resource
	VIPClusterOverloadCode = 150510
	// FuncMetaNotFoundErrCode function meta not found, this error occurs only when the internal service is abnormal.
	FuncMetaNotFoundErrCode = 150424
	// FuncMetaNotFoundErrMsg is error message of function metadata not found
	FuncMetaNotFoundErrMsg = "function metadata not found"
	// InstanceNotFoundErrCode is error code of instance not found
	InstanceNotFoundErrCode = 150425
	// InstanceNotFoundErrMsg is error message of instance not found
	InstanceNotFoundErrMsg = "instance not exist"
	// NoInstanceAvailableErrCode is error message of no available instance
	NoInstanceAvailableErrCode = 150431
	// InstanceStatusAbnormalCode -
	InstanceStatusAbnormalCode = 150427
	// InstanceStatusAbnormalMsg -
	InstanceStatusAbnormalMsg = "instance status is abnormal"
	// ReachMaxInstancesCode reach function max instances
	ReachMaxInstancesCode = 150429
	// ReachMaxInstancesErrMsg is error message of reach max instance
	ReachMaxInstancesErrMsg = "reach max instance num"
	// InsThdReqTimeoutCode acquire instance lease timeout, FG: cluster is overload and unavailable now
	InsThdReqTimeoutCode = 150430
	// InsThdReqTimeoutErrMsg acquire instance lease timeout
	InsThdReqTimeoutErrMsg = "instance thread request timeout"
	// ReachMaxInstancesPerTenantErrCode reach tenant max on-demand instances
	ReachMaxInstancesPerTenantErrCode = 150432
	// GettingPodErrorCode getting pod error code
	GettingPodErrorCode = 150431
	// ReachMaxOnDemandInstancesPerTenant reach tenant max on-demand instances
	ReachMaxOnDemandInstancesPerTenant = 150432
	// ReachMaxInstancesPerTenantErrMsg reach tenant max on-demand instances
	ReachMaxInstancesPerTenantErrMsg = "reach max instance number per tenant"
	// ReachMaxReversedInstancesPerTenant reach tenant max reversed instances
	ReachMaxReversedInstancesPerTenant = 150433
	// FunctionIsDisabled function is disabled
	FunctionIsDisabled = 150434
	// RefreshSilentFunc waiting for silent function to refresh, retry required
	RefreshSilentFunc = 150435
	// NotEnoughNIC marked that there were not enough network cards
	NotEnoughNIC = 150436
	// InsufficientEphemeralStorage marked that ephemeral storage is insufficient
	InsufficientEphemeralStorage = 150438
	// ClusterIsUpgrading -
	ClusterIsUpgrading = 150439
	// DesignateInsNotAvailableErrCode -
	DesignateInsNotAvailableErrCode = 150440
	// InstanceLabelNotFoundErrCode -
	InstanceLabelNotFoundErrCode = 150444
	// InstanceLabelNotFoundErrMsg -
	InstanceLabelNotFoundErrMsg = "instance label not found"
	// CancelGeneralizePod user update function metadata to cancel generalize pod while generalizing is not finished
	CancelGeneralizePod = 150439

	// ScaleUpRequestErrCode failed to send scale up request to worker-manager
	ScaleUpRequestErrCode = 214501
	// ScaleUpRequestErrMsg -
	ScaleUpRequestErrMsg = "send scale up request to worker-manager error"

	// SpecificInstanceNotFound -
	SpecificInstanceNotFound = 150460
	// InstanceExceedConcurrency -
	InstanceExceedConcurrency = 150461

	LeaseIDIllegalCode  = 150462
	LeaseIDIllegalMsg   = "lease id is illegal"
	LeaseIDNotFoundCode = 150463
	LeaseIDNotFoundMsg  = "lease id is not found"

	// AcquireNonOwnerSchedulerErrorCode -
	AcquireNonOwnerSchedulerErrorCode = 150464
)

var (
	// ErrMap frontend code map to http code
	// Only return 200 to the management interface if the execution is successful
	ErrMap = map[int]int{
		// system error
		InnerResponseSuccessCode: http.StatusOK,
		InternalErrorCode:        http.StatusInternalServerError,
		// frontend error
		FrontendStatusOk:                    http.StatusOK,
		FrontendStatusAccepted:              http.StatusAccepted,
		FrontendStatusNoContent:             http.StatusNoContent,
		FrontendStatusBadRequest:            http.StatusBadRequest,
		FrontendStatusUnAuthorized:          http.StatusUnauthorized,
		FrontendStatusForbidden:             http.StatusForbidden,
		FrontendStatusNotFound:              http.StatusNotFound,
		FrontendStatusRequestEntityTooLarge: http.StatusRequestEntityTooLarge,
		FrontendStatusTooManyRequests:       http.StatusTooManyRequests,
		FrontendStatusInternalError:         http.StatusInternalServerError,
		FuncMetaNotFound:                    http.StatusInternalServerError,
		HeavyLoadCode:                       http.StatusInternalServerError,
		FrontendStatusTrafficLimitEffective: http.StatusInternalServerError,
		HTTPStreamNOTEnableError:            http.StatusInternalServerError,
		CreateStreamProducerError:           http.StatusInternalServerError,
		QueryStreamCustomerError:            http.StatusInternalServerError,
		SendDataToStreamError:               http.StatusInternalServerError,
		WriteResponseError:                  http.StatusInternalServerError,
		// frontend caas / multidata error
		// 500
		DsUploadFailed:          http.StatusInternalServerError,
		DsDownloadFailed:        http.StatusInternalServerError,
		DsDeleteFailed:          http.StatusInternalServerError,
		DsKeyNotFound:           http.StatusInternalServerError,
		UserFunctionInvokeError: http.StatusInternalServerError,
		// user error
		UserFuncEntryNotFoundErrCode:        http.StatusInternalServerError,
		UserFuncRunningExceptionErrCode:     http.StatusInternalServerError,
		UserFuncRspExceedLimitErrCode:       http.StatusInternalServerError,
		FrontendStatusMaxRequestBodySize:    http.StatusInternalServerError,
		FrontendStatusUnableSpecifyResource: http.StatusInternalServerError,
		UserFuncInvokeTimeout:               http.StatusInternalServerError,
		UserFuncInitFailCode:                http.StatusInternalServerError,
		UserFuncInitTimeoutCode:             http.StatusInternalServerError,
		StsConfigErrCode:                    http.StatusInternalServerError,
		// executor error
		ExecutorErrCodeInitFail: http.StatusInternalServerError,
	}
)

const (
	// VpcNoOperationalPermissions vpc has no operational permissions
	VpcNoOperationalPermissions = 4212
	// VPCNotFound error code of VPC not found
	VPCNotFound = 4219
	// VPCXRoleNotFound vcp xrole not func
	VPCXRoleNotFound = 4222
)

// vpc err comes from vpc controller
var (
	// ErrNoOperationalPermissionsVpc no operational permissions vpc
	ErrNoOperationalPermissionsVpc = errors.New("no operational permissions vpc, check the func xrole permissions")
	// ErrNoAvailableVpcPatInstance no available vpc pat instance
	ErrNoAvailableVpcPatInstance = errors.New("no available vpc pat instance")
	// ErrVPCNotFound VPC item not found error
	ErrVPCNotFound = errors.New("vpc item not found")
	// ErrVPCXRoleNotFound VPC xrole not found error
	ErrVPCXRoleNotFound = errors.New("can't find xrole")

	vpcErrorMap = map[string]int{
		ErrNoOperationalPermissionsVpc.Error(): VpcNoOperationalPermissions,
		ErrNoAvailableVpcPatInstance.Error():   NotEnoughNIC,
		ErrVPCNotFound.Error():                 VPCNotFound,
		ErrVPCXRoleNotFound.Error():            VPCXRoleNotFound,
	}

	vpcErrorCodeMsg = map[int]string{
		VpcNoOperationalPermissions: "no operational permissions vpc, check the func xrole permissions",
		NotEnoughNIC:                "not enough network cards",
		VPCNotFound:                 "VPC item not found",
		VPCXRoleNotFound:            "VPC can't find xrole",
	}
)

const (
	// InvalidState -
	InvalidState = 4040
	// InvalidStateErrMsg -
	InvalidStateErrMsg = "invalid state, expect not blank"
	// StateMismatch -
	StateMismatch = 4006
	// StateMismatchErrMsg -
	StateMismatchErrMsg = "invoke state id and function stateful flag are not matched"
	// StateExistedErrCode -
	StateExistedErrCode = 4027
	// StateExistedErrMsg -
	StateExistedErrMsg = "state cannot be created repeatedly"
	// StateNotExistedErrCode -
	StateNotExistedErrCode = 4026
	// StateNotExistedErrMsg -
	StateNotExistedErrMsg = "state not existed"
	// StateInstanceNotExistedErrCode -
	StateInstanceNotExistedErrCode = 4028
	// StateInstanceNotExistedErrMsg -
	StateInstanceNotExistedErrMsg = "state instance not existed"
	// StateInstanceNoLease -
	StateInstanceNoLease = 4025
	// StateInstanceNoLeaseMsg -
	StateInstanceNoLeaseMsg = "maximum number of leases reached"
	// FaaSSchedulerInternalErrCode -
	FaaSSchedulerInternalErrCode = 4029
	// FaaSSchedulerInternalErrMsg -
	FaaSSchedulerInternalErrMsg = "internal system error"
)

// worker error code
const (
	// WorkerInternalErrorCode code of unexpected error in worker
	WorkerInternalErrorCode = 161900
	// ReadingCodeTimeoutCode reading code package timed out
	ReadingCodeTimeoutCode = 161901
	// CallFunctionErrorCode code of calling other function error
	CallFunctionErrorCode = 161902
	// FuncInsExceptionCode function instance exception
	FuncInsExceptionCode = 161903
	// CheckSumErrorCode code of check sum error
	CheckSumErrorCode = 161904
	// DownLoadCodeErrorCode code of download code error
	DownLoadCodeErrorCode = 161905
	// RPCClientEmptyErrorCode code of when rpc client is nil
	RPCClientEmptyErrorCode = 161906
	// RuntimeManagerProcessExited runtime-manager process exited code
	RuntimeManagerProcessExited = 161907
	// WorkerPingVpcGatewayError code of worker ping vpc gateway error
	WorkerPingVpcGatewayError = 161908
	// UploadSnapshotErrorCode code of worker upload snapshot error
	UploadSnapshotErrorCode = 161909
	// RestoreDeadErrorCode code of restore is dead
	RestoreDeadErrorCode = 161910
	// ContentInconsistentErrorCode code of worker content inconsistent error
	ContentInconsistentErrorCode = 161911
	// CreateLimitErrorCode code of POSIX create limit error
	CreateLimitErrorCode = 161912
	// KernelEtcdWriteFailedCode code of core write etcd failed or circuit
	KernelEtcdWriteFailedCode = 161913
	// KernelResourceNotEnoughErrCode code of core resource not enough or schedule failure
	KernelResourceNotEnoughErrCode = 161914
	// WiseCloudNuwaColdStartErrCode code of use nuwa cold start failed
	WiseCloudNuwaColdStartErrCode = 161915
)

// Code trans frontend code to http code
func Code(frontendCode int) int {
	httpCode, exist := ErrMap[frontendCode]
	if !exist {
		return http.StatusInternalServerError
	}
	return httpCode
}

// Message trans frontend code to message
func Message(frontendCode int) string {
	httpCode, exist := ErrMap[frontendCode]
	if !exist {
		return ""
	}

	return fasthttp.StatusMessage(int(httpCode))
}

// VpcCode vpc controller err map to vpc err code
func VpcCode(errMsg string) int {
	if errCode, ok := vpcErrorMap[errMsg]; ok {
		return errCode
	}
	return 0
}

// VpcErMsg vpc err code map to err msg
func VpcErMsg(errCode int) string {
	if errMsg, ok := vpcErrorCodeMsg[errCode]; ok {
		return errMsg
	}
	return ""
}

var (
	errCodeRegCompile = regexp.MustCompile("code:[ 0-9]+,")
	errMsgRegCompile  = regexp.MustCompile("message:.+")
	codeRegCompile    = regexp.MustCompile("[0-9]+")
)

// GetKernelErrorCode will get kernel error code from error message
func GetKernelErrorCode(errMsg string) int {
	res := errCodeRegCompile.FindStringSubmatch(errMsg)
	if len(res) < 1 {
		return InternalErrorCode
	}
	res = codeRegCompile.FindStringSubmatch(errMsg)
	if len(res) != 1 {
		return InternalErrorCode
	}
	code, err := strconv.Atoi(res[0])
	if err != nil {
		return InternalErrorCode
	}
	return code
}

// GetKernelErrorMessage will get kernel error message from error message
func GetKernelErrorMessage(errMsg string) string {
	res := errMsgRegCompile.FindStringSubmatch(errMsg)
	if len(res) < 1 {
		return ""
	}
	trimRes := strings.TrimPrefix(res[0], "message: ")
	return trimRes
}
