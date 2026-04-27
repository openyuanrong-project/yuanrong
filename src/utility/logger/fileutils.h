/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#pragma once

#include <unistd.h>

#include <cstdint>
#include <string>
#include <vector>


namespace YR {
namespace utility {

constexpr char PATH_SEPARATOR = '/';

struct FileUnit {
    FileUnit(std::string name, size_t size) : name(std::move(name)), size(size) {}
    ~FileUnit() = default;

    // filename.
    std::string name;

    // file size.
    size_t size;
};

enum DIR_AUTH {
    DIR_AUTH_600 = 600,
    DIR_AUTH_700 = 700,
    DIR_AUTH_750 = 0750,
};

size_t FileSize(const std::string &filename);
bool FileExist(const std::string &filename, int mode = F_OK);
bool IsAbsolute(const std::string &filePath);
std::string GetCurrentPath();
void Glob(const std::string &pathPattern, std::vector<std::string> &paths);
void Read(FILE *f, uint8_t *buf, size_t *pSize);
int CompressFile(const std::string &src, const std::string &dest);
void DeleteFile(const std::string &filename);
void GetFileModifiedTime(const std::string &filename, int64_t &timestamp);
bool RenameFile(const std::string &srcFile, const std::string &targetFile) noexcept;
std::string GetRealPath(const std::string &inputPath);
bool ExistPath(const std::string &path);
bool Mkdir(const std::string &directory, const bool recursive = true, const DIR_AUTH dirAuth = DIR_AUTH_750);
std::uintmax_t Rm(const std::string &path);
std::vector<std::string> Tokenize(const std::string &s, const std::string &delims, const size_t maxTokens = 0);

}  // namespace utility
}  // namespace YR
