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
import { TinyLink } from '@opentiny/vue';
import { GetInstSummAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import InstancesCard from '@/pages/overview/components/instances-card.vue';

describe('InstancesCard', () => {
    const wrapper = mount(InstancesCard, {
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

describe('InstancesCard.TinyLink', async () => {
    const wrapper = mount(TinyLink, {
        props: { href: '#/instances' },
        slots: {
            default: 'View All Instances >',
        },
    });
    await wrapper.trigger('click');
    it('renders TinyLink correctly',  () => {
        expect(wrapper.emitted('click')).toHaveLength(1);
        expect(wrapper.text()).toBe('View All Instances >');
    });
})

describe('InstancesCard.initStatus', () => {
    vi.mock('@/api/api', () => ({
        GetInstSummAPI: vi.fn().mockResolvedValue({
            data: {
                total: 5,
                running: 2,
                exited: 1,
                fatal: 0,
            }}),
    }));

    const wrapper = mount(InstancesCard);
    const vm = wrapper.vm;
    it('renders when initStatus correctly', async () => {
        vm.initStatus();
        await flushPromises();
        expect(wrapper.text()).toBe('Instances Total: 5Running: 2Exited: 1Fatal: 0Others: 2View All Instances >');
    });
})
