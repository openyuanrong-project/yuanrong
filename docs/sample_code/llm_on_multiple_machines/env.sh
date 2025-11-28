#!/bin/bash
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

# configure by yourself
export SERVER_IP=127.0.0.1
export SERVER_PORT=9000

export PREFILL_INS_NUM=1
export DECODE_INS_NUM=1
export PTP=4
export DTP=2
export PDP=1
export DDP=1

# model
export MODEL_PATH="/workspace/models/qwen2.5_7B" # 模型路径
export PYTHONPATH=$PYTHONPATH:/workspace/tools/deploy # 代码路径

export NET_IFACE="enp67s0f5"

export VLLM_USE_V1=1                      # 启用 vLLM 的 v1 API 模式
export VLLM_WORKER_MULTIPROC_METHOD=spawn # Python 多进程启动方式为 spawn
export vLLM_MODEL_MEMORY_USE_GB=20        # 模型在单卡上执行需要的显存容量，Qwen2.5-7B设置20刚好合适
export PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION=python
export LD_LIBRARY_PATH=/usr/local/python3.11.13/lib/python3.11/site-packages/yr/inner/functionsystem/lib:$LD_LIBRARY_PATH # 需要找到yr的安装路径，可通过yr version 查看
export HCL_OP_EXPANSION_MODE="AIV"
export USING_PREFIX_CONNECTOR=1 # 是否开启前缀匹配，当前数据系统patch  默认开启， commitid：648c6a54b1c5896e85f3e136577deceda8551091
