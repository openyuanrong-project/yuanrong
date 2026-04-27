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

#include <future>
#include <iostream>
#include <memory>
#include <msgpack.hpp>
#include <string>
#include "yr/api/check_initialized.h"
#include "yr/api/exception.h"
#include "yr/api/runtime_manager.h"
namespace YR {
namespace internal {
class LocalModeRuntime;
}

/**
 * @brief the object ref.
 *
 * @tparam T Type of referenced object.
 *
 * @note
    - openYuanRong encourages users to store large objects in the data system using YR::Put and obtain a unique
    ObjectRef (object reference). When invoking user functions, use the ObjectRef instead of the object itself as a
    function parameter to reduce the overhead of transmitting large objects between openYuanRong and user function
    components, ensuring efficient flow.
    - The return value of each user function call will also be returned in the form of an ObjectRef, which the user can
    use as an input parameter for the next call or retrieve the corresponding object through YR::Get.
    - Currently, users cannot construct ObjectRef by themselves.
 */
template <typename T>
class ObjectRef {
public:
    ObjectRef() = default;
    ObjectRef(const std::string &id, bool needIncre = true, bool isLocal = false) : objId(id), isLocal(isLocal)
    {
        CheckInitialized();
        if (internal::IsLocalMode()) {
            this->isLocal = true;
        } else if (needIncre) {
            YR::internal::GetRuntime()->IncreGlobalReference({id}, false);
        }
    }

    ObjectRef(const ObjectRef &rhs) : objId(rhs.objId), isLocal(rhs.isLocal)
    {
        CheckInitialized();
        if (isLocal) {
            future_ = rhs.future_;
        } else {
            YR::internal::GetRuntime()->IncreGlobalReference({rhs.objId}, false);
        }
    }

    ObjectRef(ObjectRef &&rhs)
    {
        *this = std::move(rhs);
    }

    ObjectRef &operator=(const ObjectRef &rhs)
    {
        CheckInitialized();
        if (this != &rhs) {
            isLocal = rhs.isLocal;
            if (isLocal) {
                future_ = rhs.future_;
            } else {
                YR::internal::GetRuntime()->IncreGlobalReference({rhs.objId}, false);
                if (!objId.empty()) {
                    YR::internal::GetRuntime()->DecreGlobalReference({objId});
                }
            }
            objId = rhs.objId;
        }
        return *this;
    }

    ObjectRef &operator=(ObjectRef &&rhs)
    {
        CheckInitialized();
        if (this != &rhs) {
            isLocal = rhs.isLocal;
            if (isLocal) {
                future_ = std::move(rhs.future_);
            } else {
                if (!objId.empty()) {
                    YR::internal::GetRuntime()->DecreGlobalReference({objId});
                }
            }
            objId = std::move(rhs.objId);
        }
        return *this;
    }

    ~ObjectRef()
    {
        // After using the move constructor, the old ObjectRef’s id is empty.
        if (!isLocal && !objId.empty() && IsInitialized()) {
            YR::internal::GetRuntime()->DecreGlobalReference({objId});
        }
    }

    /**
    * @brief Get objId.
    *
    * @return The ID of the objectRef.
    */
    inline std::string ID() const
    {
        return objId;
    }

    /**
    * @brief Check whether the objectRef is local.
    *
    * @return `true` if objectRef is local, `false` otherwise.
    */
    bool IsLocal() const
    {
        return isLocal;
    }

    friend class YR::internal::LocalModeRuntime;

private:
    template <typename T_ = T, typename = std::enable_if_t<!std::is_void<T_>::value>>
    inline void Put(const T_ &val)
    {
        std::promise<std::shared_ptr<T>> p;
        future_ = p.get_future();
        // shared_future can store state even if p is discarded
        p.set_value(std::make_shared<T>(val));
    }

    inline std::shared_ptr<T> Get(int timeout, bool allowPartial = false) const
    {
        if (!future_.valid()) {
            if (!allowPartial) {
                std::string msg = "Get: object " + objId + " does not exist";
                throw YR::Exception::GetException(msg);
            }
            std::cerr << "Get: object " << objId << " does not exist" << std::endl;
            return nullptr;
        }

        if (timeout != -1 && future_.wait_for(std::chrono::seconds(timeout)) != std::future_status::ready) {
            if (allowPartial) {
                return nullptr;
            }
            std::string msg = "WaitFor wait result timeout -- " + std::to_string(timeout);
            throw YR::Exception::GetException(msg);
        }

        try {
            return future_.get();
        } catch (YR::Exception &e) {
            throw e;
        } catch (std::exception &e) {
            throw YR::Exception::GetException("Get object failed, exception happens when executing user's function: " +
                                          std::string(e.what()));
        }
    }

    inline bool Wait(int timeout) const
    {
        if (!future_.valid()) {
            std::string msg = "Wait: object " + objId + " does not exist";
            throw YR::Exception::InnerSystemException(msg);
        }
        if (timeout == -1) {
            future_.get();  // get throws exception
            return true;
        }
        bool ready = (future_.wait_for(std::chrono::seconds(timeout)) == std::future_status::ready);
        if (ready) {
            future_.get();
        }
        return ready;
    }

    inline bool IsReady() const
    {
        if (!future_.valid()) {
            std::string msg = "IsReady: object " + objId + " does not exist";
            throw YR::Exception::InnerSystemException(msg);
        }
        bool ready = (future_.wait_for(std::chrono::seconds(0)) == std::future_status::ready);
        if (ready) {
            future_.get();
        }
        return ready;
    }

    inline void PutFuture(std::shared_future<std::shared_ptr<T>> &&fut)
    {
        future_ = std::move(fut);
    }

private:
    std::string objId;
    bool isLocal = false;
    mutable std::shared_future<std::shared_ptr<T>> future_;
};

template <typename T>
bool CheckRepeat(const std::vector<ObjectRef<T>> &objs)
{
    std::unordered_set<std::string> obj_ids;
    for (auto &obj : objs) {
        auto res = obj_ids.insert(obj.ID());
        if (!res.second) {
            return true;
        }
    }
    return false;
}
}  // namespace YR

// msgpack serialize, deserialize

extern thread_local std::unordered_set<std::string> localNestedObjList;
namespace msgpack {
using namespace YR;
MSGPACK_API_VERSION_NAMESPACE(MSGPACK_DEFAULT_API_NS)
{
    namespace adaptor {
    template <typename T>
    struct convert<YR::ObjectRef<T>> {
        msgpack::object const &operator()(msgpack::object const &o, YR::ObjectRef<T> &v) const
        {
            std::string objID;
            if (o.type == msgpack::type::STR) {
                objID.assign(o.via.str.ptr, o.via.str.size);
            } else if (o.type == msgpack::type::ARRAY) {
                objID = o.via.array.ptr[0].as<std::string>();
            } else {
                std::string msg = "invalid msgpack type " + std::to_string(o.type) + " for ObjectRef with type" +
                                  std::string(typeid(T).name());
                throw YR::Exception(msg);
            }
            v = YR::ObjectRef<T>(objID, true, false);
            localNestedObjList.insert(objID);
            return o;
        }
    };

    template <typename T>
    struct pack<YR::ObjectRef<T>> {
        template <typename Stream>
        packer<Stream> &operator()(msgpack::packer<Stream> &o, YR::ObjectRef<T> const &v) const
        {
            if (v.IsLocal()) {
                throw YR::Exception::InvalidParamException("cannot serialize local object ref");
            }
            o.pack_array(1);
            o.pack(v.ID());
            localNestedObjList.insert(v.ID());
            return o;
        }
    };
    }  // namespace adaptor
}  // MSGPACK_API_VERSION_NAMESPACE(MSGPACK_DEFAULT_API_NS)
}  // namespace msgpack