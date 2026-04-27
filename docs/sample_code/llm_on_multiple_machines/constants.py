#!/usr/bin/env python3
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VLLM_NAMESPACE = "vllm"
CONTROLLER_ACTOR_NAME = "Controller"
BALANCER_ACTOR_NAME = "Balancer"

ENDPOINT_APPLICATION_NAME = "vllm-pd-endpoint"
ENDPOINT_PROXY_DEPLOYMENT_NAME = "vllm-pd-endpoint"

VLLM_INSTANCE_HEALTH_CHECK_INTERVAL_S = 10

HTTP_OK = 200

HTTP_PARAM_INVALID = 400

HTTP_TOO_MANY_REQUESTS = 429

HTTP_INTERNAL_ERROR = 500

# The number of running requests on VLLM instances
NUM_REQUESTS_RUNNING = "vllm:num_requests_running"

# The number of waiting requests on VLLM instances
NUM_REQUESTS_WAITING = "vllm:num_requests_waiting"

# The usage of gpu cache on VLLM instances
GPU_CACHE_USAGE_PERC = "vllm:gpu_cache_usage_perc"

# Time Unit: s
METRICS_UPDATE_CYCLE = 0.5

INSTANCE_STATUS_CHECK_INTERVAL = 10

MIN_KV_PORT = 1024

MAX_KV_PORT = 65535

DEFAULT_KV_PORT = 14579

YR_LOG_LEVEL = "INFO"
