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
import CommonCard from '@/components/common-card.vue';
import ClusterChart from '@/pages/cluster/cluster-chart.vue';

describe('ClusterChart', () => {
    const wrapper = mount(ClusterChart, {
        global: {
            stubs: {
                CommonCard: true,
                TinyGrid: true,
                TinyGridColumn: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(CommonCard).exists()).toBe(true);
    });
})

describe('ClusterChart.initChartWithProm', () => {
    vi.mock('@/api/api', () => ({
        GetCompAPI: vi.fn().mockResolvedValue({
            data: {
                'nodes': [
                    {
                        'hostname': 'dggphis232340-2846342',
                        'status': 'healthy',
                        'address': '127.0.0.1',
                        'cap_cpu': 10000,
                        'cap_mem': 38912,
                        'alloc_cpu': 10000,
                        'alloc_mem': 38912,
                        'alloc_npu': 0
                    }
                ],
                'components': {
                    'dggphis232340-2846342': [
                        {
                            'hostname': 'function-agent-127.0.0.1-58866',
                            'status': 'alive',
                            'address': '127.0.0.1:58866',
                            'cap_cpu': 10000,
                            'cap_mem': 38912,
                            'alloc_cpu': 10000,
                            'alloc_mem': 38912,
                            'alloc_npu': 0
                        }
                    ]
                }
            },
        }),
        GetInstAPI: vi.fn().mockResolvedValue({
            data: [
                {
                    'id': 'app-6e334abe-3554-454f-94d0-493f24118292',
                    'status': 'fatal',
                    'create_time': '1762428525',
                    'job_id': 'job-his232340-2846342',
                    'pid': '2855872',
                    'ip': '127.0.0.1:22773',
                    'node_id': 'dggphis232340-2846342',
                    'agent_id': 'function-agent-127.0.0.1-58866',
                    'parent_id': 'driver-faas-frontend-dggphis232340-2846342',
                    'required_cpu': 500,
                    'required_mem': 500,
                    'required_gpu': 0,
                    'required_npu': 0,
                    'restarted': 0,
                    'exit_detail': 'Instance(app-6e334abe-3554-454f-94d0-493f24118292) exitStatus:1'
                }]}),
        GetPromQueryAPI: vi.fn().mockResolvedValue([]),
    }));
    const wrapper = mount(ClusterChart);

    it('renders initChartWithProm correctly', async () => {
        wrapper.vm.initChartWithProm();
        await flushPromises();
        expect(wrapper.text()).toContain('Components HostnameStatusIP/Address' +
            'CPUMemory(GB)NPUDisk(GB)Logical Resources');
        expect(wrapper.text()).toContain('dggphis232340-2846342healthy');
    });
})