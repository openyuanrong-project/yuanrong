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

package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

// TestDomain2IP convert domain to ip
func TestDomain2IP(t *testing.T) {
	patches := [...]*gomonkey.Patches{
		gomonkey.ApplyFunc(net.LookupHost, func(_ string) ([]string, error) {
			return []string{"1.1.1.1"}, nil
		}),
	}
	defer func() {
		for idx := range patches {
			patches[idx].Reset()
		}
	}()

	type args struct {
		endpoint string
	}
	tests := []struct {
		args    args
		want    string
		wantErr bool
	}{
		{
			args{endpoint: "1.1.1.1:9000"},
			"1.1.1.1:9000",
			false,
		},
		{
			args{endpoint: "1.1.1.1"},
			"1.1.1.1",
			false,
		},
		{
			args{endpoint: "test:9000"},
			"1.1.1.1:9000",
			false,
		},
		{
			args{endpoint: "test"},
			"1.1.1.1",
			false,
		},
	}
	for _, tt := range tests {
		got, err := Domain2IP(tt.args.endpoint)
		if (err != nil) != tt.wantErr {
			t.Errorf("Domain2IP() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if got != tt.want {
			t.Errorf("Domain2IP() got = %v, want %v", got, tt.want)
		}
	}
}

func TestGenStateIDByKey(t *testing.T) {
	convey.Convey("Test gen stateID by UUID", t, func() {
		stateID := GenStateIDByKey("tenantID", "serviceID", "funcName", "")
		convey.So(stateID, convey.ShouldNotBeNil)
	})
	convey.Convey("Test gen stateID by params", t, func() {
		stateID := GenStateIDByKey("tenantID", "serviceID", "funcName", "key")
		convey.So(stateID, convey.ShouldEqual, "993e96b4-0550-523f-a412-a4b58682cb2e")
	})
}

func TestGetFileHashInfo(t *testing.T) {
	convey.Convey("Test get file hashInfo failed", t, func() {
		_, _, err := GetFileHashInfo("/xyz")
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestFloat64ToByte(t *testing.T) {
	value := 123.45
	bytesValue := Float64ToByte(value)
	if ByteToFloat64(bytesValue) != value {
		t.Errorf("Float64ToByte and ByteToFloat64 failed")
	}
}

func TestExistPath(t *testing.T) {
	path := os.Args[0]
	if !ExistPath(path) {
		t.Errorf("test path exist true failed, path: %s", path)
	}
	if ExistPath(path + "abc") {
		t.Errorf("test path exist false failed, path: %s", path+"abc")
	}
}

func TestUniqueID(t *testing.T) {
	uuid1 := UniqueID()
	uuid2 := UniqueID()
	assert.NotEqual(t, uuid1, uuid2)
}

func Test_parseAzEnv(t *testing.T) {
	assert.Equal(t, constant.DefaultAZ, AzEnv())
	tests := []struct {
		name      string
		zoneValue string
		want      string
	}{
		{
			name:      "empty zoneValue",
			zoneValue: "",
			want:      constant.DefaultAZ,
		},
		{
			name: fmt.Sprintf("ZoneName > %d", constant.ZoneNameLen),
			zoneValue: "12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"123456",
			want: "12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"12345678901234567890123456789012345678901234567890" +
				"1234",
		},
		{
			name:      "Normal",
			zoneValue: "1234567890",
			want:      "1234567890",
		},
	}
	for _, tt := range tests {
		if err := os.Setenv(constant.ZoneKey, tt.zoneValue); err != nil {
			t.Errorf("failed to set Zone env, %s", err)
		}
		actual := parseAzEnv()
		assert.Equal(t, tt.want, actual)
	}
}

func TestIsConnRefusedErr(t *testing.T) {
	err := errors.New("abc")
	assert.False(t, IsConnRefusedErr(err))
	err = syscall.EADDRINUSE
	assert.False(t, IsConnRefusedErr(err))
	_, err = net.Dial("tcp", "127.0.0.1:33334")
	assert.True(t, IsConnRefusedErr(err))
}

func TestContainsConnRefusedErr(t *testing.T) {
	err := errors.New("dial tcp x.x.x.x:22668: connect: connection refused")
	assert.True(t, ContainsConnRefusedErr(err))
}

// TestWriteFileToPath is used to test the function of writing a file to a specified path.
func TestWriteFileToPath(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	assert.Equal(t, err, nil)
	addFile, err := ioutil.TempFile(dir, "test")
	err = WriteFileToPath(addFile.Name(), []byte("test"))
	assert.Equal(t, err, nil)
}

// TestIsHexString is used to test whether the character string meets the requirements.
func TestIsHexString(t *testing.T) {
	flag := IsHexString("2345")
	assert.True(t, flag)
	flag = IsHexString("test")
	assert.False(t, flag)
}

// TestValidateTimeout: indicates whether the timeout interval exceeds the maximum value or is the default value.
func TestValidateTimeout(t *testing.T) {
	var timeout int64 = -1
	var defaultTimeout int64 = 1
	ValidateTimeout(&timeout, defaultTimeout)
	assert.Equal(t, timeout, int64(1))
	timeout = 100*24*3600 + 1
	ValidateTimeout(&timeout, defaultTimeout)
	assert.Equal(t, timeout, int64(100*24*3600))
}

// TestDeepCopy is used to test the deep copy of maps and slices.
func TestDeepCopy(t *testing.T) {
	str := []string{"test1", "test2"}
	cpyStr := DeepCopy(str)
	curStr, ok := cpyStr.([]string)
	assert.True(t, ok)
	assert.Equal(t, len(curStr), 2)
	assert.Equal(t, curStr[0], "test1")
	assert.Equal(t, curStr[1], "test2")

	tmpMap := make(map[string]string)
	tmpMap["test1"] = "test1"
	tmpMap["test2"] = "test2"
	cpyMap := DeepCopy(tmpMap)
	curMap, ok := cpyMap.(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, len(curMap), 2)
	assert.Equal(t, curMap["test1"], "test1")
	assert.Equal(t, curMap["test2"], "test2")
}

func TestIsInputParameterValid(t *testing.T) {
	res1 := IsInputParameterValid("|")
	assert.Equal(t, res1, false)
	res2 := IsInputParameterValid("ddd")
	assert.Equal(t, res2, true)
	res3 := IsInputParameterValid("ab(d)e")
	assert.Equal(t, res3, false)
	res4 := IsInputParameterValid("abde;")
	assert.Equal(t, res4, false)
	res5 := IsInputParameterValid("&abde")
	assert.Equal(t, res5, false)
}

func TestDefaultString(t *testing.T) {
	convey.Convey("TestDefaultString", t, func() {
		convey.So(DefaultStringEnv("abc", "def"), convey.ShouldEqual, "def")
	})
}

func Test_replaceByDNS(t *testing.T) {
	convey.Convey("Test_replaceByDNSError", t, func() {
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc(ReadLines, func(path string) ([]string, error) {
				return nil, errors.New("mock error")
			}),
		}
		defer func() {
			for idx := range patches {
				patches[idx].Reset()
			}
		}()
		err := ReplaceByDNS("", nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("Test_replaceByDNS", t, func() {
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc(ReadLines, func(path string) ([]string, error) {
				return []string{"192.168.1.1 www.example.com"}, nil
			}),
			gomonkey.ApplyFunc(WriteLines, func(path string, lines []string) error {
				return nil
			}),
		}
		defer func() {
			for idx := range patches {
				patches[idx].Reset()
			}
		}()
		err := ReplaceByDNS("", map[string]string{"www.example.com": "192.168.1.2"})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("Test_replaceByDNSFileReWrite", t, func() {
		testLine := []string{"192.168.1.1 www.example.com"}
		err1 := WriteLines("/tmp/dnsTestFile", testLine)
		convey.So(err1, convey.ShouldBeNil)

		// 能够将第一次的文件内容覆盖
		testLine = []string{"192.168.1.1 www.example.com",
			"192.168.1.2 www.example2.com",
			"192.168.1.3 www.example3.com",
			"192.168.1.4 www.example4.com",
			"192.168.1.5 www.example5.com",
			"192.168.1.6 www.example6.com",
			"192.168.1.7 www.example7.com",
			"192.168.1.8 www.example8.com",
			"192.168.1.9 www.example9.com",
			"192.168.1.10 www.example10.com",
			"192.168.1.11 www.example11.com",
			"192.168.1.12 www.example12.com",
			"192.168.1.13 www.example13.com test-array-len-3-exception"}
		err1 = WriteLines("/tmp/dnsTestFile", testLine)
		lineContext, err1 := ReadLines("/tmp/dnsTestFile")
		convey.So(err1, convey.ShouldBeNil)
		convey.So(len(lineContext), convey.ShouldEqual, 13)
		var rst = 1
		for i := range lineContext {
			if strings.Contains(lineContext[i], "www.example.com") {
				if strings.Contains(lineContext[i], "192.168.1.1") {
					rst = 0
				}
			}
		}
		convey.So(rst, convey.ShouldEqual, 0)

		// 修改其中的一个条，其他内容不变
		err := ReplaceByDNS("/tmp/dnsTestFile", map[string]string{"www.example.com": "192.168.1.4"})
		err = ReplaceByDNS("/tmp/dnsTestFile", map[string]string{"www.example.com": "192.168.1.4"})
		convey.So(err, convey.ShouldBeNil)
		lineContext, err1 = ReadLines("/tmp/dnsTestFile")
		fmt.Println(lineContext)
		convey.So(err1, convey.ShouldBeNil)
		convey.So(len(lineContext), convey.ShouldEqual, 13)
		lineContext, err = ReadLines("/tmp/dnsTestFile")
		rst = 1
		for i := range lineContext {
			if strings.Contains(lineContext[i], "www.example.com") {
				if strings.Contains(lineContext[i], "192.168.1.4") {
					rst = 0
				}
			}
		}
		convey.So(rst, convey.ShouldEqual, 0)
		rst = 1
		for i := range lineContext {
			if strings.Contains(lineContext[i], "www.example2.com") {
				if strings.Contains(lineContext[i], "192.168.1.2") {
					rst = 0
				}
			}
		}
		convey.So(rst, convey.ShouldEqual, 0)

		// 新增一条,修改一条，一条不变，能够保持成功
		testLine = []string{"192.168.1.1 www.example.com",
			"192.168.1.2 www.example2.com",
			"192.168.1.14 www.example14.com"}
		err = ReplaceByDNS("/tmp/dnsTestFile", map[string]string{"www.example.com": "192.168.1.1",
			"www.example14.com": "192.168.1.14",
			"www.example2.com":  "192.168.1.2"})
		convey.So(err, convey.ShouldBeNil)
		lineContext, err1 = ReadLines("/tmp/dnsTestFile")
		convey.So(err1, convey.ShouldBeNil)
		convey.So(len(lineContext), convey.ShouldEqual, 14)
		lineContext, err = ReadLines("/tmp/dnsTestFile")
		rst = 1
		for i := range lineContext {
			if strings.Contains(lineContext[i], "www.example.com") {
				if strings.Contains(lineContext[i], "192.168.1.1") {
					rst = 0
				}
			}
		}
		convey.So(rst, convey.ShouldEqual, 0)
		rst = 1
		for i := range lineContext {
			if strings.Contains(lineContext[i], "www.example2.com") {
				if strings.Contains(lineContext[i], "192.168.1.2") {
					rst = 0
				}
			}
		}
		convey.So(rst, convey.ShouldEqual, 0)
		rst = 1
		for i := range lineContext {
			if strings.Contains(lineContext[i], "www.example14.com") {
				if strings.Contains(lineContext[i], "192.168.1.14") {
					rst = 0
				}
			}
		}
		convey.So(rst, convey.ShouldEqual, 0)
	})
}

func TestReadLines(t *testing.T) {
	convey.Convey("Test_replaceByDNSFileReWrite", t, func() {
		defer gomonkey.ApplyFunc(os.Open, func(name string) (*os.File, error) {
			return nil, fmt.Errorf("os.Open error")
		}).Reset()
		_, err := ReadLines("/test")
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestShortUUID(t *testing.T) {
	assert.Greater(t, 30, len(ShortUUID()))
}

func TestIsNetworkError(t *testing.T) {
	assert.Equal(t, false, IsNetworkError(nil))
	assert.Equal(t, true, IsNetworkError(syscall.EHOSTUNREACH))
	assert.Equal(t, true, IsNetworkError(os.ErrDeadlineExceeded))
	assert.Equal(t, false, IsNetworkError(errors.New("test error")))
}

func Test_IsUserError(t *testing.T) {
	flag := IsUserError(errors.New("test"))
	assert.Equal(t, false, flag)

	snErr := snerror.New(100, "test")
	flag = IsUserError(snErr)
	assert.Equal(t, false, flag)

	snErr = snerror.New(10500, "test")
	flag = IsUserError(snErr)
	assert.Equal(t, false, flag)

	snErr = snerror.New(4001, "test")
	flag = IsUserError(snErr)
	assert.Equal(t, true, flag)
}

func TestCalculateCPUByMemory(t *testing.T) {
	cpuInfo := CalculateCPUByMemory(10)
	assert.Equal(t, cpuInfo, 200)
}

func TestGenerateInstanceID(t *testing.T) {
	instanceID := GenerateInstanceID("podName")
	assert.Equal(t, instanceID, "defaultaz-#-podName")
}

func TestGetPodNameByInstanceID(t *testing.T) {
	podName := GetPodNameByInstanceID("defaultaz-#-podName")
	assert.Equal(t, podName, "podName")
}

func TestShuffleOneArray(t *testing.T) {
	arr1 := []string{"1"}
	arr2 := ShuffleOneArray(arr1)
	assert.Equal(t, arr1, arr2)

	arr3 := []string{"1", "2", "5", "6", "7"}
	arr4 := ShuffleOneArray(arr3)
	assert.NotEqual(t, arr3, arr4)

	arr5 := make([]string, 0)
	arr6 := ShuffleOneArray(arr5)
	assert.Equal(t, len(arr6), 0)
}

func TestIsCAEFunc(t *testing.T) {
	assert.Equal(t, true, IsCAEFunc(constant.BusinessTypeCAE))
	assert.Equal(t, false, IsCAEFunc(constant.WorkerManagerApplier))
}

func TestIsDirectFunc(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{
			name:     "python3.6",
			expected: false,
		},
		{
			name:     "java8",
			expected: false,
		},
		{
			name:     "javax",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsDirectFunc(tt.name), tt.name)
		})
	}
}

func TestIsNil(t *testing.T) {
	var obj *os.File
	assert.Equal(t, true, IsNil(obj))
	getObjFunc := func() interface{} {
		return obj
	}
	assert.Equal(t, true, IsNil(getObjFunc()))
}

func TestCalcFileMD5(t *testing.T) {
	os.Remove("./testFile")
	assert.Equal(t, "", CalcFileMD5("invalidPath"))
	os.WriteFile("./testFile",
		[]byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
	assert.Equal(t, "4fca8f1c736ca30135ed16538f4aebfc", CalcFileMD5("./testFile"))
	os.Remove("./testFile")
}

func TestFileMD5(t *testing.T) {
	os.Remove("./testFile")
	md5, err := FileMD5("invalidPath")
	assert.NotNil(t, err)
	assert.Equal(t, "", md5)
	os.WriteFile("./testFile",
		[]byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
	md5, err = FileMD5("./testFile")
	assert.Nil(t, err)
	assert.Equal(t, "4fca8f1c736ca30135ed16538f4aebfc", md5)
	os.Remove("./testFile")
}

func TestFnvHashInt(t *testing.T) {
	hashInt := FnvHashInt("123")
	assert.Equal(t, 1916298011, hashInt)
}

func TestSafeCloseStopCh(t *testing.T) {
	convey.Convey("stopCh", t, func() {
		stopCh := make(chan struct{}, 1)
		stopCh <- struct{}{}
		SafeCloseChannel(stopCh)
		_, ok := <-stopCh
		assert.Equal(t, false, ok)
	})
	convey.Convey("default", t, func() {
		stopCh := make(chan struct{}, 1)
		SafeCloseChannel(stopCh)
		_, ok := <-stopCh
		assert.Equal(t, false, ok)
	})
	convey.Convey("chan is nil", t, func() {
		SafeCloseChannel(nil)
	})
}

func TestMessageTruncation(t *testing.T) {
	message := "aaaaaaaaaaaaaaaaaaa"
	truncationMessage := MessageTruncation(message)
	assert.Equal(t, message, truncationMessage)
	rawMessage := ""
	for i := 0; i < 300; i++ {
		rawMessage = rawMessage + "a"
	}
	truncationMessage = MessageTruncation(rawMessage)
	assert.Equal(t, len(truncationMessage), 256)
}

func TestGetFunctionInstanceInfoFromEtcdKey(t *testing.T) {
	convey.Convey("Test GetFunctionInstanceInfoFromEtcdKey", t, func() {
		key := "/sn/instance/business/yrk/tenant/0/function/faasscheduler/version/latest/defaultaz/falseParam/requestID/3f079541-15fc-4009-8c41-50b2b2936772"
		_, err := GetFunctionInstanceInfoFromEtcdKey(key)
		convey.So(err, convey.ShouldNotBeNil)

		key = "/sn/instance/business/yrk/tenant/0/function/faasscheduler/version/latest/defaultaz/requestID/3f079541-15fc-4009-8c41-50b2b2936772"
		info, err := GetFunctionInstanceInfoFromEtcdKey(key)
		convey.So(err, convey.ShouldBeNil)
		convey.So(info.FunctionName, convey.ShouldEqual, "faasscheduler")
		convey.So(info.TenantID, convey.ShouldEqual, "0")
		convey.So(info.Version, convey.ShouldEqual, "latest")
		convey.So(info.InstanceName, convey.ShouldEqual, "3f079541-15fc-4009-8c41-50b2b2936772")

		key = "/sn/instance/business/yrk/tenant/0/function/faasscheduler/version/$latest/defaultaz/requestID/876a3352-44ea-4f0f-83b2-851c50aa89e1"
		info, err = GetFunctionInstanceInfoFromEtcdKey(key)
		convey.So(err, convey.ShouldBeNil)
		convey.So(info.FunctionName, convey.ShouldEqual, "faasscheduler")
		convey.So(info.TenantID, convey.ShouldEqual, "0")
		convey.So(info.Version, convey.ShouldEqual, "$latest")
		convey.So(info.InstanceName, convey.ShouldEqual, "876a3352-44ea-4f0f-83b2-851c50aa89e1")
	})
}

func TestGetModuleSchedulerInfoFromEtcdKey(t *testing.T) {
	convey.Convey("Test GetModuleSchedulerInfoFromEtcdKey", t, func() {
		key := "/sn/faas-scheduler/instances/cluster1/node1/falseParam/faas-scheduler-123"
		_, err := GetModuleSchedulerInfoFromEtcdKey(key)
		convey.So(err, convey.ShouldNotBeNil)

		key = "/sn/faas-scheduler/instances/cluster1/node1/faas-scheduler-123"
		info, err := GetModuleSchedulerInfoFromEtcdKey(key)
		convey.So(err, convey.ShouldBeNil)
		convey.So(info.FunctionName, convey.ShouldEqual, defaultFunctionName)
		convey.So(info.TenantID, convey.ShouldEqual, defaultTenant)
		convey.So(info.Version, convey.ShouldEqual, defaultVersion)
		convey.So(info.InstanceName, convey.ShouldEqual, "faas-scheduler-123")
	})
}

func TestCheckFaaSSchedulerInstanceFault(t *testing.T) {
	convey.Convey("Test CheckFaaSSchedulerInstanceFault", t, func() {
		testCases := []struct {
			name     string
			input    types.InstanceStatus
			expected bool
		}{
			{
				name:     "should return true for KernelInstanceStatusFatal",
				input:    types.InstanceStatus{Code: int32(constant.KernelInstanceStatusFatal)},
				expected: true,
			},
			{
				name:     "should return true for KernelInstanceStatusScheduleFailed",
				input:    types.InstanceStatus{Code: int32(constant.KernelInstanceStatusScheduleFailed)},
				expected: true,
			},
			{
				name:     "should return true for KernelInstanceStatusEvicting",
				input:    types.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicting)},
				expected: true,
			},
			{
				name:     "should return true for KernelInstanceStatusEvicted",
				input:    types.InstanceStatus{Code: int32(constant.KernelInstanceStatusEvicted)},
				expected: true,
			},
			{
				name:     "should return true for KernelInstanceStatusExiting",
				input:    types.InstanceStatus{Code: int32(constant.KernelInstanceStatusExiting)},
				expected: true,
			},
			{
				name:     "should return true for KernelInstanceStatusExited",
				input:    types.InstanceStatus{Code: int32(constant.KernelInstanceStatusExited)},
				expected: true,
			},
			{
				name:     "should return false for unknown status",
				input:    types.InstanceStatus{Code: 999},
				expected: false,
			},
		}

		for _, tc := range testCases {
			convey.Convey(tc.name, func() {
				result := CheckFaaSSchedulerInstanceFault(tc.input)
				convey.So(result, convey.ShouldEqual, tc.expected)
			})
		}
	})
}
