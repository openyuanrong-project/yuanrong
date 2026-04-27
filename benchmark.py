#!/usr/bin/env python3
# coding=UTF-8
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

import argparse
import json
import logging
import threading
import time
from concurrent.futures import ProcessPoolExecutor, ThreadPoolExecutor, wait
from datetime import datetime

import yr

logger = logging.getLogger(__name__)


# ========================
# 实例定义
# ========================
@yr.instance
class Counter:
    def __init__(self, init):
        self.cnt = init

    def add_one(self):
        self.cnt += 1
        return self.cnt

    def get(self):
        return self.cnt


def run_single_request(stats, termination_pool):
    """执行一次压测请求：仅测量 create + call 时间"""
    # 调用选项
    invoke_opts = yr.InvokeOptions()
    container_opts = {"runtime": "runsc", "imageurl": "xxxx.xxx.tag", "readonly": False}
    invoke_opts.custom_extensions["CONTAINER_OPTS"] = json.dumps(container_opts)
    start_time = time.time()
    ins = None
    try:
        # Step 1: 创建实例
        ins = Counter.options(invoke_opts).invoke(0)

        # Step 2: 调用方法（同步等待结果）
        _ = yr.get(ins.add_one.invoke())

        # ✅ 请求成功完成，记录延迟
        elapsed = time.time() - start_time
        stats["success"] += 1
        stats["latencies"].append(elapsed)

        # Step 3: 提交 terminate 到异步线程池（不阻塞主流程）
        if ins is not None:
            termination_pool.submit(safe_terminate, ins)

    except Exception:
        stats["failed"] += 1
    finally:
        stats["total"] += 1


def safe_terminate(instance):
    """安全销毁实例（防止异常中断压测）"""
    try:
        instance.terminate()
    except Exception:
        logger.debug("terminate failed in stress worker", exc_info=True)


def worker_loop(duration):
    """工作线程：持续发起请求"""
    stats = {"total": 0, "success": 0, "failed": 0, "latencies": []}
    termination_pool = ThreadPoolExecutor(max_workers=2)
    yr.init()
    end_time = time.time() + duration
    while time.time() < end_time:
        run_single_request(stats, termination_pool)
    termination_pool.shutdown(wait=True)
    yr.finalize()
    return stats


def main(concurrency, duration):
    print(
        f"🚀 Starting stress test (exclude terminate): concurrency={concurrency}, duration={duration}s"
    )

    # 统计结构

    request_pool = ProcessPoolExecutor(max_workers=concurrency)

    # 启动工作线程
    rets = []
    print(f"🧵 Spawning {concurrency} worker process...")
    for _ in range(concurrency):
        rets.append(request_pool.submit(worker_loop, duration))

    # 开始压测
    start_time = time.time()
    stats_fs = wait(rets)
    end_time = time.time()

    # 合并所有worker的统计数据
    merged_stats = {'total': 0, 'success': 0, 'failed': 0, 'latencies': []}
    for future in stats_fs.done:
        s = future.result()
        merged_stats['total'] += s['total']
        merged_stats['success'] += s['success']
        merged_stats['failed'] += s['failed']
        merged_stats['latencies'].extend(s['latencies'])

    # === 输出报告 ===
    total_time = end_time - start_time
    rps = merged_stats["total"] / total_time if total_time > 0 else 0

    print("\n" + "=" * 50)
    print("📊 STRESS TEST RESULT (exclude .terminate())")
    print("=" * 50)
    print(f"⏱️  Duration       : {total_time:.2f}s")
    print(f"🔁 Total Requests : {merged_stats['total']}")
    print(f"✅ Success        : {merged_stats['success']}")
    print(f"❌ Failed         : {merged_stats['failed']}")
    print(f"⚡ RPS            : {rps:.2f}")

    # 延迟统计（仅 create + call）
    if merged_stats["latencies"]:
        latencies_ms = sorted([lat * 1000 for lat in merged_stats["latencies"]])
        p50 = latencies_ms[len(latencies_ms) // 2]
        p90_idx = int(0.9 * len(latencies_ms))
        p99_idx = int(0.99 * len(latencies_ms))
        p90 = latencies_ms[p90_idx] if p90_idx < len(latencies_ms) else latencies_ms[-1]
        p99 = latencies_ms[p99_idx] if p99_idx < len(latencies_ms) else latencies_ms[-1]

        print(f"📉 Latency P50    : {p50:.2f} ms")
        print(f"📉 Latency P90    : {p90:.2f} ms")
        print(f"📉 Latency P99    : {p99:.2f} ms")

    # 清理
    request_pool.shutdown(wait=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="YR Stress Test: Only measure create + call time"
    )
    parser.add_argument(
        "--concurrency", type=int, default=10, help="Number of concurrent workers"
    )
    parser.add_argument("--duration", type=int, default=30, help="Duration in seconds")

    args = parser.parse_args()
    main(args.concurrency, args.duration)
