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

// Package models serve_api_models Refers to ray-operator/controllers/ray/utils/serve_api_models.go
package models

// ServeGetResponse is the ServeDetails in kube-ray
type ServeGetResponse struct {
	ServeDetails
}

// ServeDeploymentStatus -
type ServeDeploymentStatus struct {
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// ServeApplicationStatus -
type ServeApplicationStatus struct {
	Deployments map[string]ServeDeploymentStatus `json:"deployments"`
	Name        string                           `json:"name,omitempty"`
	Status      string                           `json:"status"`
	Message     string                           `json:"message,omitempty"`
}

// ServeDeploymentDetails -
type ServeDeploymentDetails struct {
	ServeDeploymentStatus
	RoutePrefix string `json:"route_prefix,omitempty"`
}

// ServeApplicationDetails -
type ServeApplicationDetails struct {
	Deployments map[string]ServeDeploymentDetails `json:"deployments"`
	ServeApplicationStatus
	RoutePrefix string `json:"route_prefix,omitempty"`
	DocsPath    string `json:"docs_path,omitempty"`
}

// ServeDetails -
type ServeDetails struct {
	Applications map[string]*ServeApplicationDetails `json:"applications"`
	DeployMode   string                              `json:"deploy_mode,omitempty"`
}
