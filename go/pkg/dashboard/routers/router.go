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

// Package routers for load routs
package routers

import (
	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/job"
	"yuanrong.org/kernel/pkg/dashboard/flags"
	"yuanrong.org/kernel/pkg/dashboard/handlers"
	"yuanrong.org/kernel/pkg/dashboard/logmanager"
)

const (
	basePath       = "/api/v1"
	apiPath        = "/api"
	rsrcSummPath   = "/logical-resources/summary"
	rsrcPath       = "/logical-resources"
	rsrcUnitIDPath = "/logical-resources/:unit-id"
	compPath       = "/components"
	compCompIDPath = "/components/:component-id"
	instPath       = "/instances"
	instSummPath   = "/instances/summary"
	instInstIDPath = "/instances/:instance-id"
	promPath       = "/prometheus/query"

	// compatible with ray serve
	servePath    = "/api/serve"
	serveAppPath = "/applications"

	clusterStatusPath = "/cluster_status"

	logsListPath = "/logs/list"
	logsPath     = "/logs"
)

// SetRouter function for set routs
func SetRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(Cors())

	r.StaticFile("/", flags.DashboardConfig.StaticPath+"/index.html")
	r.StaticFile("/favicon.ico", flags.DashboardConfig.StaticPath+"/favicon.ico")
	r.Static("/assets", flags.DashboardConfig.StaticPath+"/assets")

	v1Group := r.Group(basePath)
	{
		// resources
		v1Group.GET(rsrcSummPath, handlers.ResourcesSummaryHandler)
		v1Group.GET(rsrcPath, handlers.ResourcesHandler)
		v1Group.GET(rsrcUnitIDPath, handlers.ResourcesByUnitIDHandler)
		// components
		v1Group.GET(compPath, handlers.ComponentsHandler)
		v1Group.GET(compCompIDPath, handlers.ComponentsByComponentIDHandler)
		// instances
		v1Group.GET(instPath, handlers.InstancesHandler)
		v1Group.GET(instSummPath, handlers.InstancesSummaryHandler)
		v1Group.GET(instInstIDPath, handlers.InstancesByInstanceIDHandler)
		// prom
		v1Group.GET(promPath, handlers.PrometheusHandler)
	}

	rayServeGroup := r.Group(servePath)
	{
		// serve
		rayServeGroup.GET(serveAppPath, handlers.ServeGetHandler)
		rayServeGroup.PUT(serveAppPath, handlers.ServePutHandler)
		rayServeGroup.DELETE(serveAppPath, handlers.ServeDelHandler)
	}

	jobGroup := r.Group(job.PathGroupJobs)
	{
		// jobs
		jobGroup.POST("", handlers.SubmitJobHandler)
		jobGroup.GET("", handlers.ListJobsHandler)
		jobGroup.GET(job.PathGetJobs, handlers.GetJobInfoHandler)
		jobGroup.DELETE(job.PathDeleteJobs, handlers.DeleteJobHandler)
		jobGroup.POST(job.PathStopJobs, handlers.StopJobHandler)
	}

	apiGroup := r.Group(apiPath)
	{
		// pending queue resources
		apiGroup.GET(clusterStatusPath, handlers.ClusterStatusHandler)

		apiGroup.GET(logsListPath, logmanager.ListLogsHandler)
		apiGroup.GET(logsPath, logmanager.ReadLogHandler)
	}

	return r
}
