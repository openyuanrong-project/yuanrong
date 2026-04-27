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

#pragma once
#include "yr/api/client_info.h"
#include "yr/api/config.h"

namespace YR {
/*!
  @brief YuanRong Init API, Configures runtime modes and system parameters.
  Refer to the data structure documentation for parameter specifications
  <a href="struct-Config.html">struct-Config </a>.
  @param conf YuanRong initialization parameter configuration. For parameter specifications, refer to
  <a href="struct-Config.html">struct-Config </a>.
  @return ClientInfo Refer to <a href="struct-Config.html">struct-Config </a>.

  @note When multi-tenancy is enabled on the YuanRong Cluster, users must configure a tenant ID.  For
  details on tenant ID configuration, refer to the "About Tenant ID" section in
  <a href="struct-Config.html">struct-Config </a>.

  @throws Exception The system will throw an exception when invalid config parameters are detected, such as an invalid
  mode type.

  @snippet{trimleft} init_and_finalize_example.cpp Init localMode

  @snippet{trimleft} init_and_finalize_example.cpp Init clusterMode
 */
ClientInfo Init(const Config &conf);

ClientInfo Init(const Config &conf, int argc, char **argv);

ClientInfo Init(int argc, char **argv);

/*!
  @brief Finalizes the Yuanrong system

  This function is responsible for releasing resources such as function instances
  and data objects that have been created during the execution of the program.
  It ensures that no resources are leaked, which could lead to issues in a
  production environment.

  @note - In a cluster deployment scenario, if worker processes exit and restart,
  it might lead to process residuals. In such cases, it is recommended to
  redeploy the cluster. Deployment scenarios like Donau or SGE can rely on
  the resource scheduling platform's capability to recycle processes.

  - This function should be called after the system has been initialized
  with the appropriate Init function. Calling Finalize before Init will result
  in an exception.

  @throws Exception If Finalize is called before the system is initialized,
  the exception "Please init YR first" will be thrown.

  @snippet{trimleft} init_and_finalize_example.cpp Init and Finalize
 */
void Finalize();
}  // namespace YR