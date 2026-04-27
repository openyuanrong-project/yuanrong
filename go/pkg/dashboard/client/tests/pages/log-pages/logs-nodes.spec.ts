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

import { mount } from '@vue/test-utils';
import { describe, it, expect, vi } from 'vitest';
import CommonCard from '@/components/common-card.vue';
import LogsNodes from '@/pages/log-pages/logs-nodes.vue';

describe('LogsNodes', () => {
    const wrapper = mount(LogsNodes, {
        global: {
            stubs: {
                CommonCard: true,
                TinyLink: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(CommonCard).exists()).toBe(true);
    });
})

describe('LogsNodes.initScrollerItem', () => {
    vi.mock('@/api/api', () => {
        return {
            GetCompAPI: vi.fn().mockResolvedValue({
                'data': {
                    'nodes': [{
                        'hostname': 'dggphis232339-189755',
                        'status': 'healthy',
                        'address': '127.0.0.1',
                        'cap_cpu': 10000,
                        'cap_mem': 38912,
                        'alloc_cpu': 10000,
                        'alloc_mem': 38912,
                        'alloc_npu': 0
                    }]}}),
            ListLogsAPI: vi.fn().mockResolvedValue({
                'data': {
                    'dggphis232339-189755': [
                        'runtime-10050000-0000-4000-b00f-8374b8dd2508-00000000009d.out',
                        'runtime-10050000-0000-4000-b00f-8374b8dd2508-00000000009d.err',
                    ]}}),
        }
    })

    const wrapper = mount(LogsNodes);
    const vm = wrapper.vm;
    it('renders when initScrollerItem correctly', async () => {
        await vm.initNodesObj();
        await vm.initScrollerItem();
        expect(wrapper.text()).toBe('Logs Select a node to view logs:' +
            'dggphis232339-189755(IP:127.0.0.1)dggphis232339-189755(IP:127.0.0.1)');
    });
})
