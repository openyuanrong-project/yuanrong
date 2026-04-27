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
	"context"
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

// PatchSlice -
type PatchSlice []*gomonkey.Patches

// PatchesFunc -
type PatchesFunc func() PatchSlice

// InitPatchSlice -
func InitPatchSlice() PatchSlice {
	return make([]*gomonkey.Patches, 0)
}

// Append -
func (p *PatchSlice) Append(patches PatchSlice) {
	if len(patches) > 0 {
		*p = append(*p, patches...)
	}
}

// ResetAll -
func (p PatchSlice) ResetAll() {
	for _, item := range p {
		item.Reset()
	}
}

// FakeLogger -
type FakeLogger struct{}

// With -
func (f *FakeLogger) With(fields ...zapcore.Field) api.FormatLogger {
	return f
}

// Infof -
func (f *FakeLogger) Infof(format string, paras ...interface{}) {}

// Errorf -
func (f *FakeLogger) Errorf(format string, paras ...interface{}) {}

// Warnf -
func (f *FakeLogger) Warnf(format string, paras ...interface{}) {}

// Debugf -
func (f *FakeLogger) Debugf(format string, paras ...interface{}) {}

// Fatalf -
func (f *FakeLogger) Fatalf(format string, paras ...interface{}) {}

// Info -
func (f *FakeLogger) Info(msg string, fields ...zap.Field) {}

// Error -
func (f *FakeLogger) Error(msg string, fields ...zap.Field) {}

// Warn -
func (f *FakeLogger) Warn(msg string, fields ...zap.Field) {}

// Debug -
func (f *FakeLogger) Debug(msg string, fields ...zap.Field) {}

// Fatal -
func (f *FakeLogger) Fatal(msg string, fields ...zap.Field) {}

// Sync -
func (f *FakeLogger) Sync() {}

// FakeEtcdLease -
type FakeEtcdLease struct {
}

// Grant -
func (m FakeEtcdLease) Grant(_ context.Context, _ int64) (*clientv3.LeaseGrantResponse, error) {
	return &clientv3.LeaseGrantResponse{ID: 1}, nil
}

// Revoke -
func (m FakeEtcdLease) Revoke(_ context.Context, _ clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	return nil, nil
}

// TimeToLive -
func (m FakeEtcdLease) TimeToLive(_ context.Context, _ clientv3.LeaseID,
	_ ...clientv3.LeaseOption) (*clientv3.LeaseTimeToLiveResponse, error) {
	return nil, nil
}

// Leases -
func (m FakeEtcdLease) Leases(_ context.Context) (*clientv3.LeaseLeasesResponse, error) {
	return nil, nil
}

// KeepAlive -
func (m FakeEtcdLease) KeepAlive(_ context.Context, _ clientv3.LeaseID) (
	<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return nil, nil
}

// KeepAliveOnce -
func (m FakeEtcdLease) KeepAliveOnce(_ context.Context, _ clientv3.LeaseID) (
	*clientv3.LeaseKeepAliveResponse, error) {
	return nil, nil
}

// Close -
func (m FakeEtcdLease) Close() error {
	return errors.New("close error")
}
