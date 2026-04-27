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

// Package types -
package types

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestServeFunctionKeyTrans(t *testing.T) {
	k := NewServeFunctionKeyWithDefault()
	k.AppName = "svc"
	k.DeploymentName = "func"
	convey.Convey("Given a serve function key", t, func() {
		convey.Convey("When trans to a func name triplet", func() {
			convey.So(k.ToFuncNameTriplet(), convey.ShouldEqual, "0@svc@func")
		})
		convey.Convey("When trans to a func meta key", func() {
			convey.So(k.ToFuncMetaKey(), convey.ShouldEqual,
				"/sn/functions/business/yrk/tenant/default/function/0@svc@func/version/latest")
		})
		convey.Convey("When trans to a instance meta key", func() {
			convey.So(k.ToInstancesMetaKey(), convey.ShouldEqual,
				"/instances/business/yrk/cluster/cluster001/tenant/default/function/0@svc@func/version/latest")
		})
		convey.Convey("When trans to a FaasFunctionUrn", func() {
			convey.So(k.ToFaasFunctionUrn(), convey.ShouldEqual,
				"sn:cn:yrk:default:function:0@svc@func")
		})
		convey.Convey("When trans ToFaasFunctionVersionUrn", func() {
			convey.So(k.ToFaasFunctionVersionUrn(), convey.ShouldEqual,
				"sn:cn:yrk:default:function:0@svc@func:latest")
		})
	})
}

func TestServeDeploySchema_ToFaaSFuncMetas(t *testing.T) {
	convey.Convey("Test ServeDeploySchema ToFaaSFuncMetas", t, func() {
		// Setup mock data
		app1 := ServeApplicationSchema{
			Name:        "app1",
			RoutePrefix: "/app1",
			ImportPath:  "path1",
			RuntimeEnv: ServeRuntimeEnvSchema{
				Pip:        []string{"package1", "package2"},
				WorkingDir: "/app1",
				EnvVars:    map[string]any{"key1": "value1"},
			},
			Deployments: []ServeDeploymentSchema{
				{
					Name:                "deployment1",
					NumReplicas:         2,
					HealthCheckPeriodS:  30,
					HealthCheckTimeoutS: 10,
				},
			},
		}

		serveDeploy := ServeDeploySchema{
			Applications: []ServeApplicationSchema{app1},
		}

		convey.Convey("It should return correct faas function metas", func() {
			result := serveDeploy.ToFaaSFuncMetas()
			convey.So(len(result), convey.ShouldBeGreaterThan, 0)
			convey.So(result[0].FuncMetaKey, convey.ShouldNotBeEmpty)
		})
	})
}

func TestServeFunctionKey(t *testing.T) {
	convey.Convey("Test FromFaasFunctionKey", t, func() {
		convey.Convey("It should return correct faas function metas", func() {
			key := "default/0@svc@func/latest"
			sfk := ServeFunctionKey{}
			err := sfk.FromFaasFunctionKey(key)
			convey.So(err, convey.ShouldBeNil)
			convey.So(sfk.Version, convey.ShouldEqual, "latest")
			convey.So(sfk.AppName, convey.ShouldEqual, "svc")
			convey.So(sfk.DeploymentName, convey.ShouldEqual, "func")
			convey.So(sfk.TenantID, convey.ShouldEqual, "default")
		})
		convey.Convey("It should return incorrect faas function metas", func() {
			key := "default/0@svc@func"
			sfk := ServeFunctionKey{}
			err := sfk.FromFaasFunctionKey(key)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})

	convey.Convey("Test FaasKey Test", t, func() {
		convey.Convey("test default faas key", func() {
			sfk := NewServeFunctionKeyWithDefault()
			convey.So(sfk.TenantID, convey.ShouldEqual, defaultTenantID)
			convey.So(sfk.Version, convey.ShouldEqual, defaultFuncVersion)
		})

		convey.Convey("test convert", func() {
			sfk := NewServeFunctionKeyWithDefault()
			sfk.AppName = "svc"
			sfk.DeploymentName = "func"

			convey.So(sfk.ToFuncNameTriplet(),
				convey.ShouldEqual,
				"0@svc@func")
			convey.So(sfk.ToFuncMetaKey(),
				convey.ShouldEqual,
				"/sn/functions/business/yrk/tenant/default/function/0@svc@func/version/latest")
			convey.So(sfk.ToInstancesMetaKey(),
				convey.ShouldEqual,
				"/instances/business/yrk/cluster/cluster001/tenant/default/function/0@svc@func/version/latest")
			convey.So(sfk.ToFaasFunctionUrn(),
				convey.ShouldEqual,
				"sn:cn:yrk:default:function:0@svc@func")
			convey.So(sfk.ToFaasFunctionVersionUrn(),
				convey.ShouldEqual,
				"sn:cn:yrk:default:function:0@svc@func:latest")
		})

		convey.Convey("It should return incorrect faas function metas", func() {
			sfk := NewServeFunctionKeyWithDefault()
			convey.So(sfk.TenantID, convey.ShouldEqual, defaultTenantID)
			convey.So(sfk.Version, convey.ShouldEqual, defaultFuncVersion)
		})
	})
}

func TestServeDeploySchemaValidate(t *testing.T) {
	convey.Convey("Test Validate", t, func() {
		sds := ServeDeploySchema{
			Applications: []ServeApplicationSchema{
				{
					Name:        "app1",
					RoutePrefix: "/app1",
					ImportPath:  "path1",
					RuntimeEnv: ServeRuntimeEnvSchema{
						Pip:        []string{"package1", "package2"},
						WorkingDir: "/app1",
						EnvVars:    map[string]any{"key1": "value1"},
					},
					Deployments: []ServeDeploymentSchema{
						{
							Name:                "deployment1",
							NumReplicas:         2,
							HealthCheckPeriodS:  30,
							HealthCheckTimeoutS: 10,
						},
					},
				},
			}}
		convey.Convey("on repeated app name", func() {
			sdsOther := sds
			app0 := sdsOther.Applications[0]
			sdsOther.Applications = append(sdsOther.Applications, app0)

			err := sdsOther.Validate()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("on repeated route prefix", func() {
			sdsOther := sds
			app0 := sdsOther.Applications[0]
			app0.Name = "othername"
			sdsOther.Applications = append(sdsOther.Applications, app0)

			err := sdsOther.Validate()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("on empty app name", func() {
			sdsOther := sds
			app0 := sdsOther.Applications[0]
			app0.Name = ""
			app0.RoutePrefix = "/other"
			sdsOther.Applications = append(sdsOther.Applications, app0)

			err := sdsOther.Validate()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("ok", func() {
			err := sds.Validate()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
