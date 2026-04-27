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
#include "yr/api/invoke_options.h"
#include <string>
namespace YR {
/*!
   @brief A class for managing the lifecycle of grouped instances.

   The `Group` class is responsible for managing the lifecycle of grouped instances, including their creation and
   destruction. It follows the fate-sharing principle, where all instances in the group are created or destroyed
   together.

   The `Group` class provides methods to create, terminate, and manage grouped instances. It ensures that all instances
   in the group are treated as a single unit, and any failure during group creation will roll back the entire group.
*/
class Group {
public:
    /**
      @brief Default constructor.
    */
    Group() = default;

    /**
    * @brief Constructor.
    * @param name Instance Group Scheduling Class Name, the name must be unique.
    * @param opts See the structure GroupOptions.
    */
    Group(std::string &name, GroupOptions &opts);

    /**
    * @brief Constructor.
    * @param name Instance Group Scheduling Class Name, the name must be unique.
    * @param opts See the structure GroupOptions.
    */
    Group(std::string &name, GroupOptions &&opts);

    /*!
      @brief Execute the creation of a group of instances following the fate-sharing principle.

      This function is used to create a group of instances that share the same fate. All instances in the group
      will be created together, and if one instance fails, the entire group will be rolled back.
      The group configuration and options are defined in the `GroupOptions`.
      @code
      int main(void) {
        YR::Config conf;
        YR::Init(conf);
        std::string groupName = "";
        YR::InvokeOptions opts;
        opts.groupName = groupName;
        YR::GroupOptions groupOpts;
        groupOpts.timeout = 60;
        auto g = YR::Group(groupName, groupOpts);
        auto ins1 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
        auto ins2 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
        g.Invoke();
        return 0;
      }
      @endcode

      @note Constraints
      - A single group can create a maximum of 256 instances.
      - Concurrent creation supports a maximum of 12 groups, with each group creating up to 256 instances.
      - Calling this interface after `NamedInstance::Export()` will cause the current thread to hang.
      - If this interface is not called, directly invoking stateful function requests and retrieving results will cause
      the current thread to hang.
      - Repeated calls to this interface will result in an exception being thrown.
      - Instances in the group do not support specifying a detached lifecycle.

      @return void.

      @throw std::runtime_error If an error occurs during group creation or if the interface is misused.
   */
    void Invoke();

    /*!
        @brief Terminate a group of instances.

        This function is used to delete or terminate a group of instances that were created together.
        All instances in the group will be cleaned up or removed as a single unit, following the fate-sharing principle.

        @code
        int main(void) {
            YR::Config conf;
            YR::Init(conf);
            std::string groupName = "";
            YR::InvokeOptions opts;
            opts.groupName = groupName;
            YR::GroupOptions groupOpts;
            auto g = YR::Group(groupName, groupOpts);
            auto ins1 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
            auto ins2 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
            g.Invoke();
            g.Terminate();
            return 0;
        }
        @endcode

        @note Constraints
        - This function can only be called on a group that has been successfully created and invoked.
        - Calling this function multiple times on the same group will result in an exception being thrown.
        - If the group does not exist or has already been terminated, an exception will be thrown.
        - Termination will release all resources associated with the group and its instances.

        @return void.

        @throw std::runtime_error If the group does not exist, has already been terminated, or an error occurs during
       termination.
    */
    void Terminate();
     /*!
        @brief Waits for the completion of grouped instance creation and execution.

        This function blocks the current process until all instances in the group have completed their creation and
        execution. If the `Invoke` method has not been called before `Wait`, an exception is thrown.

        @code
        int main(void) {
            YR::Config conf;
            YR::Init(conf);
            std::string groupName = "";
            YR::InvokeOptions opts;
            opts.groupName = groupName;
            YR::GroupOptions groupOpts;
            auto g = YR::Group(groupName, groupOpts);
            auto ins1 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
            auto ins2 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
            g.Invoke();
            g.Wait();  // Wait for all instances in the group to complete
            return 0;
        }
        @endcode

        @return void.

        @throw Exception Thrown if the `Invoke` method has not been called before `Wait` or if the wait operation times
        out. The timeout duration is specified in the `GroupOptions` structure.
    */
    void Wait();

    /**
     * @brief Get the name of the group.
     *
     * @return The name of the group.
     */
    std::string GetGroupName() const;

private:
    /*!
       @brief The name of the group.

       The group name must be unique and cannot be duplicated across different groups.
    */
    std::string groupName;
    /*!
       @brief The configuration options for the group.

       This structure contains parameters such as timeouts and other settings that define the behavior of the group. For
       more details, refer to the `GroupOptions` documentation.
    */
    GroupOptions groupOpts;
};

}  // namespace YR