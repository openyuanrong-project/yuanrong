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
import { describe, it, expect, vi, afterAll } from 'vitest';
import { GetInstAPI, GetInstParentIDAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import InstancesChart from '@/pages/instances/instances-chart.vue';

describe('InstancesChart', () => {
    const wrapper = mount(InstancesChart, {
        global: {
            stubs: {
                CommonCard: true,
                TinyGrid: true,
                TinyGridColumn: true,
                TinyLink: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(CommonCard).exists()).toBe(true);
    });
})

describe('InstancesChart.initChart', () => {
    vi.mock('@/api/api', () => {
        const data = [
            {
                'id': '1025e641-f911-4500-8000-000000f8dbc2',
                'status': 'fatal',
                'create_time': '1762222864',
                'job_id': 'job-febb4a18',
                'pid': '249466',
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
            },
            {
                'id': 'app-ab00977c-682e-4b5e-9cb3-f928c55a7d27',
                'status': 'fatal',
                'create_time': '1762222864',
                'job_id': 'job-his232339-189755',
                'pid': '249351',
                'ip': '127.0.0.1:22773',
                'node_id': 'dggphis232339-189755',
                'agent_id': 'function-agent-127.0.0.1-58866',
                'parent_id': 'driver-faas-frontend-dggphis232339-189755',
                'required_cpu': 500,
                'required_mem': 500,
                'required_gpu': 0,
                'required_npu': 0,
                'restarted': 0,
                'exit_detail': 'Instance(app-ab00977c-682e-4b5e-9cb3-f928c55a7d27) exitStatus:0'
            },
        ];
        return {
            GetInstAPI: vi.fn().mockResolvedValue({data}),
            GetInstParentIDAPI: vi.fn().mockResolvedValue(data),
        }
    });

    const wrapper = mount(InstancesChart);
    const vm = wrapper.vm;

    it('renders when initChart.GetInstAPI correctly', async () => {
        vm.initChart();
        await flushPromises();
        expect(wrapper.text()).toContain('IDStatusJobIDPIDIPNodeIDParentIDCreateTimeRequired ' +
            'CPURequired Memory(MB)Required GPURequired NPURestartedExitDetailLog');
        expect(wrapper.text()).toContain('app-ab00977c-682e-4b5e-9cb3-f928c55a7d27fatal');
        expect(wrapper.text()).toContain('1025e641-f911-4500-8000-000000f8dbc2fatal');
    });

    const originalHash = window.location.hash;
    afterAll(()=>{ window.location.hash = originalHash });

    it('renders when initChart.GetInstParentIDAPI correctly', async () => {
        window.location.hash = '#/jobs/123';

        vm.initChart();
        await flushPromises();
        expect(wrapper.text()).toContain('app-ab00977c-682e-4b5e-9cb3-f928c55a7d27fatal');
        expect(wrapper.text()).toContain('1025e641-f911-4500-8000-000000f8dbc2fatal');
    });
})
