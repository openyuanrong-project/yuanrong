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

// Package tenantquota -
package tenantquota

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func Test_tenantCache_UpdateOrAddTenantQuota(t *testing.T) {
	type fields struct {
		tenantQuotaList sync.Map
	}
	type args struct {
		tenantID string
		tenant   types.TenantMetaInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "case_01 normal scenario",
			fields: fields{},
			args: args{
				tenantID: "test",
				tenant:   types.TenantMetaInfo{TenantInstanceMetaData: types.TenantInstanceMetaData{100, 100}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &tenantCache{
				tenantQuotaList: tt.fields.tenantQuotaList,
			}
			tc.UpdateOrAddTenantQuota(tt.args.tenantID, tt.args.tenant)
		})
	}
}

func Test_tenantCache_DeleteTenantQuota(t *testing.T) {
	tenantIns := &tenantQuota{
		maxOnDemandInstance: defaultMaxOnDemandInstanceNum,
		maxReversedInstance: defaultMaxReversedInstanceNum,
		mux:                 sync.RWMutex{},
	}
	type fields struct {
		tenantQuotaList sync.Map
	}
	type args struct {
		tenantID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "case_01 normal scenario",
			fields: fields{},
			args: args{
				tenantID: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &tenantCache{
				tenantQuotaList: tt.fields.tenantQuotaList,
			}
			tc.DeleteTenantQuota(tt.args.tenantID)
			tc.tenantQuotaList.Store(tt.args.tenantID, tenantIns)
			tc.DeleteTenantQuota(tt.args.tenantID)
			tc.tenantQuotaList.Store(tt.args.tenantID, "")
			tc.DeleteTenantQuota(tt.args.tenantID)
		})
	}
}

func Test_tenantCache_UpdateDefaultQuota2(t *testing.T) {
	var tenantQuotaList sync.Map
	tc := &tenantCache{
		tenantQuotaList: tenantQuotaList,
	}
	defaultTenantMetaInfo := types.TenantMetaInfo{TenantInstanceMetaData: types.TenantInstanceMetaData{1000, 1000}}

	tc.getTenantQuota("test")
	tc.UpdateDefaultQuota(defaultTenantMetaInfo)
	maxOnDemandInstance, maxReversedInstance := tc.GetTenantQuotaNum("test")
	assert.Equal(t, maxOnDemandInstance, int64(1000))
	assert.Equal(t, maxReversedInstance, int64(1000))

	tenantMetaInfo := types.TenantMetaInfo{TenantInstanceMetaData: types.TenantInstanceMetaData{100, 100}}
	tc.UpdateOrAddTenantQuota("test", tenantMetaInfo)
	maxOnDemandInstance, maxReversedInstance = tc.GetTenantQuotaNum("test")
	assert.Equal(t, maxOnDemandInstance, int64(100))
	assert.Equal(t, maxReversedInstance, int64(100))
}

func Test_tenantCache_UpdateDefaultQuota(t *testing.T) {
	type fields struct {
		tenantQuotaList sync.Map
	}
	type args struct {
		defaultTenant types.TenantMetaInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "case_01 normal scenario",
			fields: fields{},
			args: args{
				defaultTenant: types.TenantMetaInfo{TenantInstanceMetaData: types.TenantInstanceMetaData{1000, 1000}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &tenantCache{
				tenantQuotaList: tt.fields.tenantQuotaList,
			}
			tc.getTenantQuota("test")
			tc.UpdateDefaultQuota(tt.args.defaultTenant)
		})
	}
}
