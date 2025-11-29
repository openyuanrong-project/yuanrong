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
import { describe, it, expect } from 'vitest';
import { TinyNavMenu } from '@opentiny/vue';
import Layout from '@/pages/layout.vue';

describe('Layout', () => {
    const wrapper = mount(Layout, {
        global: {
            stubs: {
                TinyNavMenu: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(TinyNavMenu).exists()).toBe(true);
    });
})

describe('Layout.text', () => {
    const wrapper = mount(Layout);

    it('renders text correctly', () => {
        expect(wrapper.text()).toContain('OverviewClusterInstancesLogs');
    });
})