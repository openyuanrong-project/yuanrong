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

// Package healthcheck implement health check
package healthcheck

import (
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	readTimeout     = 180 * time.Second
	writeTimeout    = 180 * time.Second
	healthCheckPort = "9994"
)

// StartHealthCheck -
func StartHealthCheck(errChan chan<- error) error {
	if !config.GlobalConfig.EnableHealthCheck {
		return nil
	}
	if errChan == nil {
		return errors.New("errChan is nil")
	}

	router := createRouter()

	podIP := os.Getenv("POD_IP")
	if net.ParseIP(podIP) == nil {
		log.GetLogger().Errorf("failed to get pod ip, pod ip is %s", podIP)
		return errors.New("failed to get pod ip")
	}
	addr := podIP + ":" + healthCheckPort

	httpServer := &http.Server{
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Addr:         addr,
	}

	go func() {
		err := startServer(httpServer)
		if err != nil {
			log.GetLogger().Errorf("failed to startServer, err %s", err.Error())
		}
		errChan <- err
	}()

	return nil
}

func startServer(httpServer *http.Server) error {
	if config.GlobalConfig.HTTPSConfig == nil || !config.GlobalConfig.HTTPSConfig.HTTPSEnable {
		err := httpServer.ListenAndServe()
		if err != nil {
			log.GetLogger().Errorf("failed to http ListenAndServe: %s", err.Error())
		}
		return err
	}
	err := tls.InitTLSConfig(*config.GlobalConfig.HTTPSConfig)
	if err != nil {
		log.GetLogger().Errorf("failed to init the HTTPS config: %s", err.Error())
		return err
	}
	httpServer.TLSConfig = tls.GetClientTLSConfig()
	err = httpServer.ListenAndServeTLS("", "")
	if err != nil {
		log.GetLogger().Errorf("failed to HTTPListenAndServeTLS: %s", err.Error())
		return err
	}
	return nil
}

func createRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/healthcheck", func(c *gin.Context) {
		check(c.Writer, c.Request)
	})
	return router
}

func check(w http.ResponseWriter, r *http.Request) {
	discoveryConfig := config.GlobalConfig.SchedulerDiscovery
	if discoveryConfig != nil && discoveryConfig.RegisterMode == types.RegisterTypeContend {
		if !selfregister.Registered {
			log.GetLogger().Warnf("health check now, scheduler is not registered")
			if config.GlobalConfig.EnableRollout && selfregister.IsRolloutObject {
				log.GetLogger().Infof("health check now, scheduler is the rollout object")
				w.WriteHeader(http.StatusOK)
			} else {
				log.GetLogger().Errorf("health check now, scheduler is not rollout object")
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	return
}
