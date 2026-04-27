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

// Package localauth authenticates requests by local configmaps
package localauth

import (
	"encoding/json"
	"os"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// GetDecryptFromEnv -
func GetDecryptFromEnv() (map[string]string, error) {
	res := make(map[string]string)
	value := os.Getenv("ENV_DELEGATE_DECRYPT")
	err := json.Unmarshal([]byte(value), &res)
	if err != nil {
		log.GetLogger().Warnf("ENV_DELEGATE_DECRYPT unmarshal error, it is null")
	}
	return res, nil
}
