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

package logmanager

import (
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/dashboard/models"
)

type listLogQuery struct {
	SubmissionID string `form:"submission_id" json:"submission_id"`
	InstanceID   string `form:"instance_id" json:"instance_id"`
}

type readLogQuery struct {
	Filename  string `form:"filename" json:"filename" validate:"required" example:"runtime-01234.log"`
	StartLine uint32 `form:"start_line" json:"start_line" example:"0"`
	EndLine   uint32 `form:"end_line" json:"end_line" example:"5000"`
}

type listLogsResponse = map[string][]string

// ListLogsHandler function for get /logs, will list all logs by filename
//
//		@Summary		list all log files (INTERNAL ONLY)
//		@Description	return a map use collectorID (like nodeID) as key, filename(s) as value
//		@Produce		json
//	 @Tags           logs,internal
//		@Param			params	query		listLogQuery											false	"params"
//		@Success		200		{object}	models.DashboardCommonResponse{data=listLogsResponse}	"success"
//		@Failure		400		{object}	models.DashboardErrorResponse							"invalid parameter"
//		@Failure		500		{object}	models.DashboardErrorResponse							"some internal error, need to check log"
//		@Router			/api/logs/list [get]
func ListLogsHandler(ctx *gin.Context) {
	if managerSingleton == nil {
		log.GetLogger().Errorf("manager is nullptr!!")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "unexpected error",
		})
		return
	}
	query := listLogQuery{}
	err := ctx.BindQuery(&query)
	if err != nil {
		log.GetLogger().Errorf("bind query failed: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, models.DashboardCommonResponse{
			Message: fmt.Sprintf("query binding failed: %s", err.Error()),
			Data:    nil,
		})
		return
	}
	logEntries := managerSingleton.LogDB.Query(logDBQuery{
		SubmissionID: query.SubmissionID,
		InstanceID:   query.InstanceID,
	})
	res := listLogsResponse{}
	logEntries.Range(func(e *LogEntry) {
		res[e.CollectorID] = append(res[e.CollectorID], e.Filename)
	})
	ctx.JSON(http.StatusOK, models.SuccessDashboardCommonResponse(res))
}

// ReadLogHandler function for read logs
//
//		@Summary		read log file content (INTERNAL ONLY)
//		@Description	return log file (use a stream)
//		@Produce		octet-stream
//	 @Tags           logs,internal
//		@Param			request	query		readLogQuery					false	"query params"
//		@Success		200		{object}	string							"success"
//		@Failure		400		{object}	models.DashboardErrorResponse	"invalid parameter"
//		@Failure		500		{object}	models.DashboardErrorResponse	"some internal error, need to check log"
//		@Router			/api/logs [get]
func ReadLogHandler(ctx *gin.Context) {
	query, err := bindLogQuery(ctx)
	if err != nil {
		return
	}
	streamLogHandler(ctx, query, false)
}

// DownloadLogHandler function for download log
//
//		@Summary		download log file (INTERNAL ONLY)
//		@Description	return log file (use a stream)
//		@Produce		octet-stream
//	 @Tags           logs,internal
//		@Param			filename query		string					        false	"query params"
//		@Success		200		 {file}	    binary  						"success"
//		@Failure		400		 {object}	models.DashboardErrorResponse	"invalid parameter"
//		@Failure		500		 {object}	models.DashboardErrorResponse	"some internal error, need to check log"
//		@Router			/api/logs/download [get]
func DownloadLogHandler(ctx *gin.Context) {
	query, err := bindLogQuery(ctx)
	if err != nil {
		return
	}
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", query.Filename))
	ctx.Header("Transfer-Encoding", "chunked") // 关键：设置分块传输
	ctx.Header("X-Content-Type-Options", "nosniff")
	query.StartLine = 0
	query.EndLine = math.MaxUint32
	streamLogHandler(ctx, query, true)
}

func bindLogQuery(ctx *gin.Context) (readLogQuery, error) {
	log.GetLogger().Infof("receive read log request %#v", ctx.Request)
	if managerSingleton == nil {
		log.GetLogger().Errorf("!! manager is nil !! this should never happened !!")
		ctx.JSON(http.StatusInternalServerError, models.DashboardCommonResponse{
			Message: "unexpected error",
			Data:    nil,
		})
		return readLogQuery{}, errors.New("manager is nil, this should never happened")
	}
	query := readLogQuery{}
	err := ctx.BindQuery(&query)
	if err != nil {
		log.GetLogger().Errorf("bind query failed: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.DashboardCommonResponse{
			Message: fmt.Sprintf("query binding failed: %s", err.Error()),
			Data:    nil,
		})
		return query, err
	}
	return query, nil
}

func streamLogHandler(ctx *gin.Context, query readLogQuery, shouldFlush bool) {
	log.GetLogger().Infof("readlog handler get log query: %#v", query.Filename)
	log.GetLogger().Infof("DB right now: %#v", managerSingleton.GetLogEntries())
	logEntries := managerSingleton.LogDB.Query(logDBQuery{
		Filename: query.Filename,
	})
	log.GetLogger().Infof("found Entries: %#v", logEntries)
	if logEntries.Len() != 1 {
		msg := fmt.Sprintf("can not find log file or duplicate log files found, get %d files", logEntries.Len())
		log.GetLogger().Warn(msg)
		ctx.JSON(http.StatusInternalServerError, models.DashboardCommonResponse{
			Message: msg,
			Data:    nil,
		})
		return
	}

	outLog := make(chan *logservice.ReadLogResponse, 100)
	entry := logEntries.FindFirst()

	wait := make(chan struct{})
	go streamResponse(ctx, outLog, wait, shouldFlush)
	publishCollectRequest(ctx, entry, outLog, query)
	<-wait
}

func publishCollectRequest(ctx *gin.Context, entry *LogEntry, outLog chan<- *logservice.ReadLogResponse,
	query readLogQuery) {
	req := &logservice.ReadLogRequest{
		Item:      entry.LogItem,
		StartLine: query.StartLine,
		EndLine:   query.EndLine,
	}
	c := managerSingleton.GetCollector(entry.CollectorID)
	if c == nil {
		msg := fmt.Sprintf("can not find collector %s for file %s", entry.CollectorID, entry.Filename)
		log.GetLogger().Warn(msg)
		ctx.JSON(http.StatusInternalServerError, models.DashboardCommonResponse{
			Message: msg,
			Data:    nil,
		})
		close(outLog)
		return
	}
	err := c.CollectLog(ctx, req, outLog)
	if err != nil {
		log.GetLogger().Warnf("failed to collect log, err: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, models.DashboardCommonResponse{
			Message: err.Error(),
			Data:    nil,
		})
		return
	}
}

func streamResponse(ctx *gin.Context, outLog <-chan *logservice.ReadLogResponse, wait chan struct{}, shouldFlush bool) {
	defer close(wait)
	for {
		select {
		case resp, ok := <-outLog:
			if !ok {
				log.GetLogger().Info("end to read log.")
				return
			}
			if resp.Code != 0 {
				log.GetLogger().Warnf("failed to read log, errCode: %d, errMsg: %s", resp.Code, resp.Message)
				return
			}
			_, err := ctx.Writer.Write(resp.Content)
			if err != nil {
				log.GetLogger().Warnf("failed to write resp, err: %s", err.Error())
				return
			}
			if shouldFlush {
				ctx.Writer.Flush()
			}
		}
	}
}
