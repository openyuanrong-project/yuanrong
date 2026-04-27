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

/*
Package libruntime
This package provides a set of functions to interface with the clibruntime.
*/
package libruntime

import (
	"yuanrong.org/kernel/runtime/libruntime/clibruntime"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

// Init Initialization entry, which is used to initialize the data system and function system.
func Init(conf config.Config) error {
	return clibruntime.Init(conf)
}

// ReceiveRequestLoop begins loop processing the received request.
func ReceiveRequestLoop() {
	clibruntime.ReceiveRequestLoop()
}

// ExecShutdownHandler exec shutdown handler.
func ExecShutdownHandler(signum int) {
	clibruntime.ExecShutdownHandler(signum)
}

// AllocReturnObject Creates an object and applies for a memory block.
// Computing operations can be performed on the memory block.
// will return a 'Buffer' that will be used to manipulate the memory
func AllocReturnObject(do *config.DataObject, size uint, nestedIds []string, totalNativeBufferSize *uint) error {
	return clibruntime.AllocReturnObject(do, size, nestedIds, totalNativeBufferSize)
}

// SetReturnObject if return by message, set return object
func SetReturnObject(do *config.DataObject, size uint) {
	clibruntime.SetReturnObject(do, size)
}

// WriterLatch Obtains the write lock of the buffer object.
func WriterLatch(do *config.DataObject) error {
	return clibruntime.WriterLatch(do)
}

// MemoryCopy Writes data to a buffer object.
func MemoryCopy(do *config.DataObject, src []byte) error {
	return clibruntime.MemoryCopy(do, src)
}

// Seal Publish the object and seal it. Sealed objects cannot be modified again.
func Seal(do *config.DataObject) error {
	return clibruntime.Seal(do)
}

// WriterUnlatch release the write lock of the buffer object.
func WriterUnlatch(do *config.DataObject) error {
	return clibruntime.WriterUnlatch(do)
}

// IsHealth -
func IsHealth() bool {
	return clibruntime.IsHealth()
}

// IsDsHealth -
func IsDsHealth() bool {
	return clibruntime.IsDsHealth()
}
