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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IsFileTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupSuite Setup Suite
func (suite *IsFileTestSuite) SetupSuite() {
	var err error

	// Create temp dir for IsFileTestSuite
	suite.tempDir, err = ioutil.TempDir("", "isfile-test")
	suite.Require().NoError(err)
}

// TearDownSuite TearDown Suite
func (suite *IsDirTestSuite) TearDownSuite() {
	defer os.RemoveAll(suite.tempDir)
}

// TestPositive Test Positive
func (suite *IsFileTestSuite) TestPositive() {

	// Create temp file
	tempFile, err := ioutil.TempFile(suite.tempDir, "temp_file")
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())

	// Verify that function isFile() returns true when file is created
	suite.Require().True(IsFile(tempFile.Name()))

}

// TestFileIsNotExist Test File Is Not Exist
func (suite *IsFileTestSuite) TestFileIsNotExist() {

	// Set path to unexisted file
	tempFile := filepath.Join(suite.tempDir, "somePath.txt")

	// Verify that function isFile() returns false when file doesn't exist in the system
	suite.Require().False(IsFile(tempFile))
}

// TestFileIsADirectory Test File Is A Directory
func (suite *IsFileTestSuite) TestFileIsADirectory() {
	suite.Require().False(IsFile(suite.tempDir))
}

type IsDirTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupSuite Setup Suite
func (suite *IsDirTestSuite) SetupSuite() {
	var err error

	// Create temp dir for IsDirTestSuite
	suite.tempDir, err = ioutil.TempDir("", "isdir-test")
	suite.Require().NoError(err)
}

// TearDownSuite TearDown Suite
func (suite *IsFileTestSuite) TearDownSuite() {
	defer os.RemoveAll(suite.tempDir)
}

// TestPositive Test Positive
func (suite *IsDirTestSuite) TestPositive() {

	// Verify that function IsDir() returns true when directory exists in the system
	suite.Require().True(IsDir(suite.tempDir))
}

// TestNegative Test Negative
func (suite *IsDirTestSuite) TestNegative() {

	// Create temp file
	tempFile, err := ioutil.TempFile(suite.tempDir, "temp_file")
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())

	// Verify that function IsDir( returns false when file instead of directory is function argument
	suite.Require().False(IsDir(tempFile.Name()))
}

type FileExistTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupSuite Setup Suite
func (suite *FileExistTestSuite) SetupSuite() {
	var err error

	// Create temp dir for FileExistTestSuite
	suite.tempDir, err = ioutil.TempDir("", "file_exists-test")
	suite.Require().NoError(err)
}

// TearDownSuite TearDown Suite
func (suite *FileExistTestSuite) TearDownSuite() {
	defer os.RemoveAll(suite.tempDir)
}

// TestPositive Test Positive
func (suite *FileExistTestSuite) TestPositive() {

	// Create temp file
	tempFile, err := ioutil.TempFile(suite.tempDir, "temp_file")
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())

	// Verify that function FileExists() returns true when file is exist
	suite.Require().True(FileExists(tempFile.Name()))
}

// TestFileNotExist Test File Not Exist
func (suite *FileExistTestSuite) TestFileNotExist() {

	// Set path to unexisted file
	tempFile := filepath.Join(suite.tempDir, "somePath.txt")

	// Verify that function FileExists() returns false when file doesn't exist
	suite.Require().False(FileExists(tempFile))
}

// TestFileIsNotAFile Test File Is Not A File
func (suite *FileExistTestSuite) TestFileIsNotAFile() {

	// Verify that function returns true when folder is exist in the system
	suite.Require().True(FileExists(suite.tempDir))
}

// TestHelperTestSuite Test Helper Test Suite
func TestHelperTestSuite(t *testing.T) {
	suite.Run(t, new(FileExistTestSuite))
	suite.Run(t, new(IsDirTestSuite))
	suite.Run(t, new(IsFileTestSuite))
}

func TestValidEnvValuePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"/home/sn/test", true},
		{"../../home/sn", false},
		{"/home/sn:/home/test:/opt", true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, ValidEnvValuePath(tt.input) == nil)
	}
}

func TestCopyDir(t *testing.T) {
	convey.Convey("CopyDir", t, func() {
		convey.Convey("CopyDir case 1", func() {
			srcPath, _ := ioutil.TempDir("", "src")
			dstPath, _ := ioutil.TempDir("", "dst")
			fileName := "fastfreeze.log"
			_, err := ioutil.TempFile(srcPath, fileName)
			err = CopyDir(srcPath, dstPath)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
