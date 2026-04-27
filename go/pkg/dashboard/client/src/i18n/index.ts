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

import { createI18n } from 'vue-i18n';
import locale from '@opentiny/vue-locale';

const initI18n = (i18n) =>
    locale.initI18n({
        i18n,
        createI18n,
        messages: {
            zhCN: {
                test: '中文',
            },
            enUS: {
                test: 'English',
            },
        },
    });

export const i18n = initI18n({ locale: 'enUS' });