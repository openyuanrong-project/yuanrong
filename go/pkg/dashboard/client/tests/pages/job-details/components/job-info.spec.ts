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

import { mount, flushPromises } from '@vue/test-utils';
import { describe, it, expect, vi } from 'vitest';
import { GetJobInfoAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import JobInfo from '@/pages/job-details/components/job-info.vue';

describe('JobInfo', () => {
    const wrapper = mount(JobInfo, {
        global: {
            stubs: {
                CommonCard: true,
                TinyCol: true,
                TinyLayout: true,
                TinyRow: true,
                AutoTip: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(CommonCard).exists()).toBe(true);
    });
})

describe('JobInfo.initJobInfo', () => {
    vi.mock('@/api/api', () => ({
        GetJobInfoAPI: vi.fn().mockResolvedValue({
            'key': '/sn/instance/business/yrk/tenant/default/function' +
                '/0-system-faasExecutorPosixCustom/version/$latest/defaultaz/0ebe2d84ad28eed000/' +
                'app-ab00977c-682e-4b5e-9cb3-f928c55a7d27',
            'type': 'SUBMISSION',
            'entrypoint': 'python3 ajobsample/three_actor.py',
            'submission_id': 'app-ab00977c-682e-4b5e-9cb3-f928c55a7d27',
            'driver_info': {
                'id': 'app-ab00977c-682e-4b5e-9cb3-f928c55a7d27',
                'node_ip_address': '127.0.0.1',
                'pid': '249351',
            },
            'status': 'FAILED',
            'start_time': '1762222864',
            'end_time': '1762222866',
            'metadata': null,
            'runtime_env': {
                'envVars': '{\'ENABLE_SERVER_MODE\':\'true\'}',
                'pip': '',
                'working_dir': 'file:///root/wxq/ajobsample/ajobsample.zip',
            },
            'driver_agent_http_address': '',
            'driver_node_id': 'dggphis232339-189755',
            'driver_exit_code': 139,
            'error_type': 'Instance(app-ab00977c-682e-4b5e-9cb3-f928c55a7d27) exitStatus:0',
        }),
    }));

    const wrapper = mount(JobInfo);
    const vm = wrapper.vm;
    it('renders when initJobInfo correctly', async () => {
        vm.initJobInfo();
        await flushPromises();
        expect(wrapper.text()).toContain('app-ab00977c-682e-4b5e-9cb3-f928c55a7d27FAILED');
    });
})