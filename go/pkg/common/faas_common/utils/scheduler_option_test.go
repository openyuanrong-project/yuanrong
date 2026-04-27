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

package utils

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestCreateCustomExtensions(t *testing.T) {
	Convey("Test CreateCustomExtensions", t, func() {
		got := CreateCustomExtensions(nil, MonopolyPolicyValue)
		So(got, ShouldNotBeNil)
	})
}

func TestCreateCreateOptions(t *testing.T) {
	Convey("Test CreateSchedulingOptions", t, func() {
		expectValue := "test"
		createOptions := make(map[string]string, 20)
		got := CreateCreateOptions(createOptions, "test", "test")
		So(got["test"], ShouldEqual, expectValue)
	})
}

func TestGenerateResourcesMap(t *testing.T) {
	Convey("Test GenerateResourcesMap", t, func() {
		res := GenerateResourcesMap(300, 128)
		So(res, ShouldResemble, map[string]float64{
			scheduleCPU:    300,
			scheduleMemory: 128,
		})
	})
}

func TestCreatePodAffinity(t *testing.T) {
	type args struct {
		key          string
		label        string
		affinityType api.AffinityType
	}
	tests := []struct {
		name string
		args args
		want []api.Affinity
	}{
		{"case1", args{
			key:          "faasfrontend",
			label:        "faasfrontend",
			affinityType: api.PreferredAntiAffinity,
		}, []api.Affinity{
			api.Affinity{
				Kind:                     api.AffinityKindInstance,
				Affinity:                 api.PreferredAntiAffinity,
				PreferredPriority:        false,
				PreferredAntiOtherLabels: false,
				LabelOps: []api.LabelOperator{{
					Type:        api.LabelOpIn,
					LabelKey:    "faasfrontend",
					LabelValues: []string{"faasfrontend"},
				},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreatePodAffinity(tt.args.key, tt.args.label, tt.args.affinityType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePodAffinity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddNodeSelector(t *testing.T) {
	type args struct {
		nodeSelectorMap map[string]string
		extraParams     *types.ExtraParams
	}
	tests := []struct {
		name string
		args args
	}{
		{"case1", args{
			nodeSelectorMap: map[string]string{"k": "v"},
			extraParams:     &types.ExtraParams{CustomExtensions: make(map[string]string)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddNodeSelector(tt.args.nodeSelectorMap, tt.args.extraParams)
		})
	}
}
