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

#include <cstdint>

namespace YR {

/*!
 * @enum DataType
 * @brief Data type enumeration for collective communication operations.
 */
enum DataType : uint8_t {
    INT = 0,    ///< Integer type (int).
    DOUBLE,     ///< Double precision floating point type (double).
    LONG,       ///< Long integer type (long).
    FLOAT,      ///< Single precision floating point type (float).
    INVALID     ///< Invalid data type.
};

/*!
 * @var const std::unordered_map<DataType, int> DATA_TYPE_SIZE_MAP
 * @brief Mapping from DataType to its size in bytes.
 */
const static std::unordered_map<DataType, int> DATA_TYPE_SIZE_MAP = {{DataType::INT, sizeof(int)},
                                                                     {DataType::DOUBLE, sizeof(double)},
                                                                     {DataType::LONG, sizeof(long)},
                                                                     {DataType::FLOAT, sizeof(float)}};

/*!
 * @enum ReduceOp
 * @brief Reduction operator enumeration for collective communication operations.
 */
enum ReduceOp : uint8_t {
    SUM = 0,     ///< Sum operation.
    PRODUCT = 1, ///< Product operation.
    MIN = 2,     ///< Minimum operation.
    MAX = 3,     ///< Maximum operation.
};
}  // namespace YR
