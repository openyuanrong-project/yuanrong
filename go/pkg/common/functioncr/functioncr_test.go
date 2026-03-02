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

package functioncr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// createTestCR 创建一个测试用的自定义资源
func createTestCR(name, functionUrn string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "yr.cap.io/v1",
			"kind":       "YRTask",
			"metadata": map[string]interface{}{
				"name":              name,
				"namespace":         "default",
				"uid":               "1234-5678-90ab-cdef-" + name,
				"resourceVersion":   "1",
				"creationTimestamp": metav1.Now().Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"funcMetaData": map[string]interface{}{
					"functionURN": functionUrn,
				},
				"extendedMetaData": map[string]interface{}{
					"customContainerConfig": map[string]interface{}{
						"image": "nginx:latest",
					},
				},
			},
		},
	}
}

func TestGetFunctionSpecData(t *testing.T) {
	t.Run("no exists spec", func(t *testing.T) {
		noSpecObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "yr.cap.io/v1",
				"kind":       "YRTask",
				"metadata": map[string]interface{}{
					"name":              "test",
					"namespace":         "default",
					"uid":               "1234-5678-90ab-cdef-test",
					"resourceVersion":   "1",
					"creationTimestamp": metav1.Now().Format(time.RFC3339),
				},
			},
		}
		_, err := GetFunctionSpecData(noSpecObj)
		assert.EqualError(t, err, "cr format err, no exists spec")
	})
	t.Run("success", func(t *testing.T) {
		_, err := GetFunctionSpecData(createTestCR("function-cr", "funcurn"))
		assert.NoError(t, err)
	})
}
