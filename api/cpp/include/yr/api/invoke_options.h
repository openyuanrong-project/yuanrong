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
#include <iostream>
#include <unordered_map>
#include <vector>

#include "yr/api/affinity.h"
#include "yr/api/exception.h"
#include "yr/api/runtime_env.h"

namespace YR {
/*!
   @brief Configuration for resource binding affinity.
*/
struct BindOptions {
    /*!
       @brief Resource type to bind to. default type is ``None``.

       - ``'None'``: No binding.

       - ``'NUMA'``: Bind to NUMA nodes.
    */
    std::string resource = "NONE";
    /*!
       @brief Binding strategy for the specified resource, default strategy is ``None``.

       - ``'None'``: Equivalent to `Spread`.

       - ``'PACK'``: Pack multiple instances into the same resource units as much as possible.

       - ``'SPREAD'``: Distribute multiple instances across different resource units as much as possible.
    */
    std::string strategy = "NONE";
};

/*!
   @brief Configuration options for grouped instance scheduling.

   The `GroupOptions` structure defines parameters for the lifecycle management of grouped instances, including timeout
   settings for rescheduling when kernel resources are insufficient.
*/
struct GroupOptions {
    /*!
       @brief Timeout for rescheduling when kernel resources are insufficient, in seconds.

       - If set to `-1`, the kernel will retry scheduling indefinitely.
       - If set to a value less than `0`, an exception will be thrown.
    */
    int timeout = NO_TIMEOUT;

    /*!
       @brief Whether to enable the fate-sharing configuration for grouped instances.

       - `true` (default): Instances in the group will be created and destroyed together.
       - `false`: Instances can have independent lifecycles.
    */
    bool sameLifecycle = true;
    /*!
       @brief The strategy to create the group, default strategy is ``None``.
            
        - ``'None'``: No strategy.

        - ``'PACK'``: Pack multiple instances into the same node as much as possible.

        - ``'SPREAD'``: Distribute multiple instances across different nodes as much as possible.

        - ``'STRICT_PACK'``: All instances must be placed on the same node, otherwise creation fails.

        - ``'STRICT_SPREAD'``: All instances must be placed on different nodes, otherwise creation fails.
    */
    std::string strategy = "None";
    /**
     * @brief Resource binding options. Used to specify resource-level binding strategy.
     */
    BindOptions bind;
};

/**
 * @struct RangeOptions
 * @brief The RangeOptions struct is a parameter within the InstanceRange struct of the Eugeo Range scheduling
 * configuration. It defines lifecycle parameters for instance Range scheduling, including the total timeout for
 * stepwise scheduling when kernel resources are insufficient.
 * @warning
 * 1. A single Range can create a maximum of 256 instances in a group.
 * 2. Concurrent creation supports up to 12 groups, with each group capable of creating a maximum of 256 instances.
 * 3. Calling the `Invoke()` interface after `NamedInstance::Export()` will cause the current thread to hang.
 * 4. Making function requests directly to stateful function instances without calling `Invoke()` and then retrieving
 * results will cause the current thread to hang.
 * 5. Repeatedly calling `Invoke()` will result in an exception being thrown.
 * 6. Instances within a Range do not support specifying a detached lifecycle.
 */
struct RangeOptions {
    /**
     * @brief Total timeout for stepwise scheduling when Eugeo kernel resources are insufficient, in seconds. -1
     * indicates no timeout, and scheduling will continue until successful or all attempts fail. Any other value less
     * than 0 will throw an exception.
     */
    int timeout = NO_TIMEOUT;
};

/**
 * @struct InstanceRange
 * @brief The InstanceRange struct configures the range of function instance counts, used to schedule/deploy a range of
 * function instances.
 * @note Range scheduling is an atomic resource allocation strategy. Based on the user-configured `max` field, the
 * current scheduling instance count is set to `now = max`. It first attempts to schedule `now` instances. If
 * successful, it returns the scheduled instances. If unsuccessful, it decrements `now` by `step` (to the maximum of
 * `now - step` and `min`) and attempts to schedule `now` instances again. This process continues until scheduling
 * succeeds, times out, or all attempts fail (considered failed if scheduling still fails after stepping down to `min`).
 */
struct InstanceRange {
    /**
     * @brief The minimum allowed number of function instances, default is -1.
     */
    int min = DEFAULT_INSTANCE_RANGE_NUM;
    /**
     * @brief The maximum allowed number of function instances, default is -1. Both `min` and `max` default to -1. When
     * both are -1, Range scheduling is disabled. When `1 <= min <= max`, Range scheduling is enabled. Any other values
     * are considered invalid and will throw an exception.
     */
    int max = DEFAULT_INSTANCE_RANGE_NUM;
    /**
     * @brief The step size for decrementing `max` towards `min`, default is 2. When Range scheduling is enabled, `step`
     * must be greater than 0; otherwise, an exception will be thrown. If `step` is larger than `max - min`, it will
     * step directly from `max` to `min`.
     */
    int step = DEFAULT_INSTANCE_RANGE_STEP;
    /**
     * @brief Specifies whether a group of instances configured with Range scheduling share the same lifecycle. Default
     * value: true.
     */
    bool sameLifecycle = true;
    /**
     * @brief Defines lifecycle parameters for instance Range scheduling, including the total timeout for stepwise
     * scheduling when kernel resources are insufficient.
     */
    RangeOptions rangeOpts;
};

/**
 * @struct DebugConfig
 * @brief 结构体 DebugConfig 是所有 debug 相关的配置，用于指定初始化 debug 实例的配置。
 * @note 当指定 enable 为 true 时，runtime 进程会关闭心跳检测，且可以被远程 debug 工具连接调试
 */
struct DebugConfig {
    DebugConfig() = default;
    ~DebugConfig() = default;
    /**
     * @brief 指定是否初始化 debug 实例。当指定 enable 为 true 时，runtime 进程会关闭心跳检测，且可以被远
     * 程 debug 工具连接调试
     */
    bool enable = false;
};

/*!
 *  @struct InvokeOptions invoke_options.h "include/yr/api/invoke_options.h"
 *  @brief used to set the invoke options.
 */
struct InvokeOptions {
    /*!
     * @var int cpu
     * @brief The minimum number of CPU cores required for the instance. 1/1000 cpu core
     */
    int cpu = 500;

    /*!
     * @var int mem
     * @brief The minimum memory size required for the instance. 1MB
     */
    int memory = 500;

    /*!
     * @var std::unordered_map<std::string, uint64_t> customResources
     * @brief 指定用户自定义资源，例如gpu、npu等
     */
    std::unordered_map<std::string, float> customResources;

    /*!
     * @var std::string customExtensions
     * @brief 指定用户自定义的配置，如函数并发度。同时也可以作为 metrics 的用户自定义 tag，用于采集用户信息。
     */
    std::unordered_map<std::string, std::string> customExtensions;

    /*!
     * @var std::unordered_map<std::string, std::string> podLabels
     * @brief 指定实例所在的 pod 标签
     */
    std::unordered_map<std::string, std::string> podLabels;

    /*!
     * @var std::vector<std::string> labels
     * @brief 指定函数的标签，用于函数实例亲和性调度。
     */
    std::vector<std::string> labels;

    /*!
     * @var std::unordered_map<std::string, std::string> affinity
     * @brief The affinity of the instance. Deprecated, use scheduleAffinities instead.
     */
    std::unordered_map<std::string, std::string> affinity;

    /*!
     * @var std::list<Affinity> scheduleAffinities
     * @brief The schedule affinity of the instance.
     */
    std::list<Affinity> scheduleAffinities;

    /*!
     * @var bool requiredPriority
     * @brief 是否开启优先级调度，开启后，当传入多个强亲和条件，按顺序匹配打分，当传入多个强亲和条件都不
     * 满足时，调度失败
     */
    bool requiredPriority = false;
    /*!
     * @var bool preferredPriority
     * @brief 是否开启优先级调度，开启后，当传入多个弱亲和条件，按顺序匹配打分，出现一个满足的即可调度成
     * 功。仅在弱亲和生效
     */
    bool preferredPriority = true;
    /*!
     * @var bool preferredAntiOtherLabels
     * @brief 是否开启反亲和不可选资源，开启后，当传入多个弱亲和条件都不满足时，调度失败。仅在弱反亲和生效
     */
    bool preferredAntiOtherLabels = false;

    /*!
    *  @var size_t retyrTime
    *  @brief 定义调用请求的重试次数。
    *  @note 关于 retryTimes 和 retryChecker，无状态函数，以下由框架重试的错误码，不占用重试次数：
             1. ERR_RESOURCE_NOT_ENOUGH
             2. ERR_INSTANCE_NOT_FOUND
             3. ERR_INSTANCE_EXITED

             建议由用户决定是否重试的错误码：
             1. ERR_USER_FUNCTION_EXCEPTION
             2. ERR_REQUEST_BETWEEN_RUNTIME_BUS
             3. ERR_INNER_COMMUNICATION
             4. ERR_SHARED_MEMORY_LIMIT
             5. EDERR_OPERATE_DISK_FAILED
             6. ERR_INSUFFICIENT_DISK_SPACE

             retryTimes 和 retryChecker 暂时不支持有状态函数，否则会抛出异常。
    */
    size_t retryTimes = 0;

    /*!
    *  @var size_t maxRetryTime
    *  @brief 定义无限重试最大重试次数，默认为-1，表示无限重试。
     */
    int maxRetryTime = -1;

    /*!
     * @var std::function retryChecker
     * @brief 无状态函数的重试判断钩子，默认为空。当 retryTimes = 0 时，本参数不生效。
     */
    bool (*retryChecker)(const Exception &e) noexcept = nullptr;

    /*!
     * @var int priority
     * @brief 设置无状态函数优先级，默认为 0。
     */
    size_t priority = 0;

    /*!
     * @var bool alwaysLocalMode
     * @brief 指定 cluster mode 强制在本地使用多线程运行用户函数，在 local mode 模式下不生效。
     */
    bool alwaysLocalMode = false;

    /*!
     * @var std::string groupName
     * @brief 指定成组实例调度器 name。
     */
    std::string groupName;

    /*!
     * @var bool needOrder
     * @brief 设置实例请求是否有序，默认为 false。仅在并发度为 1 时生效。
     */
    bool needOrder = false;

    /*!
     * @var InstanceRange instanceRange
     * @brief 函数实例个数 Range 范围配置，用于配置调度/部署一组函数实例的范围（详见 @ref InstanceRange ）。
     */
    InstanceRange instanceRange;

    /*!
     * @var int recoverRetryTimes
     * @brief 实例恢复次数,当实例异常退出时，自动用当前最新状态恢复实例。
     */
    int recoverRetryTimes = 0;

    /*!
     * @var std::unordered_map<std::string, std::string> envVars
     * @brief 设置实例启动时的环境变量。
     */
    std::unordered_map<std::string, std::string> envVars;

    DebugConfig debug;

    /*!
     * @var std::sring traceId
     * @brief 设置函数调用的 traceId，用于链路追踪。
     */
    std::string traceId = "";

    /*!
     * @var int timeout
     * @brief 实例创建和函数调用超时时间
     */
    int timeout = -1;

    /*!
     * @var bool preemptedAllowed
     * @brief 实例是否可以被抢占，仅在优先级场景下(openYuanrong 部署的 maxPriority 配置项大于 0 的场景)生效。默认为 false
     */
    bool preemptedAllowed = false;

    /*!
     * @var int instancePriority
     * @brief 实例的优先级，数值越大优先级越高，高优先级的实例可以抢占低优先级且被配置为（preemptedAllowed = true）
     * 的实例。仅在优先级场景下(openYuanrong 部署的 maxPriority 配置项大于 0 的场景)生效。instancePriority 的最小值为 0，最
     * 大值为 openYuanrong 部署的 maxPriority 配置。默认为 0
     */
    int instancePriority = 0;

    /*!
     * @var size_t scheduleTimeoutMs
     * @brief 实例的调度超时时间，单位毫秒，取值[-1, int64_t 类型最大值]。默认为 30000
     */
    int64_t scheduleTimeoutMs = 30000;

    /*!
     * @var RuntimeEnv runtimeEnv
     * @brief 设置python的virtual env
     */
    RuntimeEnv runtimeEnv;

    void CheckOptionsValid()
    {
        // check retry time
        if (retryTimes > MAX_OPTIONS_RETRY_TIME) {
            std::string msg = "invalid opts retryTimes: " + std::to_string(retryTimes);
            throw YR::Exception::InvalidParamException(msg);
        }

        if (retryTimes == 0 && retryChecker != nullptr) {
            std::cerr << "InvokeOptions check: retry time is zero but retry checker is assigned; retry will not work"
                      << std::endl;
        }

        if (!((instanceRange.min <= instanceRange.max && instanceRange.min > 0) ||
              (instanceRange.min == DEFAULT_INSTANCE_RANGE_NUM && instanceRange.max == DEFAULT_INSTANCE_RANGE_NUM))) {
            std::string msg = "invalid opts instanceRange, min: " + std::to_string(instanceRange.min) +
                              ", max: " + std::to_string(instanceRange.max) +
                              ", please set the min and the max as follows: max = min = -1 or max >= min > 0.";
            throw YR::Exception::InvalidParamException(msg);
        } else if (instanceRange.min <= instanceRange.max && instanceRange.min > 0 && !groupName.empty()) {
            throw YR::Exception::InvalidParamException(
                "gang scheduling and range scheduling cannot be used at the same time, please select one scheduling to "
                "set.");
        }
    }

    InvokeOptions &AddAffinity(std::list<Affinity> affinities)
    {
        scheduleAffinities.insert(scheduleAffinities.end(), affinities.begin(), affinities.end());
        return *this;
    }

    InvokeOptions &AddAffinity(Affinity affinity)
    {
        scheduleAffinities.push_back(affinity);
        return *this;
    }
};

namespace internal {
enum class FunctionLanguage : int {
    FUNC_LANG_CPP,
    FUNC_LANG_PYTHON,
    FUNC_LANG_JAVA,
};

struct FuncMeta {
    std::string appName;
    std::string moduleName;
    std::string funcName;
    std::string funcUrn;
    std::string className;
    FunctionLanguage language;
    std::string name;
    std::string ns;
    bool isAsync;
    bool isGenerator;
};

}  // namespace internal
}  // namespace YR
