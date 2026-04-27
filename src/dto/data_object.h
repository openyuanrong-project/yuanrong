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

#pragma once

#include <msgpack.hpp>
#include <iostream>
#include <memory>

#include "buffer.h"
#include "src/libruntime/utils/serializer.h"
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
const uint64_t MetaDataLen = 16;
const uint64_t MetaDataTypeLen = 8;
struct DataObject {
    DataObject() : putDone(false) {}
    DataObject(const std::string &objId) : id(objId), putDone(false) {}
    DataObject(const std::string &objId, std::shared_ptr<Buffer> buf) : id(objId), buffer(buf), putDone(false)
    {
        this->SetBuffer(buf);
    }

    DataObject(uint64_t metaSize, uint64_t dataSize) : putDone(false)
    {
        if (metaSize == 0) {
            metaSize = MetaDataLen;
        }
        totalSize = metaSize + dataSize;
        buffer = std::make_shared<NativeBuffer>(totalSize);
        meta = std::make_shared<NativeBuffer>(buffer, 0, metaSize);
        data = std::make_shared<NativeBuffer>(buffer, metaSize, dataSize);
    }

    uint64_t GetSize() const
    {
        return totalSize;
    }

    void SetBuffer(std::shared_ptr<Buffer> buf)
    {
        if (!buf) {
            return;
        }
        totalSize = buf->GetSize();
        buffer = buf;
        if (totalSize >= MetaDataLen) {
            auto dataSize = totalSize - MetaDataLen;
            if (buf->IsNative()) {
                meta = std::make_shared<NativeBuffer>(buffer, 0, MetaDataLen);
                data = std::make_shared<NativeBuffer>(buffer, MetaDataLen, dataSize);
            } else {
                meta = std::make_shared<SharedBuffer>(buffer, 0, MetaDataLen);
                data = std::make_shared<SharedBuffer>(buffer, MetaDataLen, dataSize);
            }
        } else {
            YRLOG_WARN("unexpect total size {}", totalSize);
            data = buf;
            meta = std::make_shared<NativeBuffer>(MetaDataLen);
            (void)memset_s(meta->MutableData(), meta->GetSize(), 0, meta->GetSize());
        }
    }

    void SetDataBuf(std::shared_ptr<Buffer> dataBuf)
    {
        data = dataBuf;
    }
    void SetNestedIds(const std::vector<std::string> &ids)
    {
        nestedObjIds = ids;
    }

    void SetMetaDataType(const uint8_t &type)
    {
        msgpack::sbuffer metaDataType = Serializer::Serialize(type);
        if (metaDataType.size() <= MetaDataTypeLen) {
            meta->MemoryCopy(metaDataType.data(), metaDataType.size());
        } else {
            YRLOG_ERROR("unexpect metaDataType size {}", metaDataType.size());
        }
    }

    uint64_t totalSize = 0;
    std::string id;
    std::shared_ptr<Buffer> buffer;
    std::shared_ptr<Buffer> meta;
    std::shared_ptr<Buffer> data;
    std::vector<std::string> nestedObjIds;
    bool putDone;  // whether this obj has already Put to ds
    bool alwaysNative = false;
};
}  // namespace Libruntime
}  // namespace YR
