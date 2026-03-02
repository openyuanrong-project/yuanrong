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

// Package flags for obtain command line params
package flags

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/asaskevich/govalidator/v11"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/reader"
)

const (
	dashboardLogFileName = "dashboard"
	defaultDashboardPort = 9080

	defaultEtcdOpTimeout                  = 60 * time.Second
	dashboardRegisterKeyKeepAliveInterval = 10 * time.Second
	dashboardRegisterKeyTTL               = 60
	dashboardRegisterKey                  = "/yr/dashboard"
)

// CertConfig -
type CertConfig struct {
	SslEnable bool   `json:"sslEnable" valid:"optional"`
	CaFile    string `json:"cafile,omitempty" valid:"optional"`
	CertFile  string `json:"certfile,omitempty" valid:"optional"`
	KeyFile   string `json:"keyfile,omitempty" valid:"optional"`
}

type dashboardConfig struct {
	StaticPath           string           `json:"staticPath"`
	Ip                   string           `json:"ip" valid:"ip"`
	GrpcIP               string           `json:"grpcIP" valid:"ip"`
	Port                 int              `json:"port" valid:"port"`
	GrpcPort             int              `json:"grpcPort" valid:"port"`
	FunctionMasterAddr   string           `json:"functionMasterAddr" valid:"url"`
	FrontendAddr         string           `json:"frontendAddr" valid:"url"`
	PrometheusAddr       string           `json:"prometheusAddr" valid:"url"`
	RouterEtcdConfig     etcd3.EtcdConfig `json:"routerEtcdConfig"`
	MetaEtcdConfig       etcd3.EtcdConfig `json:"metaEtcdConfig"`
	FunctionSystemConfig CertConfig       `json:"functionSystemConfig"`
	PrometheusConfig     CertConfig       `json:"prometheusConfig"`
	SslConfig            CertConfig       `json:"sslConfig"`

	ServerAddr string
}

const (
	// frontend jobs instance url
	defaultListenIP = "0.0.0.0"
	appBasePath     = "/app/v1"
	httpPrefix      = "http://"
	httpsPrefix     = "https://"
)

var (
	version     string
	showVersion bool
	// DashboardConfig is the global config struct
	DashboardConfig dashboardConfig

	dashboardConfigPath    string
	dashboardLogConfigPath string
)

// AddHTTPPrefix -
func AddHTTPPrefix(url string, enableSsl bool) string {
	if strings.Contains(url, "://") {
		return url
	}
	if enableSsl {
		url = httpsPrefix + url
	} else {
		url = httpPrefix + url
	}
	return url
}

func initLog() error {
	if err := log.InitRunLog(dashboardLogFileName, true); err != nil {
		log.GetLogger().Errorf("failed to init dashboard log, err: %s", err.Error())
	}
	return nil
}

func initConfig(configFilePath string) error {
	// InitConfig get config info from configPath
	data, err := reader.ReadFileWithTimeout(configFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read config, filename: %s, error: %s", configFilePath, err.Error())
		return err
	}

	err = json.Unmarshal(data, &DashboardConfig)
	if err != nil {
		log.GetLogger().Errorf("failed to unmarshal config, configPath: %s, error: %s", configFilePath, err.Error())
		return err
	}

	_, err = govalidator.ValidateStruct(DashboardConfig)
	if err != nil {
		log.GetLogger().Errorf("failed to validate config, err: %s", err.Error())
		return err
	}
	return nil
}

func init() {
	cCmd := cobraCmd()
	initWithConfigFiles(cCmd)
	initWithParams(cCmd)

	log.GetLogger().Infof("init done")
}

// InitEtcdClient -
func InitEtcdClient() error {
	stopCh := make(chan struct{}, 1)
	log.GetLogger().Infof("init etcd client...")
	return etcd3.InitParam().
		WithRouteEtcdConfig(DashboardConfig.RouterEtcdConfig).
		WithMetaEtcdConfig(DashboardConfig.MetaEtcdConfig).
		WithStopCh(stopCh).
		InitClient()
}

// RegisterSelfToEtcd -
func RegisterSelfToEtcd(stopCh <-chan struct{}) error {
	log.GetLogger().Infof("registering self to etcd")
	if stopCh == nil {
		return errors.New("etcd stopCh should not be nil")
	}
	leaseID, err := etcd3.GetRouterEtcdClient().Grant(etcd3.CreateEtcdCtxInfoWithTimeout(context.TODO(),
		defaultEtcdOpTimeout), dashboardRegisterKeyTTL)
	if err != nil {
		log.GetLogger().Infof("lease: %d", leaseID)
		return err
	}
	go func() {
		for {
			select {
			case <-stopCh:
				log.GetLogger().Warnf("stopCh signal recevied, exit the keep alive channel")
				return
			default:
				err := etcd3.GetRouterEtcdClient().KeepAliveOnce(etcd3.CreateEtcdCtxInfoWithTimeout(context.TODO(),
					dashboardRegisterKeyKeepAliveInterval), leaseID)
				if err != nil {
					log.GetLogger().Errorf("failed to revoke self registry in etcd: %s", err)
				}
				time.Sleep(dashboardRegisterKeyKeepAliveInterval)
			}
		}
	}()
	log.GetLogger().Infof("registering self to etcd with key and v: %s with lease: %d", DashboardConfig.ServerAddr,
		leaseID)
	err = etcd3.GetRouterEtcdClient().Put(etcd3.CreateEtcdCtxInfoWithTimeout(context.TODO(), defaultEtcdOpTimeout),
		dashboardRegisterKey, fmt.Sprintf("%s:%d", DashboardConfig.GrpcIP, DashboardConfig.GrpcPort),
		clientv3.WithLease(leaseID))
	if err != nil {
		return err
	}
	return nil
}

func initWithConfigFiles(cCmd *cobra.Command) {
	cmdErr(cCmd.Execute())
	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if err := initLog(); err != nil {
		log.GetLogger().Errorf("failed to init dashboard log, err: %s", err.Error())
	}
	if err := initConfig(dashboardConfigPath); err != nil {
		log.GetLogger().Errorf("failed to init dashboard config: %s", err)
	}
}

func initWithParams(cCmd *cobra.Command) {
	cmdErr(cCmd.Execute())
	DashboardConfig.FunctionMasterAddr = AddHTTPPrefix(DashboardConfig.FunctionMasterAddr,
		DashboardConfig.FunctionSystemConfig.SslEnable)
	DashboardConfig.FrontendAddr = AddHTTPPrefix(DashboardConfig.FrontendAddr,
		DashboardConfig.FunctionSystemConfig.SslEnable) + appBasePath
	DashboardConfig.PrometheusAddr = AddHTTPPrefix(DashboardConfig.PrometheusAddr,
		DashboardConfig.PrometheusConfig.SslEnable)
	ip := DashboardConfig.Ip
	if len(ip) == 0 {
		ip = defaultListenIP
	}
	DashboardConfig.ServerAddr = fmt.Sprintf("%s:%d", ip, DashboardConfig.Port)
}

func cobraCmd() *cobra.Command {
	cCmd := &cobra.Command{
		Long: "This params for starting dashboard server.",
		Run:  func(cmd *cobra.Command, args []string) {},
	}

	// if config path, use config path
	cCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")
	cCmd.Flags().StringVarP(&dashboardConfigPath, "config_path", "", "", "config file path")
	cCmd.Flags().StringVarP(&dashboardLogConfigPath, "log_config_path", "", "", "log config file path")

	// if command line, use command line
	cCmd.Flags().StringVarP(&DashboardConfig.FunctionMasterAddr, "function_master_addr", "", "",
		"FunctionMasterURL format is <function-master-ip>:<port>")
	cCmd.Flags().StringVarP(&DashboardConfig.FrontendAddr, "frontend_addr", "", "",
		"FrontendURL format is <frontend-ip>:<port>")
	cCmd.Flags().StringVarP(&DashboardConfig.PrometheusAddr, "prometheus_addr", "", "",
		"PrometheusURL format is <prometheus-ip>:<port>")
	cCmd.Flags().StringVarP(&DashboardConfig.Ip, "ip", "i", "0.0.0.0", "this service listening with this ip")
	cCmd.Flags().IntVarP(&DashboardConfig.Port, "port", "p", defaultDashboardPort, "this service listening with this port")
	cCmd.Flags().StringVarP(&DashboardConfig.StaticPath, "static_path", "s", "./client/dist",
		"this is client static resources path")
	return cCmd
}

func cmdErr(err error) {
	if err != nil {
		log.GetLogger().Fatal(err.Error())
	}
}

func loadConfig(configPath string, loadFunc func(configPath string) error) {
	if loadFunc == nil {
		return
	}
	if err := loadFunc(configPath); err != nil {
		log.GetLogger().Fatal(err.Error())
	}
}
