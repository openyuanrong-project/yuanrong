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
import EmptyLogCard from '@/pages/instance-details/components/empty-log-card.vue';
import InstanceInfo from '@/pages/instance-details/components/instance-info.vue';
import InstanceDetailsLayout from '@/pages/instance-details/instance-details-layout.vue';

describe('InstanceDetailsLayout', () => {
    const wrapper = mount(InstanceDetailsLayout, {
        global: {
            stubs: {
                BreadcrumbComponent: true,
                LogContentTemplate: true,
                EmptyLogCard: true,
                InstanceInfo: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(BreadcrumbComponent).exists()).toBe(true);
        expect(wrapper.findComponent(EmptyLogCard).exists()).toBe(true);
        expect(wrapper.findComponent(InstanceInfo).exists()).toBe(true);
    });
})

describe('InstanceDetailsLayout.ListLogsAPI', () => {
    vi.mock('@/api/api', () => ({
        ListLogsAPI: vi.fn().mockResolvedValue({
            'message': '',
            'data': {
                'dggphis232340-2846342': []
            }
        }),
    }));

    const wrapper = mount(InstanceDetailsLayout);
    it('renders when initScrollerItem error or empty', async () => {
        wrapper.vm.initScrollerItem();
        await flushPromises();
        expect(wrapper.text()).toContain('InstanceInfos IDStatusJobIDPIDIPNodeIDParentID' +
            'CreateTimeRequiredCPURequiredMemory(MB)RequiredGPURequiredNPURestartedExitDetailLog');
        expect(wrapper.text()).toContain('Driver logs are only available when submitting jobs ' +
            'via the Job Submission API, SDK.');
    });
})

