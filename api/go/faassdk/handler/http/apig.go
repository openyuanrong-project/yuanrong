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

// Package http -
package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	log "yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

// APIGTriggerEvent extern interface of web request
type APIGTriggerEvent struct {
	IsBase64Encoded       bool                   `json:"isBase64Encoded"`
	HTTPMethod            string                 `json:"httpMethod"`
	Path                  string                 `json:"path"`
	Body                  string                 `json:"body"`
	PathParameters        map[string]string      `json:"pathParameters"`
	RequestContext        APIGRequestContext     `json:"requestContext"`
	Headers               map[string]interface{} `json:"headers"`
	QueryStringParameters map[string]interface{} `json:"queryStringParameters"`
	UserData              string                 `json:"user_data"`
}

// APIGRequestContext -
type APIGRequestContext struct {
	APIID     string `json:"apiId"`
	RequestID string `json:"requestId"`
	Stage     string `json:"stage"`
	SourceIP  string `json:"sourceIp"`
}

// APIGTriggerResponse extern interface of web response
type APIGTriggerResponse struct {
	Body            string              `json:"body"`
	Headers         map[string][]string `json:"headers"`
	StatusCode      int                 `json:"statusCode"`
	IsBase64Encoded bool                `json:"isBase64Encoded"`
}

func parseAPIGEvent(ctx context.Context, serializedEvent []byte, header map[string]string,
	baseURLPath, callRoute string) (*http.Request, error) {
	apigEvent := &APIGTriggerEvent{}
	if err := json.Unmarshal(serializedEvent, apigEvent); err != nil {
		log.GetLogger().Errorf("failed to unmarshal event to APIG event, error: %s", err)
		return nil, err
	}
	if apigEvent.Headers == nil {
		apigEvent.Headers = make(map[string]interface{})
	}
	for key, val := range header {
		apigEvent.Headers[key] = val
	}
	apigEvent.HTTPMethod = http.MethodPost
	return apigToHTTPRequest(ctx, apigEvent, baseURLPath, callRoute)
}

func constructAPIGResponse(resp *http.Response, body *bytes.Buffer) ([]byte, error) {
	apigResponse, err := fromHTTPResponse(resp, body)
	if err != nil {
		return nil, err
	}

	serializedResponse, err := json.Marshal(apigResponse)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal APIG response, error: %s", err)
		return nil, err
	}
	return serializedResponse, nil
}

func apigToHTTPRequest(ctx context.Context, event *APIGTriggerEvent, baseURLPath,
	callRoute string) (*http.Request, error) {
	var (
		requestBody []byte
		err         error
		requestURI  string
	)
	if event.IsBase64Encoded {
		requestBody, err = base64.StdEncoding.DecodeString(event.Body)
		if err != nil {
			log.GetLogger().Errorf("failed to decode body string, error: %s", err)
			return nil, err
		}
	} else {
		requestBody = []byte(event.Body)
	}
	if event.Path != "" {
		requestURI = fmt.Sprintf("%s/%s", baseURLPath, event.Path)
	} else {
		requestURI = fmt.Sprintf("%s/%s", baseURLPath, callRoute)
	}
	request, err := http.NewRequestWithContext(ctx, event.HTTPMethod, requestURI, bytes.NewBuffer(requestBody))
	if err != nil {
		log.GetLogger().Errorf("failed to construct HTTP request, error: %s", err.Error())
		return nil, err
	}
	queries := request.URL.Query()
	for k, v := range event.QueryStringParameters {
		switch valueType := v.(type) {
		case string:
			queries.Set(k, v.(string))
		case []string:
			for _, param := range v.([]string) {
				queries.Add(k, param)
			}
		default:
			log.GetLogger().Warnf("invalid type in query parameters: %T", valueType)
		}
	}
	request.URL.RawQuery = queries.Encode()

	initRequestHeader(request, event)
	return request, nil
}

func initRequestHeader(request *http.Request, event *APIGTriggerEvent) {
	for k, v := range event.Headers {
		if k == "Transfer-Encoding" {
			continue
		}
		switch valueType := v.(type) {
		case string:
			request.Header.Set(k, v.(string))
		case []string:
			for _, param := range v.([]string) {
				request.Header.Add(k, param)
			}
		default:
			log.GetLogger().Warnf("invalid type in headers: %T", valueType)
		}
	}

	request.Header.Set("X-APIG-Api-Id", event.RequestContext.APIID)
	request.Header.Set("X-APIG-Request-Id", event.RequestContext.RequestID)
	request.Header.Set("X-APIG-Source-Ip", event.RequestContext.SourceIP)
	request.Header.Set("X-APIG-Source-Stage", event.RequestContext.Stage)
	if host := request.Header.Get("Host"); host != "" {
		request.Host = host
	}
}

func fromHTTPResponse(response *http.Response, body *bytes.Buffer) (*APIGTriggerResponse, error) {
	apigResponse := &APIGTriggerResponse{
		Body:            base64.StdEncoding.EncodeToString(body.Bytes()),
		StatusCode:      response.StatusCode,
		IsBase64Encoded: true,
	}
	if response.Header != nil {
		apigResponse.Headers = make(map[string][]string)
		for k, v := range response.Header {
			apigResponse.Headers[k] = v
		}
	}
	return apigResponse, nil
}
