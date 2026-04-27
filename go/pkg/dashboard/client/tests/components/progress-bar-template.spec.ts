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

import { h } from 'vue';
import { describe, it, expect } from 'vitest';
import { ProgressBar, SimpleProgressBar, MultiProgressBar } from '@/components/progress-bar-template';

describe('ProgressBarTemplate', () => {
    it('renders ProgressBar correctly', () => {
        const progressBar = ProgressBar(h, 2.93, 38);
        expect(JSON.stringify(progressBar)).toContain('2.93/38(7.71%)');
    });
    it('renders SimpleProgressBar correctly', () => {
        const simpleProgressBar  = SimpleProgressBar(h, 10.68);
        expect(JSON.stringify(simpleProgressBar)).toContain('10.68%');
    });
    it('renders MultiProgressBar correctly', () => {
        const multiProgressBar = MultiProgressBar(h, {1: {used : 11.8, capacity: 100}});
        expect(JSON.stringify(multiProgressBar)).toContain('11.8%');
    });
})