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

// Package constants implements vars of all
package constants

const (
	// InsReqSuccessCode is the return code when instance request succeeds
	InsReqSuccessCode = 6030
	// UnsupportedOperationErrorCode is the return code when operation is not supported
	UnsupportedOperationErrorCode = 6031
	// FuncNotExistErrorCode is the return code when function does not exist
	FuncNotExistErrorCode = 6032
	// InsNotExistErrorCode is the return code when instance does not exist
	InsNotExistErrorCode = 6033
	// InsAcquireFailedErrorCode is the return code when acquire instance fails
	InsAcquireFailedErrorCode = 6034
	// LeaseExpireOrDeletedErrorCode is the return code when lease expires or be deleted
	LeaseExpireOrDeletedErrorCode = 6036
)
