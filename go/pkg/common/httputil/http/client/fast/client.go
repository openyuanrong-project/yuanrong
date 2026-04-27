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

// Package fast is fasthttp implementation of client
package fast

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"time"

	fhttp "github.com/valyala/fasthttp"

	"yuanrong.org/kernel/pkg/common/constants"
	"yuanrong.org/kernel/pkg/common/httputil/http"
	"yuanrong.org/kernel/pkg/common/httputil/utils"
	"yuanrong.org/kernel/pkg/common/uuid"
)

// FastClient fasthttp implement
type FastClient struct {
	Client *fhttp.Client
}

const (
	DefaultResponseHeadersSize = 16
	DeployTimeout              = 90
	Base                       = 10
)

func setRequestHeaders(request *fhttp.Request, headers map[string]string) {
	request.Header.Set(constants.HeaderTraceID, uuid.New().String())
	for key, value := range headers {
		request.Header.Set(key, value)
	}
}

// ParseFastResponse parse fhttp Response
func ParseFastResponse(response *fhttp.Response) (*http.SuccessResponse, error) {
	if response.StatusCode() == fhttp.StatusInternalServerError {
		// The call fails and the returned status code is 500, and the body contains the returned error message
		return nil, nil
	}
	if response.StatusCode() == fhttp.StatusOK {
		// The call is successful and the returned status code is 200 The body contains the returned information
		successResponse := &http.SuccessResponse{
			Body:    response.Body(),
			Headers: getResponseHeaders(response),
		}
		return successResponse, nil
	}
	// Other error codes return error information
	return nil, errors.New(fhttp.StatusMessage(response.StatusCode()))
}

func getResponseHeaders(response *fhttp.Response) map[string]string {
	headers := make(map[string]string, DefaultResponseHeadersSize)
	response.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	return headers
}

// ProcessMultipartRequestParams process multipart request params into fhttp request
func ProcessMultipartRequestParams(request *fhttp.Request, params map[string]string,
	bodyWriter *multipart.Writer, bodyBuffer *bytes.Buffer) (*fhttp.Request, error) {
	for key, val := range params {
		if err := bodyWriter.WriteField(key, val); err != nil {
			return nil, err
		}
	}
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}
	contentType := bodyWriter.FormDataContentType()
	request.Header.SetContentType(contentType)
	request.SetBody(bodyBuffer.Bytes())
	return request, nil
}

func (fast *FastClient) processMultipartRequest(request *fhttp.Request, params map[string]string,
	filePath string) (*fhttp.Request, error) {
	fileSize := utils.GetFileSize(filePath)
	request.Header.Set(http.HeaderContentType, http.Multipart)
	request.Header.Set(http.HeaderFileDigest, strconv.FormatInt(fileSize, Base))
	request.SetBodyString(strconv.FormatInt(fileSize, Base))

	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)
	if err := writeFile(bodyWriter, filePath); err != nil {
		return nil, err
	}
	return ProcessMultipartRequestParams(request, params, bodyWriter, bodyBuffer)
}

func writeFile(bodyWriter *multipart.Writer, filePath string) error {
	var (
		fileWriter io.Writer
		err        error
	)

	fileWriter, err = bodyWriter.CreateFormFile("file", path.Base(filePath))
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}
	return nil
}

// PostMultipart PostMultipart request
func (fast *FastClient) PostMultipart(url string, params map[string]string,
	headers map[string]string, filePath string) (*http.SuccessResponse, error) {
	request := fhttp.AcquireRequest()
	response := fhttp.AcquireResponse()
	setRequestHeaders(request, headers)
	request.SetRequestURI(url)
	request.Header.SetMethod(fhttp.MethodPost)
	request, err := fast.processMultipartRequest(request, params, filePath)
	if err != nil {
		return nil, err
	}
	fast.Client.ReadTimeout = DeployTimeout * time.Second
	if err := fast.Client.DoTimeout(request, response, DeployTimeout*time.Second); err != nil {
		return nil, err
	}
	return ParseFastResponse(response)
}

// Get Get request
func (fast *FastClient) Get(url string, headers map[string]string) (*http.SuccessResponse, error) {
	request := fhttp.AcquireRequest()
	response := fhttp.AcquireResponse()
	setRequestHeaders(request, headers)
	request.Header.Set(http.HeaderContentType, http.ApplicationJSONUTF8)
	request.Header.SetMethod(fhttp.MethodGet)
	request.SetRequestURI(url)

	if err := fast.Client.Do(request, response); err != nil {
		return nil, err
	}
	return ParseFastResponse(response)
}

// PutMultipart PutMultipart request
func (fast *FastClient) PutMultipart(url string, params map[string]string, headers map[string]string,
	filePath string) (*http.SuccessResponse, error) {
	request := fhttp.AcquireRequest()
	response := fhttp.AcquireResponse()
	setRequestHeaders(request, headers)
	request.SetRequestURI(url)
	request.Header.SetMethod(fhttp.MethodPut)
	request, err := fast.processMultipartRequest(request, params, filePath)
	if err != nil {
		return nil, err
	}
	fast.Client.ReadTimeout = DeployTimeout * time.Second
	if err := fast.Client.DoTimeout(request, response, DeployTimeout*time.Second); err != nil {
		return nil, err
	}
	return ParseFastResponse(response)
}
