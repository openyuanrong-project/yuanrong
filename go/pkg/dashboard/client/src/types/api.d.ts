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

interface RsrcSummData {
    cap_cpu: number;
    cap_mem: number;
    alloc_cpu: number;
    alloc_mem: number;
    proxy_num: number;
}

interface GetRsrcSummAPIRes {
    code: number;
    data: RsrcSummData;
    msg: string;
}

interface UsageInfo {
    cap_cpu: number;
    cap_mem: number;
    alloc_cpu: number;
    alloc_mem: number;
    alloc_npu: number;
}

interface CompInfo extends UsageInfo {
    hostname: string;
    status: string;
    address: string;
}

interface CompsObj {
    [key: string]: CompInfo[];
}

interface CompsData extends UsageInfo {
    nodes: CompInfo[];
    components: CompsObj;
}

interface GetCompAPIRes {
    code: number;
    data: CompsData;
    msg: string;
}

interface InstInfo {
    id: string;
    status: string;
    create_time: string;
    job_id: string;
    pid: string;
    ip: string;
    node_id: string;
    agent_id: string;
    parent_id: string;
    required_cpu: number;
    required_mem: number;
    required_gpu: number;
    required_npu: number;
    restarted: number;
    exit_detail: string;
}

interface GetInstAPIRes {
    code: number;
    data: InstInfo[];
    msg: string;
}

interface InstSummData {
    total: number;
    running: number;
    exited: number;
    fatal: number;
}

interface GetInstSummAPIRes {
    code: number;
    data: InstSummData;
    msg: string;
}

interface GetInstInstIDAPIRes {
    code: number;
    data: string;
    msg: string;
}

interface DriverDetail {
    id: string;
    node_ip_address: string;
    pid: string;
}

interface StringObj {
    [key: string]: string;
}

interface UnknownObj {
    [key: string]: unknown;
}

interface JobInfo {
    type: string;
    entrypoint: string;
    submission_id: string;
    driver_info: DriverDetail;
    status: string;
    message: string;
    error_type: string;
    start_time: string;
    end_time: string;
    metadata: object;
    runtime_env: UnknownObj;
    driver_agent_http_address: string;
    driver_node_id: string;
    driver_exit_code: number;
}

type GetListJobsAPIRes = JobInfo[]

interface ListLogsObj {
    [key: string]: string[];
}

interface ListLogsRes {
    message: string;
    data: ListLogsObj;
}

interface MetricData {
    metric: StringObj;
    value: [number, string];
}

type PromData = MetricData[]