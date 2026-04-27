/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package functioncr function cr common code
package functioncr

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CRD-related configurations are consistent with the AI cloud platform
const (
	yrGroup    = "yr.cap.io"
	yrVersion  = "v1"
	yrResource = "yrtasks"
	// CrdEventsQueue -
	CrdEventsQueue = "crdEventsQueue"
)

// GetCrdGVR -
func GetCrdGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    yrGroup,
		Version:  yrVersion,
		Resource: yrResource,
	}
}

// GetFunctionSpecData -
func GetFunctionSpecData(obj *unstructured.Unstructured) ([]byte, error) {
	specRaw, ok := obj.UnstructuredContent()["spec"]
	if !ok {
		return nil, errors.New("cr format err, no exists spec")
	}
	specBytes, err := json.Marshal(specRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cr spec, err: %s", err.Error())
	}
	return specBytes, nil
}
