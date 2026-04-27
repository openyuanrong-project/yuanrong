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

package types

import (
	"fmt"
	"regexp"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
)

const (
	defaultServeAppRuntime       = "python3.9"
	defaultServeAppTimeout       = 900
	defaultServeAppCpu           = 1000
	defaultServeAppMemory        = 1024
	defaultServeAppConcurrentNum = 1000
)

// ServeDeploySchema -
type ServeDeploySchema struct {
	Applications []ServeApplicationSchema `json:"applications"`
}

// ServeApplicationSchema -
type ServeApplicationSchema struct {
	Name        string                  `json:"name"`
	RoutePrefix string                  `json:"route_prefix"`
	ImportPath  string                  `json:"import_path"`
	RuntimeEnv  ServeRuntimeEnvSchema   `json:"runtime_env"`
	Deployments []ServeDeploymentSchema `json:"deployments"`
}

// ServeDeploymentSchema -
type ServeDeploymentSchema struct {
	Name                string `json:"name"`
	NumReplicas         int64  `json:"num_replicas"`
	HealthCheckPeriodS  int64  `json:"health_check_period_s"`
	HealthCheckTimeoutS int64  `json:"health_check_timeout_s"`
}

// ServeRuntimeEnvSchema -
type ServeRuntimeEnvSchema struct {
	Pip        []string       `json:"pip"`
	WorkingDir string         `json:"working_dir"`
	EnvVars    map[string]any `json:"env_vars"`
}

// ServeFuncWithKeysAndFunctionMetaInfo -
type ServeFuncWithKeysAndFunctionMetaInfo struct {
	FuncMetaKey     string
	InstanceMetaKey string
	FuncMetaInfo    *FunctionMetaInfo
}

// Validate serve deploy schema by set of rules
func (s *ServeDeploySchema) Validate() error {
	// 1. app name unique
	appNameSet := make(map[string]struct{})
	for _, app := range s.Applications {
		if _, ok := appNameSet[app.Name]; ok {
			return fmt.Errorf("duplicated application name: %s", app.Name)
		}
		appNameSet[app.Name] = struct{}{}
	}
	// 2. app routes unique
	appRouteSet := make(map[string]struct{})
	for _, app := range s.Applications {
		if _, ok := appRouteSet[app.RoutePrefix]; ok {
			return fmt.Errorf("duplicated application route prefix: %s", app.RoutePrefix)
		}
		appRouteSet[app.RoutePrefix] = struct{}{}
	}
	// 3. app name non empty
	for _, app := range s.Applications {
		if app.Name == "" {
			return fmt.Errorf("application names must be nonempty")
		}
	}
	return nil
}

// ToFaaSFuncMetas -
func (s *ServeDeploySchema) ToFaaSFuncMetas() []*ServeFuncWithKeysAndFunctionMetaInfo {
	var allMetas []*ServeFuncWithKeysAndFunctionMetaInfo
	for _, a := range s.Applications {
		// we don't really check it there are some repeated part? and just assume translate won't fail
		for _, deploymentFuncMeta := range a.ToFaaSFuncMetas() {
			allMetas = append(allMetas, deploymentFuncMeta)
		}
	}
	return allMetas
}

// ToFaaSFuncMetas -
func (s *ServeApplicationSchema) ToFaaSFuncMetas() []*ServeFuncWithKeysAndFunctionMetaInfo {
	var allMetas []*ServeFuncWithKeysAndFunctionMetaInfo
	for _, d := range s.Deployments {
		meta := d.ToFaaSFuncMeta(s)
		allMetas = append(allMetas, meta)
	}
	return allMetas
}

// ToFaaSFuncMeta -
func (s *ServeDeploymentSchema) ToFaaSFuncMeta(
	belongedApp *ServeApplicationSchema) *ServeFuncWithKeysAndFunctionMetaInfo {
	faasFuncUrn := NewServeFunctionKeyWithDefault()
	faasFuncUrn.AppName = belongedApp.Name
	faasFuncUrn.DeploymentName = s.Name

	// make a copied app to make it contains only this deployment info
	copiedApp := *belongedApp
	copiedApp.Deployments = []ServeDeploymentSchema{*s}

	return &ServeFuncWithKeysAndFunctionMetaInfo{
		FuncMetaKey:     faasFuncUrn.ToFuncMetaKey(),
		InstanceMetaKey: faasFuncUrn.ToInstancesMetaKey(),
		FuncMetaInfo: &FunctionMetaInfo{
			FuncMetaData: FuncMetaData{
				Name:               faasFuncUrn.DeploymentName,
				Runtime:            defaultServeAppRuntime,
				Timeout:            defaultServeAppTimeout,
				Version:            faasFuncUrn.Version,
				FunctionURN:        faasFuncUrn.ToFaasFunctionUrn(),
				TenantID:           faasFuncUrn.TenantID,
				FunctionVersionURN: faasFuncUrn.ToFaasFunctionVersionUrn(),
				FuncName:           faasFuncUrn.DeploymentName,
				BusinessType:       constant.BusinessTypeServe,
			},
			ResourceMetaData: ResourceMetaData{
				CPU:    defaultServeAppCpu,
				Memory: defaultServeAppMemory,
			},
			InstanceMetaData: InstanceMetaData{
				MaxInstance:   s.NumReplicas,
				MinInstance:   s.NumReplicas,
				ConcurrentNum: defaultServeAppConcurrentNum,
				IdleMode:      false,
			},
			ExtendedMetaData: ExtendedMetaData{
				ServeDeploySchema: ServeDeploySchema{
					Applications: []ServeApplicationSchema{
						copiedApp,
					},
				},
			},
		},
	}
}

const (
	defaultTenantID    = "default"
	defaultFuncVersion = "latest"

	faasMetaKey              = constant.MetaFuncKey
	instanceMetaKey          = "/instances/business/yrk/cluster/cluster001/tenant/%s/function/%s/version/%s"
	faasFuncURN6tuplePattern = "sn:cn:yrk:%s:function:%s"
	faasFuncURN7tuplePattern = "sn:cn:yrk:%s:function:%s:%s"
)

// ServeFunctionKey is a faas urn with necessary parts
type ServeFunctionKey struct {
	TenantID       string
	AppName        string
	DeploymentName string
	Version        string
}

// NewServeFunctionKeyWithDefault returns a struct with default values
func NewServeFunctionKeyWithDefault() *ServeFunctionKey {
	return &ServeFunctionKey{
		TenantID: defaultTenantID,
		Version:  defaultFuncVersion,
	}
}

// ToFuncNameTriplet - 0@svc@func
func (f *ServeFunctionKey) ToFuncNameTriplet() string {
	return fmt.Sprintf("0@%s@%s", f.AppName, f.DeploymentName)
}

// ToFuncMetaKey - /sn/functions/business/yrk/tenant/default/function/0@svc@func/version/latest
func (f *ServeFunctionKey) ToFuncMetaKey() string {
	return fmt.Sprintf(faasMetaKey, f.TenantID, f.ToFuncNameTriplet(), f.Version)
}

// ToInstancesMetaKey - /instances/business/yrk/cluster/cluster001/tenant/125...346/function/0@svc@func/version/latest
func (f *ServeFunctionKey) ToInstancesMetaKey() string {
	return fmt.Sprintf(instanceMetaKey, f.TenantID, f.ToFuncNameTriplet(), f.Version)
}

// ToFaasFunctionUrn - sn:cn:yrk:default:function:0@service@function
func (f *ServeFunctionKey) ToFaasFunctionUrn() string {
	return fmt.Sprintf(faasFuncURN6tuplePattern, f.TenantID, f.ToFuncNameTriplet())
}

// ToFaasFunctionVersionUrn - sn:cn:yrk:default:function:0@svc@func:latest
func (f *ServeFunctionKey) ToFaasFunctionVersionUrn() string {
	return fmt.Sprintf(faasFuncURN7tuplePattern, f.TenantID, f.ToFuncNameTriplet(), f.Version)
}

// FromFaasFunctionKey - default/0@svc@func/latest
func (f *ServeFunctionKey) FromFaasFunctionKey(funcKey string) error {
	const (
		serveFaasFuncKeyMatchesIdxTenantID = iota + 1
		serveFaasFuncKeyMatchesIdxAppName
		serveFaasFuncKeyMatchesIdxDeploymentName
		serveFaasFuncKeyMatchesIdxVersion
		serveFaasFuncKeyMatchesIdxMax
	)
	re := regexp.MustCompile(`^([a-zA-Z0-9]*)/.*@([^@]+)@([^/]+)/(.*)$`)
	matches := re.FindStringSubmatch(funcKey)
	if len(matches) < serveFaasFuncKeyMatchesIdxMax {
		return fmt.Errorf("extract failed from %s", funcKey)
	}
	f.TenantID = matches[serveFaasFuncKeyMatchesIdxTenantID]
	f.AppName = matches[serveFaasFuncKeyMatchesIdxAppName]
	f.DeploymentName = matches[serveFaasFuncKeyMatchesIdxDeploymentName]
	f.Version = matches[serveFaasFuncKeyMatchesIdxVersion]
	return nil
}
