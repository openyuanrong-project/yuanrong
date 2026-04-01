# Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

BLOG_DATA = [
    {
        "title": "破解分布式计算多租困局：openYuanrong 多租隔离",
        "description": (
                       "openYuanrong 作为一个通用的 Serverless 分布式计算引擎，在构建之初就充分考虑了多租隔离。"
                       "支持多租户共享同一套集群，同时，通过 DaemonSet 形式部署 FunctionProxy 和 DataWorker 组件，"
                       "实现节点级别资源共享，有效加速租户工作负载的冷启动效率。"
        ),
        "date": "2026-03-04",
        "read_time": "8 分钟",
        "link": "https://www.openeuler.openatom.cn/zh/blog/20260316-openYuanrong_06/20260316-openYuanrong_06.html"
    },
    {
        "title": "openYuanrong 数据流：近计算高性能分布式数据流转",
        "description": (
                       "openYuanrong 数据流是 openYuanrong 数据系统提供的一个基于共享内存 Page 的流语义，"
                       "旨在消除传统消息中间件中的冗余拷贝与中转开销。 数据系统中包含三个组件："
                       "Client、Worker 和 Directory。"
        ),
        "date": "2026-02-28",
        "read_time": "10 分钟",
        "link": "https://www.openeuler.openatom.cn/zh/blog/20260316-openYuanrong_05/20260316-openYuanrong_05.html"
    },
    {
        "title": "YuanRong: A Production General-purpose Serverless System for Distributed Applications in the Cloud",
        "description": (
            "We design, implement, and evaluate YuanRong, the first production general-purpose serverless platform "
            "with a unified programming interface, multi-language runtime, and a distributed computing kernel for "
            "cloud-based applications."
        ),
        "date": "2024-08-04",
        "read_time": "30 分钟",
        "link": "https://dl.acm.org/doi/10.1145/3651890.3672216"
    },
]