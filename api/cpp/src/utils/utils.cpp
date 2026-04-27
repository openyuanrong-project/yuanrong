/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

#include "utils.h"

#include "src/dto/config.h"
#include "src/libruntime/err_type.h"
#include "src/utility/string_utility.h"
#include "yr/api/exception.h"

namespace YR {
const int TENANT_ID_INDEX = 3;
const int FUNCTION_NAME_INDEX = 5;
const int FUNCTION_VERSION_INDEX = 6;
const int URN_CUT_NUM = 7;

void ThrowIfTrue(bool condition, Libruntime::ErrorCode code, const std::string &msg)
{
    if (condition) {
        YRLOG_ERROR(msg);
        throw YR::Exception(code, msg);
    }
}

std::string ConvertFunctionUrnToId(const std::string &functionUrn)
{
    char sep = ':';
    const std::string splitStr{"/"};
    std::vector<std::string> result;
    if (!functionUrn.empty()) {
        YR::utility::Split(functionUrn, result, sep);
    } else if (!YR::Libruntime::Config::Instance().YRFUNCID().empty()) {
        YR::utility::Split(YR::Libruntime::Config::Instance().YRFUNCID(), result, sep);
    }
    if (result.size() != URN_CUT_NUM) {
        throw YR::Exception(static_cast<int>(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID),
                            static_cast<int>(YR::Libruntime::ModuleCode::RUNTIME),
                            "Failed to split functionUrn: split num " + std::to_string(result.size()) +
                                " is expected to be " + std::to_string(URN_CUT_NUM));
    }
    return result[TENANT_ID_INDEX] + splitStr + result[FUNCTION_NAME_INDEX] + splitStr + result[FUNCTION_VERSION_INDEX];
}

YR::Libruntime::ErrorInfo WriteDataObject(const void *data, std::shared_ptr<YR::Libruntime::DataObject> &dataObj,
                                          uint64_t size, const std::unordered_set<std::string> &nestedIds)
{
    // copy data to DataObject data
    auto err = dataObj->buffer->WriterLatch();
    if (!err.OK()) {
        return err;
    }
    (void)memset_s(dataObj->meta->MutableData(), dataObj->meta->GetSize(), 0, dataObj->meta->GetSize());
    err = dataObj->data->MemoryCopy(data, size);
    if (!err.OK()) {
        return err;
    }
    err = dataObj->buffer->Seal(nestedIds);
    if (!err.OK()) {
        return err;
    }
    err = dataObj->buffer->WriterUnlatch();
    return err;
}

std::string GetEnv(const std::string &key)
{
    if (const char *env = std::getenv(key.c_str())) {
        return std::string(env);
    }
    return std::string("");
}

bool SetEnv(const std::string &k, const std::string &v)
{
    const char *name = k.c_str();
    const char *value = v.c_str();
    int overwrite = 1;
    if (setenv(name, value, overwrite) != 0) {
        std::cerr << "Failed to set environment variable." << std::endl;
        return false;
    }
    return true;
}

}  // namespace YR
