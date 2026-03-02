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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	commontype "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionmanager/types"
)

// GenerateErrorResponse -
func GenerateErrorResponse(errorCode int, errorMessage string) *commontype.CallHandlerResponse {
	return &commontype.CallHandlerResponse{
		Code:    errorCode,
		Message: errorMessage,
	}
}

// GenerateSuccessResponse -
func GenerateSuccessResponse(code int, message string) *commontype.CallHandlerResponse {
	return &commontype.CallHandlerResponse{
		Code:    code,
		Message: message,
	}
}

// GetPatName by subnetId and securityGroup
func GetPatName(subnetId string, securityGroup []string) string {
	sort.Strings(securityGroup)
	hash := sha256.Sum256([]byte(strings.Join(securityGroup, "")))
	hashStr := hex.EncodeToString(hash[:])[:8] // hashStr len is 8
	return fmt.Sprintf("pat-%s-%s", subnetId, hashStr)
}

// UnstructuredToPat convert Unstructured to pat
func UnstructuredToPat(unstruct *unstructured.Unstructured) (*types.Pat, error) {
	var pat types.Pat
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.Object, &pat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert unstructured to custom resource: %v", err)
	}
	return &pat, nil
}

// PatToUnstructured convert pat to unstructured
func PatToUnstructured(customResource *types.Pat) (*unstructured.Unstructured, error) {
	unstructObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(customResource)
	if err != nil {
		return nil, fmt.Errorf("failed to convert custom resource to unstructured: %v", err)
	}
	return &unstructured.Unstructured{Object: unstructObj}, nil
}
