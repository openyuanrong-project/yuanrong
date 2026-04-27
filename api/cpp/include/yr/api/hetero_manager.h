/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
#include <future>
#include <vector>

#include "constant.h"
#include "runtime_manager.h"
#include "yr/api/check_initialized.h"
#include "yr/api/hetero_exception.h"
#include "yr/api/future.h"

namespace YR {
class HeteroManager {
public:

    /*!
      @brief Delete all device memory data corresponding to the specified object IDs.

      This function removes the device memory data associated with the provided object IDs.

      @param objectIds A vector of object IDs specifying the data to be deleted.
      @param failedObjectIds A vector to store the object IDs that failed to be deleted.
      @throw HeteroException If an error occurs during the deletion process, containing error code, error message, and
      failed keys.

      @code
      int main()
      {
          YR::Config conf;
          conf.mode = Config::Mode::CLUSTER_MODE;
          YR::Init(conf);
          std::vector<std::string> objectIds = {};
          std::vector<std::string> failedObjectIds = {};
          try {
              YR::HeteroManager.Delete(objectIds, failedObjectIds);
          } catch (YR::HeteroException &e) {
              // Handle exception...
          }
          return 0;
      }
      @endcode
    */
    static void DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds)
    {
        CheckInitialized();
        if (YR::internal::IsLocalMode()) {
            throw HeteroException::IncorrectFunctionUsageException("Delete is not supported in local mode");
        }
        YR::internal::GetRuntime()->DevDelete(objectIds, failedObjectIds);
    }

    /*!
      @brief Delete local device memory data corresponding to the specified object IDs.

      This function removes the local device memory data associated with the provided object IDs.

      @param objectIds A vector of object IDs specifying the data to be deleted.
      @param failedObjectIds A vector to store the object IDs that failed to be deleted.
      @throw HeteroException If an error occurs during the deletion process, containing error code, error message, and
      failed keys.

      @code
      int main()
      {
        YR::Config conf;
        conf.mode = Config::Mode::CLUSTER_MODE;
        YR::Init(conf);
        std::vector<std::string> objectIds = {};
        std::vector<std::string> failedObjectIds = {};
        try {
            YR::HeteroManager.LocalDelete(objectIds, failedObjectIds);
        } catch (YR::HeteroException &e) {
            // Handle exception...
        }
        return 0;
      }
      @endcode
    */
    static void DevLocalDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds)
    {
        CheckInitialized();
        if (YR::internal::IsLocalMode()) {
            throw HeteroException::IncorrectFunctionUsageException("LocalDelete is not supported in local mode");
        }
        YR::internal::GetRuntime()->DevLocalDelete(objectIds, failedObjectIds);
    }

    /*!
       @brief Subscribe to device memory data based on the specified keys and device memory structures.

       This function allows you to subscribe to device memory data by providing a list of keys and corresponding device
       memory structures. The subscription process is asynchronous, and the results are returned through a vector of
       Future objects.

       @param keys A vector of strings representing the keys used to identify the data to be subscribed. Must not be
       empty.
       @param blob2dList A vector of DeviceBlobList describing the device memory structures. The size must match that of
       keys.
       @param futureVec A vector of std::shared_ptr<YR::Future> objects representing the asynchronous subscription
       results. If an operation fails, an exception will be thrown.
       @throw HeteroException If an error occurs during the subscription process, containing error code, error
       message, and failed keys.

       @code
       int main()
       {
           YR::Config conf;
           conf.mode = Config::Mode::CLUSTER_MODE;
           YR::Init(conf);
           std::vector<std::string> keys = {};
           std::vector<YR::DeviceBlobList> blob2dList = {};
           std::vector<std::shared_ptr<YR::Future>> futureVec = {};
           try {
               YR::HeteroManager.DevSubscribe(keys, blob2dList, futureVec);
           } catch (YR::HeteroException &e) {
               // Handle exception...
           }
           return 0;
       }
       @endcode
    */
    static void DevSubscribe(const std::vector<std::string> &keys, const std::vector<YR::DeviceBlobList> &blob2dList,
                             std::vector<std::shared_ptr<YR::Future>> &futureVec)
    {
        CheckInitialized();
        if (YR::internal::IsLocalMode()) {
            throw HeteroException::IncorrectFunctionUsageException("DevSubscribe is not supported in local mode");
        }
        if (keys.size() != blob2dList.size()) {
            throw HeteroException::InvalidParamException(
                "The size of keys and blob2dList of DevSubscribe operation is inconsistent");
        }
        YR::internal::GetRuntime()->DevSubscribe(keys, blob2dList, futureVec);
    }

    /*!
      @brief Publish device memory data to the data system based on specified keys and device memory structures.

      This function allows you to publish device memory data to the data system by providing a list of keys and
      corresponding device memory structures. The data can be subscribed to using the DevSubscribe interface with
      matching keys. The publishing process is asynchronous, and the results are returned through a vector of Future
      objects.

      @param keys A vector of strings representing the keys used to identify the data to be published. Must not be
      empty.
      @param blob2dList A vector of DeviceBlobList describing the device memory structures. The size must match that of
      keys.
      @param futureVec A vector of std::shared_ptr<YR::Future> objects representing the asynchronous publishing results.
      If an operation fails, an exception will be thrown.
      @throw YR::HeteroException If an error occurs during the publishing process, containing error code, error message,
      and failed keys.

      @code
      int main()
      {
          YR::Config conf;
          conf.mode = Config::Mode::CLUSTER_MODE;
          YR::Init(conf);
          std::vector<std::string> keys = {};
          std::vector<YR::DeviceBlobList> blob2dList = {};
          std::vector<std::shared_ptr<YR::Future>> futureVec = {};
          try {
              YR::HeteroManager.DevPublish(keys, blob2dList, futureVec);
          } catch (YR::HeteroException &e) {
              // Handle exception...
          }
          return 0;
      }
      @endcode
    */
    static void DevPublish(const std::vector<std::string> &keys, const std::vector<YR::DeviceBlobList> &blob2dList,
                           std::vector<std::shared_ptr<YR::Future>> &futureVec)
    {
        CheckInitialized();
        if (YR::internal::IsLocalMode()) {
            throw HeteroException::IncorrectFunctionUsageException("DevPublish is not supported in local mode");
        }
        if (keys.size() != blob2dList.size()) {
            throw HeteroException::InvalidParamException(
                "The size of keys and blob2dList of DevPublish operation is inconsistent");
        }
        YR::internal::GetRuntime()->DevPublish(keys, blob2dList, futureVec);
    }

    /*!
       @brief Store device memory cache and save corresponding metadata to the data system.

       This function allows you to store device memory data using specified keys and device memory structures.
       The metadata associated with the stored data is also saved to the data system. The function returns a list
       of keys that failed to be stored.

       @param keys A vector of strings representing the keys used to identify the data to be stored. Must not be empty.
       @param blob2dList A vector of DeviceBlobList describing the device memory structures. The size must match that of
      keys.
       @param failedKeys A vector of strings that will store the keys which failed to be stored during the operation.
       @throw HeteroException If an error occurs during the storage process, containing error code, error message, and
      failed keys.

       @code
       int main()
       {
            YR::Config conf;
            conf.mode = Config::Mode::CLUSTER_MODE;
            YR::Init(conf);
            std::vector<std::string> keys = {};
            std::vector<YR::DeviceBlobList> blob2dList = {};
            std::vector<std::string> failedKeys = {};
            try {
                YR::HeteroManager.DevMSet(keys, blob2dList, failedKeys);
            } catch (YR::HeteroException &e) {
                // Handle exception...
            }
            return 0;
      }
      @endcode
    */
    static void DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                        std::vector<std::string> &failedKeys)
    {
        CheckInitialized();
        if (YR::internal::IsLocalMode()) {
            throw HeteroException::IncorrectFunctionUsageException("DevMSet is not supported in local mode");
        }
        if (keys.size() != blob2dList.size()) {
            throw HeteroException::InvalidParamException(
                "The size of keys and blob2dList of DevMSet operation is inconsistent");
        }
        YR::internal::GetRuntime()->DevMSet(keys, blob2dList, failedKeys);
    }

    /*!
      @brief Retrieve data from device memory and store it in the specified device memory structures.

      This function fetches data from device memory based on the provided keys and stores it into the device memory
      locations described by the DeviceBlobList structures. The operation is synchronous and may throw an exception if
      any errors occur.

      @param keys A vector of strings representing the keys used to identify the data to be retrieved. Must not be
      empty.
      @param blob2dList A vector of DeviceBlobList describing the device memory structures where the retrieved data will
      be stored. The size must match that of keys.
      @param failedKeys A vector of strings that will store the keys which failed to retrieve data during the operation.
      @param timeoutSec The timeout for the operation in seconds. Default is ``0``.
      @throw HeteroException If an error occurs during the retrieval process, containing error code, error message, and
      failed keys.

      @code
      int main()
      {
          YR::Config conf;
          conf.mode = Config::Mode::CLUSTER_MODE;
          YR::Init(conf);
          std::vector<std::string> keys = {};
          std::vector<YR::DeviceBlobList> blob2dList = {};
          std::vector<std::string> failedKeys = {};
          try {
              YR::HeteroManager.DevPreFetch(keys, blob2dList);
              YR::HeteroManager.DevMGet(keys, blob2dList, failedKeys, 10);
          } catch (YR::HeteroException &e) {
              // Handle exception...
          }
          return 0;
      }
      @endcode
    */
    static void DevMGet(const std::vector<std::string> &keys, const std::vector<YR::DeviceBlobList> &blob2dList,
                        std::vector<std::string> &failedKeys, int32_t timeoutSec)
    {
        CheckInitialized();
        if (YR::internal::IsLocalMode()) {
            throw HeteroException::IncorrectFunctionUsageException("DevMGet is not supported in local mode");
        }
        if (timeoutSec <= 0) {
            throw HeteroException::InvalidParamException("The timeoutSec value cannot be less than or equal to 0");
        }
        if (keys.size() != blob2dList.size()) {
            throw HeteroException::InvalidParamException(
                "The size of keys and blob2dList of DevMGet operation is inconsistent");
        }
        YR::internal::GetRuntime()->DevMGet(keys, blob2dList, failedKeys, timeoutSec);
    }
};

}  // namespace YR