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

# 部署函数
function deploy() {
  # 检查必要的环境变量是否设置
  if [ -z "${SERVER_PORT}" ] || [ -z "${MODEL_PATH}" ]; then
    echo "错误：SERVER_PORT 或 MODEL_PATH 环境变量未设置！"
    exit 1
  fi

  # 执行部署命令并放入后台运行
  nohup python scripts.py deploy \
    --prefill-instances-num=${PREFILL_INS_NUM} \
    --decode-instances-num=${DECODE_INS_NUM} \
    -ptp=${PTP} \
    -dtp=${DTP} \
    -pdp=${PDP} \
    -ddp=${DDP} \
    --proxy-host=${SERVER_IP} \
    --proxy-port=${SERVER_PORT} \
    --prefill-startup-params "vllm serve --model=${MODEL_PATH} --trust-remote-code --max-model-len=2048 --gpu-memory-utilization 0.9" \
    --decode-startup-params "vllm serve --model=${MODEL_PATH} --trust-remote-code --max-model-len=2048 --gpu-memory-utilization 0.9" \
    >deploy.log 2>&1 &

  # 记录进程ID
  echo $! >deploy.pid
  echo "部署已启动，PID: $!, 日志写入 deploy.log"
}

# 清理函数
function clean() {
  # 执行清理命令
  python scripts.py clean

  # 杀死后台部署进程
  if [ -f "deploy.pid" ]; then
    kill -9 $(cat deploy.pid) 2>/dev/null
    rm -f deploy.pid
    echo "已终止部署进程"
  else
    echo "没有找到运行的部署进程"
  fi

  # 清理日志文件
  rm -f ds_c* *log 2>/dev/null
  echo "已清理日志文件"
}

# 主逻辑
case "$1" in
"deploy")
  deploy
  ;;
"clean")
  clean
  ;;
*)
  echo "使用方式: $0 [deploy|clean]"
  exit 1
  ;;
esac
