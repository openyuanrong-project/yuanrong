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

// Package etcdkey contains etcd key definition and tools
package etcdkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateWorkersKey_ParseFrom(t *testing.T) {
	type fields struct {
		KeyType    string
		BusinessID string
		TenantID   string
		Function   string
		Version    string
		Zone       string
		StateID    string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test001",
			fields: fields{
				KeyType:    "stateworkers",
				BusinessID: "yrk",
				TenantID:   "tenantID",
				Function:   "function",
				Version:    "$latest",
				Zone:       "defaultaz",
				StateID:    "stateID",
			},
			args: args{
				key: "/sn/stateworkers/business/yrk/tenant/tenantID" +
					"/function/function/version/$latest/defaultaz/stateID",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StateWorkersKey{
				TypeKey:    tt.fields.KeyType,
				BusinessID: tt.fields.BusinessID,
				TenantID:   tt.fields.TenantID,
				Function:   tt.fields.Function,
				Version:    tt.fields.Version,
				Zone:       tt.fields.Zone,
				StateID:    tt.fields.StateID,
			}
			if err := s.ParseFrom(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("ParseFrom() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStateWorkersKey_String(t *testing.T) {
	type fields struct {
		KeyType    string
		BusinessID string
		TenantID   string
		Function   string
		Version    string
		Zone       string
		StateID    string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test002",
			fields: fields{
				KeyType:    "stateworkers",
				BusinessID: "yrk",
				TenantID:   "tenantID",
				Function:   "function",
				Version:    "$latest",
				Zone:       "defaultaz",
				StateID:    "stateID",
			},
			want: "/sn/stateworkers/business/yrk/tenant/tenantID" +
				"/function/function/version/$latest/defaultaz/stateID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StateWorkersKey{
				TypeKey:    tt.fields.KeyType,
				BusinessID: tt.fields.BusinessID,
				TenantID:   tt.fields.TenantID,
				Function:   tt.fields.Function,
				Version:    tt.fields.Version,
				Zone:       tt.fields.Zone,
				StateID:    tt.fields.StateID,
			}
			if got := s.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkerInstanceKey_ParseFrom(t *testing.T) {
	tests := []struct {
		arg  string
		want WorkerInstanceKey
		err  bool
	}{
		{
			arg: "/sn/workers/business/busid/tenant/abcd/function/" +
				"0-counter-addone/version/$latest/defaultaz/defaultaz-#-pool7-500-500-python3.7-58f588848d-smss8",
			want: WorkerInstanceKey{
				TypeKey:    "workers",
				BusinessID: "busid",
				TenantID:   "abcd",
				Function:   "0-counter-addone",
				Version:    "$latest",
				Zone:       "defaultaz",
				Instance:   "defaultaz-#-pool7-500-500-python3.7-58f588848d-smss8",
			},
			err: false,
		},
		{
			arg: "/sn/workers/business/busid/tenant/abcd/function/" +
				"0-counter-addone/version/$latest/defaultaz",
			want: WorkerInstanceKey{},
			err:  true,
		},
		{
			arg: "/sn/workers/business/yrk/tenant/0/function/function-task/version/$latest/defaultaz/dggphis36581",
			want: WorkerInstanceKey{
				TypeKey:    "workers",
				BusinessID: "yrk",
				TenantID:   "0",
				Function:   "function-task",
				Version:    "$latest",
				Zone:       "defaultaz",
				Instance:   "dggphis36581",
			},
			err: false,
		},
	}
	for _, tt := range tests {
		worker := WorkerInstanceKey{}
		err := worker.ParseFrom(tt.arg)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, worker)
		}
	}
}

func TestWorkerInstanceKey_String(t *testing.T) {
	tests := []struct {
		arg  WorkerInstanceKey
		want string
	}{
		{
			want: "/sn/workers/business/busid/tenant/abcd/function/" +
				"0-counter-addone/version/$latest/defaultaz/defaultaz-#-pool7-500-500-python3.7-58f588848d-smss8",
			arg: WorkerInstanceKey{
				TypeKey:    "workers",
				BusinessID: "busid",
				TenantID:   "abcd",
				Function:   "0-counter-addone",
				Version:    "$latest",
				Zone:       "defaultaz",
				Instance:   "defaultaz-#-pool7-500-500-python3.7-58f588848d-smss8",
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.arg.String())
	}
}

func TestAnonymizeTenantCommonEtcdKey(t *testing.T) {
	keyA := "/yr/functions/business/yrk"
	anonymizeKeyA := AnonymizeTenantCommonEtcdKey(keyA)
	assert.Equal(t, keyA, anonymizeKeyA)
	keyB := "/yr/functions/business/yrk/tenant/8e08d5cc0ad34032bba8d636040a278c/function/0-test1-addone/version/$latest"
	anonymizeKeyB := AnonymizeTenantCommonEtcdKey(keyB)
	assert.NotEqual(t, keyB, anonymizeKeyB)
}
func TestAnonymizeTenantCronTriggerEtcdKey(t *testing.T) {
	keyA := "/sn/triggers/triggerType/CRON/business/yrk/tenant"
	anonymizeKeyA := AnonymizeTenantCronTriggerEtcdKey(keyA)
	assert.Equal(t, keyA, anonymizeKeyA)
	keyB := "/sn/triggers/triggerType/CRON/business/yrk/tenant/i1fe539427b2/function/pytzip/version/$latest/398e2ca2"
	anonymizeKeyB := AnonymizeTenantCronTriggerEtcdKey(keyB)
	assert.NotEqual(t, keyB, anonymizeKeyB)
}
func TestAnonymizeTenantFunctionMetadataEtcdKey(t *testing.T) {
	keyA := "/repo/FunctionVersion/business"
	anonymizeKeyA := AnonymizeTenantFunctionMetadataEtcdKey(keyA)
	assert.Equal(t, keyA, anonymizeKeyA)
	keyB := "/repo/FunctionVersion/business/tenant/funcName/version/"
	anonymizeKeyB := AnonymizeTenantFunctionMetadataEtcdKey(keyB)
	assert.NotEqual(t, keyB, anonymizeKeyB)
}

func TestFunctionInstanceKey_ParseFrom(t *testing.T) {
	tests := []struct {
		arg  string
		want FunctionInstanceKey
		err  bool
	}{
		{
			arg: "/sn/instance/business/yrk/tenant/tenantID/defaultaz/instanceID",
			want: FunctionInstanceKey{
				TypeKey:    "instance",
				BusinessID: "yrk",
				TenantID:   "tenantID",
				InstanceID: "instanceID",
				Version:    "",
				Zone:       "defaultaz",
			},
		},
		{
			arg:  "/sn/instance/business/b1/tenant/t1/function/0-s1-test/version/1/defaultaz",
			want: FunctionInstanceKey{},
		},
	}
	for _, tt := range tests {
		instance := FunctionInstanceKey{}
		err := instance.ParseFrom(tt.arg)
		assert.Error(t, err)
	}
}
