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

import { BASE_URL, instance} from './index';

const apiV1Path = '/api/v1';
const request = instance;

export const GetRsrcSummAPI = (): Promise<GetRsrcSummAPIRes> =>
    request.get(apiV1Path + '/logical-resources/summary');

export const GetCompAPI = (): Promise<GetCompAPIRes> => request.get(apiV1Path + '/components');

export const GetInstAPI = (): Promise<GetInstAPIRes> => request.get(apiV1Path + '/instances');

export const GetInstSummAPI = (): Promise<GetInstSummAPIRes> => request.get(apiV1Path + '/instances/summary');

export const GetInstInstIDAPI = (instanceID: string): Promise<InstInfo> =>
    request.get(apiV1Path + '/instances/' + instanceID);

export const GetInstParentIDAPI = (parentID: string): Promise<InstInfo[]> =>
    request.get(apiV1Path + '/instances?parent_id=' + parentID);

const jobPath = '/api/jobs';

export const ListJobsAPI = (): Promise<GetListJobsAPIRes> => request.get(jobPath);

export const GetJobInfoAPI = (submissionID: string): Promise<JobInfo> =>
    request.get(jobPath + '/' + submissionID);

export const ListLogsAPI = (instanceID: string): Promise<ListLogsRes> =>
    request.get('/api/logs/list?instance_id=' + instanceID);

export const GetLogByFilenameAPI = (filename: string, start: number, end: number): Promise<string> =>
    request.get('/api/logs?filename=' + filename + '&start_line=' + start + '&end_line=' + end);

const urlSplit = BASE_URL.endsWith('/') ? '' : '/';
export const DownloadLogPath = BASE_URL + urlSplit + 'api/logs/download?filename=';

export const GetPromQueryAPI = (query: string): Promise<PromData> =>
    request.get('/api/v1/prometheus/query?query=' + query);