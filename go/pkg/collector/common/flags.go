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

package common

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// CertConfig -
type CertConfig struct {
	SslEnable bool   `json:"sslEnable" valid:"optional"`
	CaFile    string `json:"cafile,omitempty" valid:"optional"`
	CertFile  string `json:"certfile,omitempty" valid:"optional"`
	KeyFile   string `json:"keyfile,omitempty" valid:"optional"`
}

// CollectorConfig -
type CollectorConfig struct {
	CollectorID          string
	IP                   string
	Port                 string
	Address              string
	ManagerAddress       string
	LogRoot              string
	UserLogPath          string
	EtcdConfig           etcd3.EtcdConfig
	FunctionSystemConfig CertConfig
}

var (
	version     string
	showVersion bool
	// CollectorConfigs -
	CollectorConfigs CollectorConfig
)

const (
	collectorIDPattern string = `^[a-zA-Z0-9~\.\-\/_!@#\%\^&\*\(\)\+\=\:;]{1,256}$`
	portLowerBound     int    = 0
	portUpperBound     int    = 65536
)

var rootCmd = &cobra.Command{
	Short: "Collector is to collect user logs.",
	Long:  "Collector is to collect user-level logs by interactively streaming or specified range of lines.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	registerCmdArgs(rootCmd)
}

// InitCmd init commandline
func InitCmd() {
	if err := rootCmd.Execute(); err != nil {
		log.GetLogger().Fatal(err.Error())
	}
	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	if err := validateCmdArgs(); err != nil {
		log.GetLogger().Fatal(err.Error())
	}
	log.GetLogger().Infof("collector start args: %+v", CollectorConfigs)
}

func registerCmdArgs(rootCmd *cobra.Command) {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")
	rootCmd.Flags().StringVarP(&CollectorConfigs.CollectorID, "collect_id", "", "",
		"the identifier; of length less than 256")
	rootCmd.Flags().StringVarP(&CollectorConfigs.IP, "ip", "", "", "the ip of collector for manager to access")
	rootCmd.Flags().StringVarP(&CollectorConfigs.Port, "port", "", "", "the port of collector for manager to access")
	rootCmd.Flags().StringVarP(&CollectorConfigs.ManagerAddress, "manager_address", "", "",
		"manager address to register collector")
	rootCmd.Flags().StringVarP(&CollectorConfigs.LogRoot, "log_root", "", "", "the default root path of all logs")
	rootCmd.Flags().StringVarP(&CollectorConfigs.UserLogPath, "user_log_path", "", "",
		"optional; specified only if user log is in other directory than log root path")

	rootCmd.Flags().StringSliceVarP(&CollectorConfigs.EtcdConfig.Servers, "etcd_config_servers", "", []string{},
		"etcd server addresses")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.User, "etcd_config_user", "", "",
		"etcd config about user")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.Password, "etcd_config_password", "", "",
		"etcd config about password")
	rootCmd.Flags().BoolVarP(&CollectorConfigs.EtcdConfig.SslEnable, "etcd_config_ssl_enable", "", false,
		"etcd config about ssl_enable")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.AuthType, "etcd_config_auth_type", "", "",
		"etcd config about auth_type")
	rootCmd.Flags().BoolVarP(&CollectorConfigs.EtcdConfig.UseSecret, "etcd_config_use_secret", "", false,
		"etcd config about use_secret")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.SecretName, "etcd_config_secret_name", "", "",
		"etcd config about secret_name")
	rootCmd.Flags().IntVarP(&CollectorConfigs.EtcdConfig.LimitRate, "etcd_config_limit_rate", "", 0,
		"etcd config about limit_rate")
	rootCmd.Flags().IntVarP(&CollectorConfigs.EtcdConfig.LimitBurst, "etcd_config_limit_burst", "", 0,
		"etcd config about limit_burst")
	rootCmd.Flags().IntVarP(&CollectorConfigs.EtcdConfig.LimitTimeout, "etcd_config_limit_timeout", "", 0,
		"etcd config about limit_timeout")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.CaFile, "etcd_config_ca_file", "", "",
		"etcd config about ca_file")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.CertFile, "etcd_config_cert_file", "", "",
		"etcd config about cert_file")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.KeyFile, "etcd_config_key_file", "", "",
		"etcd config about key_file")
	rootCmd.Flags().StringVarP(&CollectorConfigs.EtcdConfig.PassphraseFile, "etcd_config_passphrase_file", "", "",
		"etcd config about passphrase_file")
	rootCmd.Flags().BoolVarP(&CollectorConfigs.FunctionSystemConfig.SslEnable, "function_system_ssl_enable", "", false,
		"function system config about ssl_enable")
	rootCmd.Flags().StringVarP(&CollectorConfigs.FunctionSystemConfig.CaFile, "function_system_ca_file", "", "",
		"function system config about ca_file")
	rootCmd.Flags().StringVarP(&CollectorConfigs.FunctionSystemConfig.CertFile, "function_system_cert_file", "", "",
		"function system config about cert_file")
	rootCmd.Flags().StringVarP(&CollectorConfigs.FunctionSystemConfig.KeyFile, "function_system_key_file", "", "",
		"function system config about key_file")
}

func checkPath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s should be absolute path", path)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("no such directory: %s", path)
	}
	return nil
}

func validateCmdArgs() error {
	matched, merr := regexp.MatchString(collectorIDPattern, CollectorConfigs.CollectorID)
	if merr != nil {
		return fmt.Errorf("collector ID failed to match regex: %s, collector id: %s", collectorIDPattern,
			CollectorConfigs.CollectorID)
	}
	if !matched {
		return fmt.Errorf(
			"collector ID %s does not match %s; please check characters and length. The valid length range is [1, 256)",
			CollectorConfigs.CollectorID, collectorIDPattern)
	}
	if net.ParseIP(CollectorConfigs.IP) == nil {
		return fmt.Errorf("collector IP %s is invalid", CollectorConfigs.IP)
	}
	if p, err := strconv.Atoi(CollectorConfigs.Port); err != nil || p < portLowerBound || p > portUpperBound {
		return fmt.Errorf("collector port %s is invalid", CollectorConfigs.Port)
	}
	CollectorConfigs.Address = CollectorConfigs.IP + ":" + CollectorConfigs.Port

	host, port, err := net.SplitHostPort(CollectorConfigs.ManagerAddress)
	if err != nil {
		return fmt.Errorf("manager address %s is invalid, error: %s", CollectorConfigs.ManagerAddress, err)
	}
	if net.ParseIP(host) == nil {
		return fmt.Errorf("manager address %s has invalid IP", CollectorConfigs.ManagerAddress)
	}
	if p, err := strconv.Atoi(port); err != nil || p < portLowerBound || p > portUpperBound {
		return fmt.Errorf("manager address %s has invalid port", CollectorConfigs.ManagerAddress)
	}

	if err := checkPath(CollectorConfigs.LogRoot); err != nil {
		return err
	}
	if CollectorConfigs.UserLogPath == "" {
		CollectorConfigs.UserLogPath = CollectorConfigs.LogRoot
		log.GetLogger().Infof("user log path inherits log root %s", CollectorConfigs.UserLogPath)
	} else {
		if err := checkPath(CollectorConfigs.UserLogPath); err != nil {
			return err
		}
	}
	return nil
}
