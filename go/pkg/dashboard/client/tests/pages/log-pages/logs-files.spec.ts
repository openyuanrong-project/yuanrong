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
import {describe, it, expect, vi } from 'vitest';
import CommonCard from '@/components/common-card.vue';
import LogsFiles from '@/pages/log-pages/logs-files.vue';

describe('LogsFiles', () => {
    const wrapper = mount(LogsFiles, {
        global: {
            stubs: {
                CommonCard: true,
                TinyLink: true,
                TinySearch: true,
                BreadcrumbComponent: true,
            },
        },
    });

    it('renders the main structure correctly', () => {
        expect(wrapper.findComponent(CommonCard).exists()).toBe(true);
    });
})

describe('LogsFiles.initScrollerItem', () => {
    vi.mock('@/api/api', () => {
        return {
            ListLogsAPI: vi.fn().mockResolvedValue({
                'message': '',
                'data': {
                    'dggphis232339-189755': [
                        'runtime-10050000-0000-4000-b00f-8374b8dd2508-00000000009d.out',
                        'runtime-10050000-0000-4000-b00f-8374b8dd2508-00000000009d.err',
                    ]}}),
        }
    })

    const wrapper = mount(LogsFiles);
    const vm = wrapper.vm;
    it('renders when initScrollerItem correctly', async () => {
        vm.initScrollerItem();
        await flushPromises();
        expect(wrapper.text()).toContain('Log Files');
    });
})
