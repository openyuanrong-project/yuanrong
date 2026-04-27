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

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileExists(t *testing.T) {
	Convey("Given a temp file", t, func() {
		file, err := ioutil.TempFile("", "test-file")
		So(err, ShouldBeNil)
		filename := file.Name()

		Convey("When it is created", func() {
			Convey("Then it should return true", func() {
				So(FileExists(filename), ShouldBeTrue)
			})
		})

		Convey("When we delete the file", func() {
			err := file.Close()
			So(err, ShouldBeNil)
			err = os.Remove(filename)
			So(err, ShouldBeNil)

			Convey("Then it should return false", func() {
				So(FileExists(filename), ShouldBeFalse)
			})
		})
	})
}

func TestValidateFilePath(t *testing.T) {
	Convey("Given a abs file path and a rel file path", t, func() {
		relPath := "a/b"
		absPath, err := filepath.Abs(relPath)
		So(err, ShouldBeNil)

		Convey("The abs path should not return an error", func() {
			err = ValidateFilePath(absPath)
			So(err, ShouldBeNil)
		})

		Convey("The rel path should return an error", func() {
			err := ValidateFilePath(relPath)
			So(err, ShouldNotBeNil)
		})
	})
}
