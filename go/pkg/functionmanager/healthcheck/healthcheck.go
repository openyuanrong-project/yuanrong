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

// Package healthcheck -
package healthcheck

import (
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionmanager/config"
)

const (
	readTimeout     = 180 * time.Second
	writeTimeout    = 180 * time.Second
	healthCheckPort = "9994"
)

// StartHealthCheck -
func StartHealthCheck(errChan chan<- error) error {
	if !config.GetConfig().EnableHealthCheck {
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
			log.GetLogger().Errorf("failed to startServer, err: %s", err.Error())
		}
		errChan <- err
	}()
	return nil
}

func startServer(httpServer *http.Server) error {
	err := httpServer.ListenAndServe()
	if err != nil {
		log.GetLogger().Errorf("failed to ListenAndServe: %s", err.Error())
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
	w.WriteHeader(http.StatusOK)
	return
}
