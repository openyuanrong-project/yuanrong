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

package localauth

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
)

func TestAuthCheckLocally(t *testing.T) {
	type args struct {
		ak          string
		sk          string
		requestSign string
		timestamp   string
		duration    int
	}
	var a args
	var b args
	b.requestSign = "aaa"
	var c args
	c.requestSign = "aaa"
	c.timestamp = strconv.FormatInt(time.Now().AddDate(1, 0, 0).Unix(), 10)
	var d args
	d.requestSign = "aaa"
	d.timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	var e args
	e.requestSign = "aaa,appid=aaa"
	e.timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", a, true},
		{"case2", b, true},
		{"case3", c, true},
		{"case4", d, true},
		{"case5", e, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AuthCheckLocally(tt.args.ak, tt.args.sk, tt.args.requestSign, tt.args.timestamp, tt.args.duration); (err != nil) != tt.wantErr {
				t.Errorf("AuthCheckLocally() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_GetTimestampDiffLimit(t *testing.T) {
	tsDiffLimit := getTimestampDiffLimit()
	assert.Equal(t, 5, int(tsDiffLimit))
	patches := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
		return "100"
	})
	tsDiffLimit = getTimestampDiffLimit()
	assert.Equal(t, 100, int(tsDiffLimit))
	defer patches.Reset()
}

func TestCreateAuthorization(t *testing.T) {
	patches := [...]*gomonkey.Patches{
		gomonkey.ApplyFunc(DecryptKeys,
			func(_ string, _ string) ([]byte, []byte, error) {
				return []byte{}, []byte{}, errors.New("aaa")
			}),
	}
	defer func() {
		for idx := range patches {
			patches[idx].Reset()
		}
	}()
	type args struct {
		ak    string
		sk    string
		url   string
		appID string
		data  []byte
	}
	var a args
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"case1", a, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CreateAuthorization(tt.args.ak, tt.args.sk, tt.args.url, tt.args.appID, tt.args.data)
			if got != tt.want {
				t.Errorf("CreateAuthorization() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CreateAuthorization() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSignOMSVC(t *testing.T) {
	patches := [...]*gomonkey.Patches{
		gomonkey.ApplyFunc(DecryptKeys,
			func(_ string, _ string) ([]byte, []byte, error) {
				return []byte{}, []byte{}, errors.New("aaa")
			}),
	}
	defer func() {
		for idx := range patches {
			patches[idx].Reset()
		}
	}()
	type args struct {
		ak   string
		sk   string
		url  string
		data []byte
	}
	var a args
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"case1", a, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := SignOMSVC(tt.args.ak, tt.args.sk, tt.args.url, tt.args.data)
			if got != tt.want {
				t.Errorf("SignOMSVC() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SignOMSVC() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSigner_buildCanonicalHeaders(t *testing.T) {
	type fields struct {
		signTime    time.Time
		serviceName string
		region      string
	}
	type args struct {
		request *http.Request
	}
	var f fields
	var a args
	request := &http.Request{
		Method:           "",
		URL:              nil,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             nil,
		GetBody:          nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Host:             "",
		Form:             nil,
		PostForm:         nil,
		MultipartForm:    nil,
		Trailer:          nil,
		RemoteAddr:       "",
		RequestURI:       "",
		TLS:              nil,
		Response:         nil,
	}
	a.request = request
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"case1", f, a, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := &Signer{
				signTime:    tt.fields.signTime,
				serviceName: tt.fields.serviceName,
				region:      tt.fields.region,
			}
			if got := sig.buildCanonicalHeaders(tt.args.request); got != tt.want {
				t.Errorf("buildCanonicalHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSigner(t *testing.T) {
	type args struct {
		mode        string
		serviceName string
		region      string
	}
	var a args
	a.mode = "aaa"
	signer := &Signer{}
	tests := []struct {
		name string
		args args
		want *Signer
	}{
		{"case1", a, signer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSigner(tt.args.mode, tt.args.serviceName, tt.args.region); !reflect.DeepEqual(got.serviceName, tt.want.serviceName) {
				t.Errorf("getSigner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_signLocalAuthRequest(t *testing.T) {
	patches := [...]*gomonkey.Patches{
		gomonkey.ApplyFunc(url.Parse,
			func(_ string) (*url.URL, error) {
				return nil, errors.New("aaa")
			}),
	}
	defer func() {
		for idx := range patches {
			patches[idx].Reset()
		}
	}()
	type args struct {
		rawURL    string
		timeStamp string
		appID     string
		key       *Authentication
		data      []byte
	}
	var a args
	var b args
	b.timeStamp = strconv.FormatInt(time.Now().Unix(), 10)
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1", a, ""},
		{"case2", b, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := signLocalAuthRequest(tt.args.rawURL, tt.args.timeStamp, tt.args.appID, tt.args.key, tt.args.data); got != tt.want {
				t.Errorf("signLocalAuthRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignLocally(t *testing.T) {
	convey.Convey("TestSignLocally", t, func() {
		auth, time := SignLocally("ak", "sk", "appID", 0)
		convey.So(auth, convey.ShouldNotBeEmpty)
		convey.So(time, convey.ShouldNotBeEmpty)
	})
}

func TestSignWithHmacSha256(t *testing.T) {
	convey.Convey("Test SignWithHmacSha256", t, func() {
		convey.Convey("when success", func() {
			req := &fasthttp.Request{}
			req.Header.SetMethod("GET")
			req.SetRequestURI("http://7.218.80.122:31223" + "/invoke")
			req.Header.Set("key1", "value1")
			req.Header.Set("key2", "value2")
			req.Header.Set(constant.HeaderSignedHeader, "key1;key2")
			ak := "yuanrong"
			sk := "C0ECCDF386D96B8BF90562BC4F25C7608A8AA8D84003F088FCDC5E8606314A81"
			err := SignWithHmacSha256(req, ak, sk)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(req.Header.Peek(constant.HeaderAuthorization)), convey.ShouldNotBeEmpty)
		})
	})
}

func TestVerifySignWithHmacSha256(t *testing.T) {
	convey.Convey("Test VerifySignWithHmacSha256", t, func() {
		convey.Convey("when success", func() {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod("GET")
			ctx.Request.SetRequestURI("http://7.218.80.122:31223" + "/invoke")
			ctx.Request.Header.Set("key1", "value1")
			ctx.Request.Header.Set("key2", "value2")
			ctx.Request.Header.Set(constant.HeaderSignedHeader, "key1;key2")
			ak := "yuanrong"
			sk := "C0ECCDF386D96B8BF90562BC4F25C7608A8AA8D84003F088FCDC5E8606314A81"
			err := SignWithHmacSha256(&ctx.Request, ak, sk)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(ctx.Request.Header.Peek(constant.HeaderAuthorization)), convey.ShouldNotBeEmpty)

			err = VerifySignWithHmacSha256(ctx, ak, sk)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_checkTimestamp(t *testing.T) {
	convey.Convey("Test checkTimestamp", t, func() {
		convey.Convey("when requestTimestamp is invalid", func() {
			requestTimestamp := "123456"
			res, err := checkTimestamp(requestTimestamp)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("when requestTimestamp parse int failed", func() {
			requestTimestamp := "timestamp=aaaaa"
			res, err := checkTimestamp(requestTimestamp)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("when requestTimestamp expired", func() {
			timestamp := time.Now().UnixMilli() + 2*signExpirationTime
			requestTimestamp := "timestamp=" + strconv.FormatInt(timestamp, base)
			res, err := checkTimestamp(requestTimestamp)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("when requestTimestamp is too early", func() {
			timestamp := time.Now().UnixMilli() - signExpirationTime - 1
			requestTimestamp := "timestamp=" + strconv.FormatInt(timestamp, base)
			res, err := checkTimestamp(requestTimestamp)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("when requestTimestamp is correct", func() {
			requestTimestamp := "timestamp=" + strconv.FormatInt(time.Now().UnixMilli(), base)
			res, err := checkTimestamp(requestTimestamp)
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldNotBeEmpty)
		})
	})
}

func Test_checkSignField(t *testing.T) {
	convey.Convey("Test checkSignField", t, func() {
		convey.Convey("when originalStr is invalid", func() {
			err := checkSignField("test", "123456", "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when originalStr is different", func() {
			err := checkSignField("test", "ak=123456", "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when originalStr is correct", func() {
			err := checkSignField("test", "ak=123456", "123456")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_getHeaders(t *testing.T) {
	convey.Convey("Test getHeaders", t, func() {
		convey.Convey("when signedHeader is empty", func() {
			req := &fasthttp.Request{}
			headers, signedHeaderKeys := getHeaders(req)
			convey.So(headers, convey.ShouldBeNil)
			convey.So(signedHeaderKeys, convey.ShouldBeNil)
		})
		convey.Convey("when signedHeader key is not exist", func() {
			req := &fasthttp.Request{}
			req.Header.Add(constant.HeaderSignedHeader, "header1;header2")
			headers, signedHeaderKeys := getHeaders(req)
			convey.So(len(headers), convey.ShouldEqual, 0)
			convey.So(len(signedHeaderKeys), convey.ShouldEqual, 0)
		})
		convey.Convey("when get header success", func() {
			req := &fasthttp.Request{}
			req.Header.Add(constant.HeaderSignedHeader, "header1;header2")
			req.Header.Add("header1", "value1")
			req.Header.Add("header2", "value2")
			req.Header.Add("header3", "value3")
			headers, signedHeaderKeys := getHeaders(req)
			convey.So(len(headers), convey.ShouldEqual, 2)
			convey.So(len(signedHeaderKeys), convey.ShouldEqual, 2)
		})
	})
}

func Test_getQuery(t *testing.T) {
	convey.Convey("Test getQuery", t, func() {
		convey.Convey("when query is empty", func() {
			req := &fasthttp.Request{}
			req.SetRequestURI("")
			res := getQuery(req)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("when get query success", func() {
			req := &fasthttp.Request{}
			req.SetRequestURI("http://127.0.0.1:8080/test?key2=value2&key1=value1&key3=value3")
			res := getQuery(req)
			convey.So(res, convey.ShouldEqual, "key1=value1&key2=value2&key3=value3")
		})
	})
}

func Test_sign(t *testing.T) {
	convey.Convey("Test sign", t, func() {
		convey.Convey("when sign success", func() {
			req := &SignRequest{
				Method: "GET",
				Path:   "/test/path1",
				Query:  "key1=value1&key2=value2",
				Body:   nil,
				Headers: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				SignedHeaderKeys: []string{
					"key1", "key2",
				},
				Timestamp: "1767616334465",
				AK:        "yuanrong",
				SK:        "C0ECCDF386D96B8BF90562BC4F25C7608A8AA8D84003F088FCDC5E8606314A81",
			}
			res := sign(req)
			convey.So(res, convey.ShouldEqual, "f23b3988d84003482f5a559bde7224b3f6c20f211fb5ebf37eb1c8612c02e24b")
		})
	})
}

func Test_buildCanonicalHeaders(t *testing.T) {
	convey.Convey("Test buildCanonicalHeaders", t, func() {
		convey.Convey("when signedHeaderKeys is empty", func() {
			headers := make(map[string]string)
			keys := make([]string, 0)
			res := buildCanonicalHeaders(headers, keys)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("when build headers success", func() {
			headers := make(map[string]string)
			headers["key1"] = "value1"
			headers["key2"] = "value2"
			headers[constant.HeaderConnection] = "test"
			keys := make([]string, 0)
			keys = append(keys, "key1")
			keys = append(keys, "key2")
			res := buildCanonicalHeaders(headers, keys)
			convey.So(res, convey.ShouldEqual, "key1:value1\nkey2:value2\n")
		})
	})
}

func Test_buildSignedHeadersString(t *testing.T) {
	convey.Convey("Test buildSignedHeadersString", t, func() {
		convey.Convey("when SignedHeadersString is empty", func() {
			keys := make([]string, 0)
			res := buildSignedHeadersString(keys)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("when build signedHeaders string success", func() {
			keys := make([]string, 0)
			keys = append(keys, "key1")
			keys = append(keys, "key2")
			keys = append(keys, constant.HeaderConnection)
			res := buildSignedHeadersString(keys)
			convey.So(res, convey.ShouldEqual, "key1;key2")
		})
	})
}

func Test_buildSignature(t *testing.T) {
	convey.Convey("Test buildSignature", t, func() {
		convey.Convey("when buildSignature success", func() {
			sysSK := "C0ECCDF386D96B8BF90562BC4F25C7608A8AA8D84003F088FCDC5E8606314A81"
			data := "hello world"
			res := buildSignature(sysSK, data)
			convey.So(res, convey.ShouldNotBeEmpty)
		})
	})
}
