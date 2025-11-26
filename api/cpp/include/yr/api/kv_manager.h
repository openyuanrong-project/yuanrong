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

#include <cstring>
#include <iostream>

#include "constant.h"
#include "runtime_manager.h"
#include "yr/api/check_initialized.h"
#include "yr/api/exception.h"
#include "yr/api/local_mode_runtime.h"
#include "yr/api/serdes.h"

namespace YR {
class KVManager {
public:
    inline static KVManager &Singleton() noexcept
    {
        static KVManager kvManager;
        return kvManager;
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set char*
     *
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value Binary data to be stored. Cloud-out storage is limited to a maximum of 100M.
     * @param[in] existence Determines if the key allows repeated writes. Optional parameters are
     * `YR::ExistenceOpt::NONE` (allow, default) and `YR::ExistenceOpt::NX` (do not allow).
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const char *value, ExistenceOpt existence = ExistenceOpt::NONE)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf;
        if (value == nullptr) {
            throw Exception::InvalidParamException("KV Set value is nullptr");
        }
        SetParam setParam;
        setParam.existence = existence;
        sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(value, strlen(value));
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, existence)
                                       : internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set char* with len
     *
     * @details The value is of type char*, and the length can be specified. This interface is typically used when the
     * value contains a null character '\0', such as data serialized by msgpack.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value Binary data to be stored.
     * @param[in] len The length of the binary data. The user must ensure the correctness of the len value.
     * @param[in] existence Determines if the key allows repeated writes. Optional parameters are YR::ExistenceOpt::NONE
     * (allow, default) and `YR::ExistenceOpt::NX` (do not allow).
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const char *value, size_t len, ExistenceOpt existence = ExistenceOpt::NONE)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf;
        if (value == nullptr) {
            throw Exception::InvalidParamException("KV Set value is nullptr");
        }
        SetParam setParam;
        setParam.existence = existence;
        sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(value, len);
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, existence)
                                       : internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set string
     *
     * @details The value is of type string, and the length can be accurately determined by string.size(), so len is not
     * required.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value The string to be stored.
     * @param[in] existence Determines if the key allows repeated writes. Optional parameters are
     * `YR::ExistenceOpt::NONE` (allow, default) and `YR::ExistenceOpt::NX` (do not allow).
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const std::string &str, ExistenceOpt existence = ExistenceOpt::NONE)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(str.c_str(), str.length());
        SetParam setParam;
        setParam.existence = existence;
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, existence)
                                       : internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set char* with param
     *
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] val Binary data to be stored. Cloud-out storage is limited to a maximum of 100M.
     * @param[in] setParam Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const char *value, SetParam setParam)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf;
        if (value == nullptr) {
            throw Exception::InvalidParamException("KV Set value is nullptr");
        }
        sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(value, strlen(value));
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParam.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set char* with len and param
     *
     * @details The value is of type char*, and the length can be specified. This interface is typically used when the
     * value contains a null character '\0', such as data serialized by msgpack.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] val Binary data to be stored.
     * @param[in] len The length of the binary data. The user must ensure the correctness of the len value.
     * @param[in] setParam Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const char *value, size_t len, SetParam setParam)
    {
        CheckInitialized();
        if (value == nullptr) {
            throw Exception::InvalidParamException("KV Set value is nullptr");
        }
        std::shared_ptr<msgpack::sbuffer> sbuf;
        sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(value, len);
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParam.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set string with param
     *
     * @details The value is of type string, and the length can be accurately determined by string.size(), so len is not
     * required.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value The string to be stored.
     * @param[in] setParam Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const std::string &str, SetParam setParam)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(str.c_str(), str.length());
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParam.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set char* with param v2
     *
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] val Binary data to be stored. Cloud-out storage is limited to a maximum of 100M.
     * @param[in] setParamV2 Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const char *val, SetParamV2 setParamV2)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf;
        if (val == nullptr) {
            throw Exception::InvalidParamException("KV Set val is nullptr");
        }
        sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(val, strlen(val));
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParamV2.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParamV2);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set char* with len and param v2
     *
     * @details The value is of type char*, and the length can be specified. This interface is typically used when the
     * value contains a null character '\0', such as data serialized by msgpack.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] val Binary data to be stored.
     * @param[in] len The length of the binary data. The user must ensure the correctness of the len value.
     * @param[in] setParamV2 Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const char *val, size_t len, SetParamV2 setParamV2)
    {
        CheckInitialized();
        if (val == nullptr) {
            throw Exception::InvalidParamException("KV Set val is nullptr");
        }
        std::shared_ptr<msgpack::sbuffer> sbuf;
        sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(val, len);
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParamV2.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParamV2);
    }

    /**
     * @brief Sets the value of a key.
     *
     * @snippet{trimleft} kv_set_example.cpp set string with param v2
     *
     * @details The value is of type string, and the length can be accurately determined by string.size(), so len is not
     * required.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value The string to be stored.
     * @param[in] setParamV2 Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void Set(const std::string &key, const std::string &str, SetParamV2 setParamV2)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf = std::make_shared<msgpack::sbuffer>();
        sbuf->write(str.c_str(), str.length());
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParamV2.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParamV2);
    }

    /**
     * @brief A transactional interface for setting multiple binary data entries in a batch.
     *
     * This function provides a Redis-like `MSET` interface for storing multiple binary data entries in a single
     * transaction. It allows you to store binary data into the system using keys, ensuring that all operations are
     * atomic.
     *
     * @param keys A list of keys used to identify the data entries. These keys are used to query the data later. The
     * list cannot be empty, and its maximum length is 8.
     * @param vals A list of binary data to be stored. Each data entry corresponds to a key in the `keys` list. The
     * length of this list must match the length of `keys`.
     * @param lens A list of lengths corresponding to each binary data entry in `vals`. Each length must match the
     * corresponding data entry in `vals`. The length of this list must match the length of `keys`.
     * @param existence The existence option for the operation. Must be set to `YR::ExistenceOpt::NX` (not supported,
     * optional).
     *
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @note Maximum call frequency: 250 times per second.
     */
    static void MSetTx(const std::vector<std::string> &keys, const std::vector<char *> &vals,
                       const std::vector<size_t> &lens, ExistenceOpt existence)
    {
        CheckInitialized();
        if (keys.size() != vals.size() || keys.size() != lens.size()) {
            throw Exception::InvalidParamException("arguments vector size not equal.");
        }
        if (existence != ExistenceOpt::NX) {
            throw Exception::InvalidParamException("ExistenceOpt should be NX.");
        }
        std::vector<std::shared_ptr<msgpack::sbuffer>> sbufVec;
        sbufVec.resize(keys.size());
        for (size_t i = 0; i < keys.size(); i++) {
            std::shared_ptr<msgpack::sbuffer> tmpBufPtr = std::make_shared<msgpack::sbuffer>();
            tmpBufPtr->write(vals[i], lens[i]);
            sbufVec[i] = tmpBufPtr;
        }
        if (internal::IsLocalMode()) {
            internal::GetLocalModeRuntime()->KVMSetTx(keys, sbufVec, existence);
            return;
        }
        YR::internal::GetRuntime()->KVMSetTx(keys, sbufVec, existence);
    }

    /**
     * @brief A transactional interface for setting multiple key-value pairs in a batch.
     *
     * This function provides a Redis-like `MSET` interface for storing multiple key-value pairs in a single
     * transaction. It allows you to store binary data into the system using keys, ensuring that all operations are
     * atomic.
     *
     * @code
     * int main() {
     *     YR::Config conf;
     *     conf.mode = Config::Mode::CLUSTER_MODE;
     *     YR::Init(conf);
     *     std::vector<std::string> keys = { "key1", "key2" };
     *     std::vector<std::string> vals = { "val1", "val2" };
     *     try {
     *         YR::KV().MSetTx(keys, vals, YR::ExistenceOpt::NX);
     *         std::cout << "Data stored successfully." << std::endl;
     *     } catch (const YR::Exception& e) {
     *         std::cerr << "Error: " << e.what() << std::endl;
     *     }
     *     return 0;
     * }
     * @endcode
     *
     * @param keys A list of keys used to identify the data entries. These keys are used to query the data later.
     * The list cannot be empty, and its maximum length is 8.
     * @param vals A list of binary data to be stored. Each data entry corresponds to a key in the `keys` list.
     * The length of this list must match the length of `keys`.
     * @param existence The existence option for the operation. Must be set to `YR::ExistenceOpt::NX`.
     *
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @note Maximum call frequency: 250 times per second.
     */
    static void MSetTx(const std::vector<std::string> &keys, const std::vector<std::string> &vals,
                       ExistenceOpt existence)
    {
        CheckMSetTxParams(keys, vals, existence);
        std::vector<std::shared_ptr<msgpack::sbuffer>> sbufVec;
        sbufVec.resize(keys.size());
        for (size_t i = 0; i < keys.size(); i++) {
            std::shared_ptr<msgpack::sbuffer> tmpBufPtr = std::make_shared<msgpack::sbuffer>();
            tmpBufPtr->write(vals[i].c_str(), vals[i].size());
            sbufVec[i] = tmpBufPtr;
        }
        if (internal::IsLocalMode()) {
            internal::GetLocalModeRuntime()->KVMSetTx(keys, sbufVec, existence);
            return;
        }
        YR::internal::GetRuntime()->KVMSetTx(keys, sbufVec, existence);
    }

    /**
     * @brief A transactional interface for setting multiple key-value pairs with binary data in a batch.
     *
     * This function provides a Redis-like `MSET` interface for storing multiple key-value pairs in a single
     * transaction. It allows you to store binary data into the system using keys, ensuring that all operations are
     * atomic. The function also allows configuration of properties such as reliability through the `MSetParam`
     * parameter.
     *
     * @param keys A list of keys used to identify the data entries. These keys are used to query the data later. The
     * list cannot be empty, and its maximum length is 8.
     * @param vals A list of binary data to be stored. Each data entry corresponds to a key in the `keys` list. The
     * length of this list must match the length of `keys`.
     * @param lens A list of lengths corresponding to each binary data entry in `vals`. Each length must match the
     * corresponding data entry in `vals`. The length of this list must match the length of `keys`. The user must ensure
     * that the `len` values are correct.
     * @param mSetParam Configuration parameters for the operation, such as reliability settings.
     *
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @note Maximum call frequency: 250 times per second.
     */
    static void MSetTx(const std::vector<std::string> &keys, const std::vector<char *> &vals,
                       const std::vector<size_t> &lens, const MSetParam &mSetParam)
    {
        CheckInitialized();
        if (keys.size() != vals.size() || keys.size() != lens.size()) {
            throw Exception::InvalidParamException(
                "arguments vector size not equal. keys size is: " + std::to_string(keys.size()) +
                ", vals size is: " + std::to_string(vals.size()) + ", lens size is: " + std::to_string(lens.size()));
        }
        if (mSetParam.existence != ExistenceOpt::NX) {
            throw Exception::InvalidParamException("ExistenceOpt should be NX.");
        }
        std::vector<std::shared_ptr<msgpack::sbuffer>> sbufferVec;
        sbufferVec.resize(keys.size());
        for (size_t i = 0; i < keys.size(); i++) {
            std::shared_ptr<msgpack::sbuffer> tmpBufPtr = std::make_shared<msgpack::sbuffer>();
            tmpBufPtr->write(vals[i], lens[i]);
            sbufferVec[i] = tmpBufPtr;
        }
        if (internal::IsLocalMode()) {
            internal::GetLocalModeRuntime()->KVMSetTx(keys, sbufferVec, mSetParam.existence);
            return;
        }
        YR::internal::GetRuntime()->KVMSetTx(keys, sbufferVec, mSetParam);
    }

    /**
     * @brief A transactional interface for setting multiple key-value pairs with binary data in a batch.
     *
     * This function provides a Redis-like `MSET` interface for storing multiple key-value pairs in a single
     * transaction. It allows you to store binary data into the system using keys, ensuring that all operations are
     * atomic. The function also allows configuration of properties such as reliability through the `MSetParam`
     * parameter.
     *
     * @code
     * int main() {
     *     YR::Config conf;
     *     conf.mode = Config::Mode::CLUSTER_MODE;
     *     YR::Init(conf);
     *     std::vector<std::string> keys = { "key1", "key2" };
     *     std::vector<std::string> vals = { "val1", "val2" };
     *     YR::MSetParam param;
     *     param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
     *     param.ttlSecond = 10;
     *     param.cacheType = YR::CacheType::MEMORY;
     *     try {
     *         YR::KV().MSetTx(keys, vals, param);
     *         std::cout << "Data stored successfully." << std::endl;
     *     } catch (const YR::Exception& e) {
     *         std::cerr << "Error: " << e.what() << std::endl;
     *     }
     *     return 0;
     * }
     * @endcode
     *
     * @param keys A list of keys used to identify the data entries. These keys are used to query the data later. The
     * list cannot be empty, and its maximum length is 8.
     * @param vals A list of binary data to be stored. Each data entry corresponds to a key in the `keys` list. The
     * length of this list must match the length of `keys`.
     * @param mSetParam Configuration parameters for the operation, such as reliability settings.
     *
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @note Maximum call frequency: 250 times per second.
     */
    static void MSetTx(const std::vector<std::string> &keys, const std::vector<std::string> &vals,
                       const MSetParam &mSetParam)
    {
        CheckInitialized();
        if (keys.size() != vals.size()) {
            throw Exception::InvalidParamException(
                "input vector size not equal. keys size is: " + std::to_string(keys.size()) +
                ", vals size is: " + std::to_string(vals.size()));
        }
        if (mSetParam.existence != ExistenceOpt::NX) {
            throw Exception::InvalidParamException("ExistenceOpt should be NX.");
        }
        std::vector<std::shared_ptr<msgpack::sbuffer>> sbufVector;
        sbufVector.resize(keys.size());
        for (size_t i = 0; i < keys.size(); i++) {
            std::shared_ptr<msgpack::sbuffer> tmpBufPtr = std::make_shared<msgpack::sbuffer>();
            tmpBufPtr->write(vals[i].c_str(), vals[i].size());
            sbufVector[i] = tmpBufPtr;
        }
        if (internal::IsLocalMode()) {
            internal::GetLocalModeRuntime()->KVMSetTx(keys, sbufVector, mSetParam.existence);
            return;
        }
        YR::internal::GetRuntime()->KVMSetTx(keys, sbufVector, mSetParam);
    }

    /**
     * @brief Sets multiple key-value pairs. Values will be serialized by msgpack. All operations on key-value pairs
     * must either succeed completely or fail completely.
     * @tparam T The type of the object.
     * @param[in] keys An array of keys used to identify stored data. Maximum array length is ``8``. Each element must
     * match the regular expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] vals An array of data to be stored in the data system. The positions of data in this array must
     * correspond to the positions of keys in the keys array. The length of this array must match the keys array.
     * @param[in] existence Must be set to `YR::ExistenceOpt::NX`.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @note Maximum call frequency specification is 250 calls per second.
     *
     * @snippet{trimleft} kv_write_and_read_example.cpp multi writeTx objects
     */
    template <typename T>
    static void MWriteTx(const std::vector<std::string> &keys, const std::vector<T> &vals, ExistenceOpt existence)
    {
        CheckMSetTxParams(keys, vals, existence);
        std::vector<std::shared_ptr<msgpack::sbuffer>> sbufVec;
        sbufVec.resize(keys.size());

        for (size_t i = 0; i < keys.size(); i++) {
            std::shared_ptr<msgpack::sbuffer> tmpBufPtr =
                std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(vals[i]));
            sbufVec[i] = tmpBufPtr;
        }

        if (internal::IsLocalMode()) {
            internal::GetLocalModeRuntime()->KVMSetTx(keys, sbufVec, existence);
            return;
        }
        YR::internal::GetRuntime()->KVMSetTx(keys, sbufVec, existence);
    }

    /**
     * @brief Sets multiple key-value pairs. Values will be serialized by msgpack. All operations on key-value pairs
     * must either succeed completely or fail completely.
     * @tparam T The type of the object.
     * @param[in] keys An array of keys used to identify stored data. Maximum array length is 8. Each element must match
     * the regular expression: ^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$.
     * @param[in] vals An array of data to be stored in the data system. The positions of data in this array must
     * correspond to the positions of keys in the keys array. The length of this array must match the keys array.
     * @param[in] mSetParam Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @note Maximum call frequency specification is 250 calls per second.
     *
     * @snippet{trimleft} kv_write_and_read_example.cpp multi writeTx objects with param
     */
    template <typename T>
    static void MWriteTx(const std::vector<std::string> &keys, const std::vector<T> &vals, const MSetParam &mSetParam)
    {
        CheckInitialized();
        if (keys.size() != vals.size()) {
            throw Exception::InvalidParamException(
                "arguments vector size not equal. keys size is: " + std::to_string(keys.size()) +
                ", vals size is: " + std::to_string(vals.size()));
        }
        if (mSetParam.existence != ExistenceOpt::NX) {
            throw Exception::InvalidParamException("ExistenceOpt should be NX.");
        }
        std::vector<std::shared_ptr<msgpack::sbuffer>> sbufVec;
        sbufVec.resize(keys.size());

        for (size_t i = 0; i < keys.size(); i++) {
            std::shared_ptr<msgpack::sbuffer> tmpBufPtr =
                std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(vals[i]));
            sbufVec[i] = tmpBufPtr;
        }

        if (internal::IsLocalMode()) {
            internal::GetLocalModeRuntime()->KVMSetTx(keys, sbufVec, mSetParam.existence);
            return;
        }
        YR::internal::GetRuntime()->KVMSetTx(keys, sbufVec, mSetParam);
    }

    /**
     * @brief Writes the value of a key.
     * @tparam T The type of the object.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value Data to be stored in the data system.
     * @param[in] existence Determines if the key allows repeated writes. Optional parameters are YR::ExistenceOpt::NONE
     * (allow, default) and YR::ExistenceOpt::NX (do not allow).
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @snippet{trimleft} kv_write_and_read_example.cpp write objects
     */
    template <typename T>
    static void Write(const std::string &key, const T &value, ExistenceOpt existence = ExistenceOpt::NONE)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(value));
        SetParam setParam;
        setParam.existence = existence;
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Writes the value of a key.
     * @tparam T The type of the object.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value Data to be stored in the data system.
     * @param[in] setParam Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @snippet{trimleft} kv_write_and_read_example.cpp write objects with param
     */
    template <typename T>
    static void Write(const std::string &key, const T &value, SetParam setParam)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(value));
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParam.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Writes the value (raw bytes) of a key.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value Raw binary data to be stored in the data system.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     */
    static void WriteRaw(const std::string &key, const char *value)
    {
        CheckInitialized();
        SetParam setParam;
        return internal::IsLocalMode()
                   ? internal::GetLocalModeRuntime()->KVWrite(
                         key, std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(value)), setParam.existence)
                   : YR::internal::GetRuntime()->KVWrite(key, value, setParam);
    }

    /**
     * @brief Writes the value of a key.
     * @tparam T The type of the object.
     * @param[in] key A key used to identify the stored data. Must not be empty. Valid characters must match the regular
     * expression: ``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``.
     * @param[in] value Data to be stored in the data system.
     * @param[in] setParamV2 Configures attributes such as reliability for the object.
     * @throw Exception
     * - **1001**: Parameter error. Detailed error information will be provided.
     * - **4206**: Key already exists. If existence is set to `YR::ExistenceOpt::NX` and one or more keys in the list
     * have been previously set or written.
     * - Other exceptions may be thrown due to errors during the `Set` operation, with detailed descriptions provided in
     * the error message.
     *
     * @snippet{trimleft} kv_write_and_read_example.cpp write objects with param v2
     */
    template <typename T>
    static void Write(const std::string &key, const T &value, SetParamV2 setParam)
    {
        CheckInitialized();
        std::shared_ptr<msgpack::sbuffer> sbuf = std::make_shared<msgpack::sbuffer>(YR::internal::Serialize(value));
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVWrite(key, sbuf, setParam.existence)
                                       : YR::internal::GetRuntime()->KVWrite(key, sbuf, setParam);
    }

    /**
     * @brief Retrieves the value of a key.
     * @tparam T The type of the object.
     * @param[in] key The single key used to query the data.
     * @param[in] timeout Timeout in seconds, default is 300. Range [0, INT_MAX/1000). -1 indicates permanent blocking
     * wait.
     * @return Returns the retrieved data.
     *
     * @throw YR::Exception Thrown in the following cases:
     * - **1001**: Invalid input parameters (e.g., empty keys, mismatched sizes, or invalid characters).
     * - **4005**: Get operation failed (e.g., key not found or timeout exceeded).
     * - **4201**: RocksDB error (e.g., disk issues).
     * - **4202**: Shared memory limit exceeded.
     * - **4203**: Disk operation failed (e.g., permission issues).
     * - **4204**: Disk space full.
     * - **1000, 1001, 1002**: Internal communication errors.
     *
     * @warning When timeout is configured, the Read method will wait for the Write method to complete, up to the
     * timeout. There is no constraint on the order of Read and Write method calls. If a Write operation is followed by
     * an exception such as a ds worker restart, metadata residue must be handled to ensure the Write operation is
     * completed; otherwise, calling Read for the same key will throw an error without waiting for a timeout.
     *
     * @snippet{trimleft} kv_write_and_read_example.cpp read objects
     */
    template <typename T>
    static std::shared_ptr<T> Read(const std::string &key, int timeout = DEFAULT_GET_TIMEOUT_SEC)
    {
        CheckInitialized();
        auto buffer =
            internal::IsLocalMode()
                ? internal::GetLocalModeRuntime()->KVRead(key, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS)
                : internal::GetRuntime()->KVRead(key, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS);
        return YR::internal::Deserialize<std::shared_ptr<T>>(buffer);
    }

    /**
     * @brief Retrieves the value of a key written by WriteRaw.
     * @param[in] key The single key used to query the data.
     * @param[in] timeout Timeout in seconds, default is `300`. Range `[0, INT_MAX/1000)`. `-1` indicates permanent
     * blocking wait.
     * @return Returns the retrieved data.
     *
     * @throw YR::Exception Thrown in the following cases:
     * - **1001**: Invalid input parameters (e.g., empty keys, mismatched sizes, or invalid characters).
     * - **4005**: Get operation failed (e.g., key not found or timeout exceeded).
     * - **4201**: RocksDB error (e.g., disk issues).
     * - **4202**: Shared memory limit exceeded.
     * - **4203**: Disk operation failed (e.g., permission issues).
     * - **4204**: Disk space full.
     * - **1000, 1001, 1002**: Internal communication errors.
     * @warning When timeout is configured, the Read method will wait for the Write method to complete, up to the
     * timeout. There is no constraint on the order of Read and Write method calls. If a Write operation is followed by
     * an exception such as a ds worker restart, metadata residue must be handled to ensure the Write operation is
     * completed; otherwise, calling Read for the same key will throw an error without waiting for a timeout.
     */
    static const void *ReadRaw(const std::string &key, int timeout = DEFAULT_GET_TIMEOUT_SEC)
    {
        CheckInitialized();
        auto buffer =
            internal::IsLocalMode()
                ? internal::GetLocalModeRuntime()->KVRead(key, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS)
                : internal::GetRuntime()->KVRead(key, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS);
        return buffer->ImmutableData();
    }

    /**
     * @brief Retrieves the values of the keys.
     * @param[in] keys A collection of keys used to query the data. Maximum specification: 10000.
     * @param[in] timeout Timeout in seconds, default is 300. Range [0, INT_MAX).
     * @param[in] allowPartial Determines if partial success results are allowed. Default is false.
     * @return Returns the retrieved data. If a key does not exist, an exception is thrown.
     *
     * @throw YR::Exception Thrown in the following cases:
     * - **1001**: Invalid input parameters (e.g., empty keys, mismatched sizes, or invalid characters).
     * - **4005**: Get operation failed (e.g., key not found or timeout exceeded).
     * - **4201**: RocksDB error (e.g., disk issues).
     * - **4202**: Shared memory limit exceeded.
     * - **4203**: Disk operation failed (e.g., permission issues).
     * - **4204**: Disk space full.
     * - **1000, 1001, 1002**: Internal communication errors.
     *
     * @warning
     * - When allowPartial is set, the timeout parameter must be explicitly provided.
     * - When allowPartial is false: all keys must be successfully retrieved to return the results;
     * any failure will throw an exception.
     * - When allowPartial is true: returns results for any successful key retrieval, with failed keys having empty
     * results at their indices; if all keys fail, an exception is thrown.
     * @snippet{trimleft} kv_write_and_read_example.cpp read objects
     */
    template <typename T>
    static std::vector<std::shared_ptr<T>> Read(const std::vector<std::string> &keys,
                                                int timeout = DEFAULT_GET_TIMEOUT_SEC, bool allowPartial = false)
    {
        CheckInitialized();
        if (keys.empty()) {
            throw Exception::InvalidParamException("KVRead does not accept empty key list");
        }
        auto result = internal::IsLocalMode()
                          ? internal::GetLocalModeRuntime()->KVRead(
                                keys, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS, allowPartial)
                          : internal::GetRuntime()->KVRead(keys, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS,
                                                           allowPartial);

        std::vector<std::shared_ptr<T>> returnObjects;
        returnObjects.reserve(result.size());
        for (auto it = result.begin(); it != result.end(); ++it) {
            if (*it == nullptr) {
                returnObjects.push_back(nullptr);
                continue;
            }
            auto obj = YR::internal::Deserialize<std::shared_ptr<T>>(*it);
            returnObjects.push_back(std::move(obj));
        }
        return returnObjects;
    }

    /**
     * @brief Retrieves multiple values associated with specified keys with additional parameters, supporting
     * offset-based reading.
     *
     * This function fetches values stored under the specified keys, allowing for partial data retrieval based on
     * specified offsets and sizes. If any key does not exist or the operation times out, an exception is thrown unless
     * `allowPartial` is set to `true`.
     *
     * @param keys A list of keys to retrieve values for. The maximum number of keys is 10000.
     * @param params A structure containing parameters for the get operation, including offset, size, and traceId for
     * each key.
     * @param timeout The maximum time in seconds to wait for the values to be retrieved. The default value is 300
     * seconds. Setting to -1 waits indefinitely.
     *
     * @return std::vector<std::shared_ptr<YR::Buffer>>, A vector containing the retrieved data. The order of results
     * corresponds to the order of keys. Failed keys have empty pointers.
     *
     * @throw YR::Exception Thrown in the following cases:
     * - **1001**: Invalid input parameters (e.g., empty keys, mismatched sizes, or invalid characters).
     * - **4005**: Get operation failed (e.g., key not found or timeout exceeded).
     * - **4201**: RocksDB error (e.g., disk issues).
     * - **4202**: Shared memory limit exceeded.
     * - **4203**: Disk operation failed (e.g., permission issues).
     * - **4204**: Disk space full.
     * - **1000, 1001, 1002**: Internal communication errors.
     *
     * @code
     * int main() {
     *     YR::Config conf;
     *     conf.mode = Config::Mode::CLUSTER_MODE;
     *     YR::Init(conf);
     *     std::string key = "kv-id-888";
     *     YR::KV().Del(key);
     *     std::string value = "kv-id-888wqdq";
     *     YR::KV().Set(key, value);
     *     YR::GetParam param = { .offset = 0, .size = 0 };
     *     YR::GetParams params;
     *     params.getParams.emplace_back(param);
     *     std::vector<std::shared_ptr<YR::Buffer>> res = YR::KV().GetWithParam({ key }, params);
     *     return 0;
     * }
     * @endcode
     */
    static std::vector<std::shared_ptr<YR::Buffer>> GetWithParam(const std::vector<std::string> &keys,
                                                                 const YR::GetParams &params,
                                                                 int timeout = DEFAULT_GET_TIMEOUT_SEC)
    {
        CheckInitialized();
        if (params.getParams.empty()) {
            throw Exception::InvalidParamException("Get params does not accept empty key list");
        }
        if (params.getParams.size() != keys.size()) {
            throw Exception::InvalidParamException("Get params size is not equal to keys size");
        }
        auto results =
            internal::GetRuntime()->KVGetWithParam(keys, params, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS);
        return results;
    }

    /**
     * @brief Retrieves a value associated with a specified key, similar to Redis's GET command.
     *
     * This function fetches the value stored under the specified key. If the key does not exist, an exception is
     * thrown.
     *
     * @param key The key associated with the value to retrieve. The key must not be empty.
     * @param timeout The maximum time in seconds to wait for the value to be retrieved. If the value is not available
     * within this time, an exception is thrown. The default value is 300 seconds. The timeout can be set to -1 to wait
     * indefinitely.
     *
     * @return std::string The value associated with the specified key.
     *
     * @throw Exception Thrown in the following cases:
     * - **1001**: Invalid input parameters (e.g., empty keys or invalid characters).
     * - **4005**: Get operation failed (e.g., key not found or timeout exceeded).
     * - **4201**: RocksDB error (e.g., disk issues).
     * - **4202**: Shared memory limit exceeded.
     * - **4203**: Disk operation failed (e.g., permission issues).
     * - **4204**: Disk space full.
     * - **1000, 1001, 1002**: Internal communication errors.
     *
     */
    static std::string Get(const std::string &key, int timeout = DEFAULT_GET_TIMEOUT_SEC)
    {
        CheckInitialized();
        auto sbuf =
            internal::IsLocalMode()
                ? internal::GetLocalModeRuntime()->KVRead(key, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS)
                : internal::GetRuntime()->KVRead(key, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS);
        std::string retStr(static_cast<const char *>(sbuf->ImmutableData()), sbuf->GetSize());
        return retStr;
    }

    /**
     * @brief Retrieves multiple values associated with specified keys, similar to Redis's MGET command.
     *
     * This function fetches values stored under the specified keys. If any key does not exist or the operation times
     * out, an exception is thrown unless `allowPartial` is set to `true`.
     *
     * @param keys A list of keys to retrieve values for. The maximum number of keys is 10000.
     * @param timeout The maximum time in seconds to wait for the values to be retrieved. The default value is 300
     * seconds. Setting to -1 waits indefinitely.
     * @param allowPartial Whether to allow partial success results. If `true`, the function returns a vector of results
     * where failed keys have empty strings. If `false`, the function throws an exception if any key fails. Default is
     * `false`.
     *
     * @return std::vector<std::string> A vector containing the retrieved values. The order of values corresponds to the
     * order of keys.
     *
     * @throw Exception Thrown in the following cases:
     * - **1001**: Invalid input parameters (e.g., empty keys or invalid characters).
     * - **4005**: Get operation failed (e.g., key not found or timeout exceeded).
     * - **4201**: RocksDB error (e.g., disk issues).
     * - **4202**: Shared memory limit exceeded.
     * - **4203**: Disk operation failed (e.g., permission issues).
     * - **4204**: Disk space full.
     * - **1000, 1001, 1002**: Internal communication errors.
     *
     * @note
     * - When `allowPartial = false`: All keys must be retrieved successfully; otherwise, an exception is thrown.
     * - When `allowPartial = true`: Returns a vector of results where failed keys have empty strings. If all keys fail,
     * an exception is thrown.
     *
     * @code
     * int main() {
     *     YR::Config conf;
     *     conf.mode = Config::Mode::CLUSTER_MODE;
     *     YR::Init(conf);
     *     std::string value1{ "result" };
     *     std::string value2{ "result1" };
     *     YR::KV().Set("key", value1.c_str());
     *     std::string returnVal = YR::KV().Get("key"); // Retrieves "result"
     *
     *     std::vector<std::string> keys{ "key1", "key2" };
     *     YR::KV().Set(keys[0], value1);
     *     YR::KV().Set(keys[1], value2);
     *
     *     std::vector<std::string> returnVal = YR::KV().Get(keys); // Retrieves { "result", "result1" }
     *     return 0;
     * }
     * @endcode
     */
    static std::vector<std::string> Get(const std::vector<std::string> &keys, int timeout = DEFAULT_GET_TIMEOUT_SEC,
                                        bool allowPartial = false)
    {
        CheckInitialized();
        if (keys.empty()) {
            throw Exception::InvalidParamException("KVGet does not accept empty key list");
        }
        auto sbufVec = internal::IsLocalMode()
                           ? internal::GetLocalModeRuntime()->KVRead(
                                 keys, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS, allowPartial)
                           : internal::GetRuntime()->KVRead(keys, timeout == NO_TIMEOUT ? timeout : timeout * S_TO_MS,
                                                            allowPartial);
        std::vector<std::string> v;
        v.resize(keys.size());
        for (size_t i = 0; i < keys.size(); i++) {
            if (sbufVec[i] != nullptr) {
                v[i] = std::string(static_cast<const char *>(sbufVec[i]->ImmutableData()), sbufVec[i]->GetSize());
            }
        }
        return v;
    }

    /**
     * @brief Deletes a key and its associated data, similar to Redis's DEL command.
     *
     * This function removes the specified key and its associated data. If the key does not exist, the operation is
     * considered successful.
     *
     * @param key The key to delete. If the key does not exist, the operation is considered successful.
     * @param delParam Optional parameters for the delete operation, such as a custom trace ID.
     *
     * @throw Exception Thrown in cluster mode if the delete operation fails.
     *
     */
    static void Del(const std::string &key, const DelParam &delParam = {})
    {
        CheckInitialized();
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVDel(key)
                                       : internal::GetRuntime()->KVDel(key, delParam);
    }

    /**
     * @brief Deletes multiple keys and their associated data, similar to Redis's DEL command.
     *
     * This function removes the specified keys and their associated data. If a key does not exist, the operation for
     * that key is considered successful.
     *
     * @param keys A list of keys to delete. The maximum number of keys is 10000. If a key does not exist, the operation
     * for that key is considered successful.
     * @param delParam Optional parameters for the delete operation, such as a custom trace ID.
     *
     * @return std::vector<std::string>, A vector containing the keys that failed to be deleted. If all keys are
     * successfully deleted, the vector is empty.
     *
     * @throw YR::Exception Thrown in cluster mode if the connection to the data system times out during the delete
     * operation.
     *
     * @code
     * int main() {
     *     YR::Config conf;
     *     conf.mode = Config::Mode::CLUSTER_MODE;
     *     YR::Init(conf);
     *     std::vector<std::string> keys{ "key1", "key2" };
     *     std::vector<std::string> values{ "val1", "val2" };
     *     YR::KV().Set(keys[0], values[0]);
     *     YR::KV().Set(keys[1], values[1]);
     *
     *     std::vector<std::string> failedKeys = YR::KV().Del(keys);
     *     if (!failedKeys.empty()) {
     *         std::cout << "Failed to delete the following keys: ";
     *         for (const auto& key : failedKeys) {
     *             std::cout << key << " ";
     *         }
     *         std::cout << std::endl;
     *     } else {
     *         std::cout << "All keys deleted successfully." << std::endl;
     *     }
     *     return 0;
     * }
     * @endcode
     */
    static std::vector<std::string> Del(const std::vector<std::string> &keys, const DelParam &delParam = {})
    {
        CheckInitialized();
        return internal::IsLocalMode() ? internal::GetLocalModeRuntime()->KVDel(keys)
                                       : internal::GetRuntime()->KVDel(keys, delParam);
    }

    /**
     * @brief Check if multiple specified keys exist.
     *
     * @param keys A list of keys to check. The maximum number of keys is 10000。
     * @return std::vector<std::string> A vector containing the existence of the corresponding key. The order of values
     * corresponds to the order of keys.
     *
     * @throw Exception Thrown in the following cases:
     * - **1001**: Invalid input parameters (e.g., empty keys or invalid characters).
     * - **1002**: Internal communication errors.
     * - **4299**: DataSystem failed.
     */
    static std::vector<bool> Exist(const std::vector<std::string> &keys)
    {
        CheckInitialized();
        return internal::IsLocalMode()
                   ? internal::GetLocalModeRuntime()->KVExist(keys)
                   : internal::GetRuntime()->KVExist(keys);
    }

private:
    template <typename T>
    static void CheckMSetTxParams(const std::vector<std::string> &keys, const std::vector<T> &vals,
                                  ExistenceOpt existence)
    {
        CheckInitialized();
        if (keys.size() != vals.size()) {
            throw Exception::InvalidParamException("arguments vector size not equal.");
        }
        if (existence != ExistenceOpt::NX) {
            throw Exception::InvalidParamException("ExistenceOpt should be NX.");
        }
        return;
    }
};

}  // namespace YR