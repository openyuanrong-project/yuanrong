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
import { GetInstInstIDAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import InstanceInfo from '@/pages/instance-details/components/instance-info.vue';

describe('InstanceInfo', () => {
    const wrapper = mount(InstanceInfo, {
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

describe('InstanceInfo.initInstanceInfo', () => {
    vi.mock('@/api/api', () => ({
        GetInstInstIDAPI: vi.fn().mockResolvedValue({
            'id': '10050000-0000-4000-b00f-8374b8dd2508',
            'status': 'fatal',
            'create_time': '1762222864',
            'job_id': 'job-febb4a18',
            'pid': '249465',
            'ip': '127.0.0.1:22773',
            'node_id': 'dggphis232339-189755',
            'agent_id': 'function-agent-127.0.0.1-58866',
            'parent_id': 'app-ab00977c-682e-4b5e-9cb3-f928c55a7d27',
            'required_cpu': 3000,
            'required_mem': 500,
            'required_gpu': 0,
            'required_npu': 0,
            'restarted': 0,
            'exit_detail': 'ancestor instance(app-ab00977c-682e-4b5e-9cb3-f928c55a7d27) is abnormal'
        }),
    }));

    const wrapper = mount(InstanceInfo);
    const vm = wrapper.vm;
    it('renders when initInstanceInfo correctly', async () => {
        vm.initInstanceInfo();
        await flushPromises();
        expect(wrapper.text()).toContain('InstanceInfos IDStatusJobID10050000-0000-4000-b00f-8374b8dd2508fatal');
    });
})