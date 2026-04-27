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
import BreadcrumbComponent from '@/components/breadcrumb-component.vue';
import InstancesChart from '@/pages/instances/instances-chart.vue';
import JobInfo from '@/pages/job-details/components/job-info.vue';
import JobDetailsLayout from '@/pages/job-details/job-details-layout.vue';

describe('JobDetailsLayout', () => {
    const wrapper = mount(JobDetailsLayout, {
        global: {
            stubs: {
                BreadcrumbComponent: true,
                LogContentTemplate: true,
                InstancesChart: true,
                JobInfo: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(BreadcrumbComponent).exists()).toBe(true);
        expect(wrapper.findComponent(InstancesChart).exists()).toBe(true);
        expect(wrapper.findComponent(JobInfo).exists()).toBe(true);
    });
})

describe('JobDetailsLayout.initScrollerItem', () => {
    vi.mock('@/api/api', () => ({
        ListLogsAPI: vi.fn().mockResolvedValue({
            'message': '',
            'data': {
                'dggphis232340-2846342': []
            }
        }),
    }));
    const wrapper = mount(JobDetailsLayout);
    const vm = wrapper.vm;
    
    it('renders when initScrollerItem correctly', async () => {
        vm.initScrollerItem();
        await flushPromises();
        expect(wrapper.text()).toContain('JobInfos EntrypointRuntimeEnvSubmissionIDStatusStartTime' +
            'EndTimeMessageErrorTypeDriverNodeIDDriverNodeIPDriverPID');
    });
})
