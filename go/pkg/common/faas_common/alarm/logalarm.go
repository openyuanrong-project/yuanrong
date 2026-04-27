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

// Package alarm alarm log by filebeat
package alarm

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/config"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
)

const (
	// ConfigKey environment variable key of alarm config
	ConfigKey = "ALARM_CONFIG"

	cacheLimit = 10 * 1 << 20 // 10 mb

	// Level3 -
	Level3 = "critical"
	// Level2 -
	Level2 = "major"
	// Level1 -
	Level1 = "minor"
	// Level0 -
	Level0 = "notice"

	// GenerateAlarmLog -
	GenerateAlarmLog = "firing"
	// ClearAlarmLog -
	ClearAlarmLog = "resolved"

	// InsufficientMinInstance00001 alarm id
	InsufficientMinInstance00001 = "InsufficientMinInstance00001"
	// MetadataEtcdConnection00001 alarm id
	MetadataEtcdConnection00001 = "MetadataEtcdConnection00001"
	// RouterEtcdConnection00001 alarm id
	RouterEtcdConnection00001 = "RouterEtcdConnection00001"
	// InitStsSdkErr00001 alarm id
	InitStsSdkErr00001 = "InitStsSdkErr00001"
	// PullStsConfiguration00001 alarm id
	PullStsConfiguration00001 = "PullStsConfiguration00001"
	// ReportToXPUManageFailed00001 alarm id
	ReportToXPUManageFailed00001 = "ReportToXPUManageFailed00001"
	// FaaSSchedulerRemovedFromHashRing00001 alarm id
	FaaSSchedulerRemovedFromHashRing00001 = "FaaSSchedulerRemovedFromHashRing00001"
	// FaaSFrontendReceiptDMQMessage00001 -
	FaaSFrontendReceiptDMQMessage00001 = "FaaSFrontendReceiptDMQMessage00001"
	// FaaSFrontendDequeueDMQMessage00001 -
	FaaSFrontendDequeueDMQMessage00001 = "FaaSFrontendDequeueDMQMessage00001"
	// NoAvailableSchedulerInstance00001 没有可用的scheduler实例的告警id
	NoAvailableSchedulerInstance00001 = "NoAvailableSchedulerInstance00001"
)

var (
	alarmLogger      *zap.Logger
	createLoggerErr  error
	createLoggerOnce sync.Once
)

// LogAlarmInfo Custom alarm info
type LogAlarmInfo struct {
	AlarmID    string
	AlarmName  string
	AlarmLevel string
}

// Detail alarm detail
type Detail struct {
	SourceTag      string // 告警来源
	OpType         string // 告警操作类型
	Details        string // 告警详情
	StartTimestamp int    // 产生时间
	EndTimestamp   int    // 清除时间
}

// GetAlarmLogger -
func GetAlarmLogger() (*zap.Logger, error) {
	createLoggerOnce.Do(func() {
		alarmLogger, createLoggerErr = newAlarmLogger()
		if createLoggerErr != nil {
			return
		}
		if alarmLogger == nil {
			createLoggerErr = errors.New("failed to new alarmLogger")
			return
		}
		// 祥云四元组 - 站点/租户ID/产品ID/服务ID
		alarmLogger = alarmLogger.With(zapcore.Field{
			Key: "site", Type: zapcore.StringType,
			String: os.Getenv(constant.WiseCloudSite),
		}, zapcore.Field{
			Key: "tenant_id", Type: zapcore.StringType,
			String: os.Getenv(constant.TenantID),
		}, zapcore.Field{
			Key: "application_id", Type: zapcore.StringType,
			String: os.Getenv(constant.ApplicationID),
		}, zapcore.Field{
			Key: "service_id", Type: zapcore.StringType,
			String: os.Getenv(constant.ServiceID),
		})
	})
	return alarmLogger, createLoggerErr
}

func newAlarmLogger() (*zap.Logger, error) {
	coreInfo, err := config.ExtractCoreInfoFromEnv(ConfigKey)
	log.GetLogger().Infof("ALARM_CONFIG is: %v", coreInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to valid log path, err: %s", err.Error())
		return nil, err
	}

	coreInfo.FilePath = filepath.Join(coreInfo.FilePath, "alarm.dat")

	sink, err := logger.CreateSink(coreInfo)
	if err != nil {
		log.GetLogger().Errorf("failed to create sink: %s", err.Error())
		return nil, err
	}

	ws := zapcore.AddSync(sink)
	priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})
	encoderConfig := zapcore.EncoderConfig{}
	rollingFileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	return zap.New(zapcore.NewCore(rollingFileEncoder, ws, priority)), nil
}

func addAlarmLogger(rollingLogger *zap.Logger, alarmInfo *LogAlarmInfo, detail *Detail) *zap.Logger {
	return rollingLogger.With(zapcore.Field{
		Key: "id", Type: zapcore.StringType,
		String: alarmInfo.AlarmID,
	}, zapcore.Field{
		Key: "name", Type: zapcore.StringType,
		String: alarmInfo.AlarmName,
	}, zapcore.Field{
		Key: "level", Type: zapcore.StringType,
		String: alarmInfo.AlarmLevel,
	}, zapcore.Field{
		Key: "source_tag", Type: zapcore.StringType,
		String: detail.SourceTag,
	}, zapcore.Field{
		Key: "op_type", Type: zapcore.StringType,
		String: detail.OpType,
	}, zapcore.Field{
		Key: "details", Type: zapcore.StringType,
		String: detail.Details,
	}, zapcore.Field{
		Key: "clear_type", Type: zapcore.StringType,
		String: "ADAC",
	}, zapcore.Field{
		Key: "start_timestamp", Type: zapcore.StringType,
		String: strconv.Itoa(detail.StartTimestamp),
	}, zapcore.Field{
		Key: "end_timestamp", Type: zapcore.StringType,
		String: strconv.Itoa(detail.EndTimestamp),
	})
}

// ReportOrClearAlarm -
func ReportOrClearAlarm(alarmInfo *LogAlarmInfo, detail *Detail) {
	alarmLog, err := GetAlarmLogger()
	if err != nil {
		log.GetLogger().Errorf("GetAlarmLogger err %v", err)
		return
	}
	logger := addAlarmLogger(alarmLog, alarmInfo, detail)
	logger.Info("")
}

// SetAlarmEnv -
func SetAlarmEnv(alarmConfigInfo config.CoreInfo) {
	alarmConfigBytes, err := json.Marshal(alarmConfigInfo)
	if err != nil {
		log.GetLogger().Errorf("json marshal alarmConfigInfo err %v", err)
	}
	if err := os.Setenv(ConfigKey, string(alarmConfigBytes)); err != nil {
		log.GetLogger().Errorf("failed to set env of %s, err: %s", ConfigKey, err.Error())
	}
	log.GetLogger().Debugf("succeeded to set env of %s, value: %s", ConfigKey, string(alarmConfigBytes))
}

// SetXiangYunFourConfigEnv -
func SetXiangYunFourConfigEnv(xiangYunFourConfig types.XiangYunFourConfig) {
	if err := os.Setenv(constant.WiseCloudSite, xiangYunFourConfig.Site); err != nil {
		log.GetLogger().Errorf("failed to set env of %s, err: %s", constant.WiseCloudSite, err.Error())
	}
	if err := os.Setenv(constant.TenantID, xiangYunFourConfig.TenantID); err != nil {
		log.GetLogger().Errorf("failed to set env of %s, err: %s", constant.TenantID, err.Error())
	}
	if err := os.Setenv(constant.ApplicationID, xiangYunFourConfig.ApplicationID); err != nil {
		log.GetLogger().Errorf("failed to set env of %s, err: %s", constant.ApplicationID, err.Error())
	}
	if err := os.Setenv(constant.ServiceID, xiangYunFourConfig.ServiceID); err != nil {
		log.GetLogger().Errorf("failed to set env of %s, err: %s", constant.ServiceID, err.Error())
	}
	log.GetLogger().Debugf("succeeded to set env, value: %v", xiangYunFourConfig)
}

// SetPodIP -
func SetPodIP() error {
	ip, err := urnutils.GetServerIP()
	if err != nil {
		log.GetLogger().Errorf("failed to get pod ip, err: %s", err.Error())
		return err
	}
	err = os.Setenv(constant.PodIPEnvKey, ip)
	if err != nil {
		log.GetLogger().Errorf("failed to set env of %s, err: %s", constant.PodIPEnvKey, err.Error())
		return err
	}
	return nil
}
