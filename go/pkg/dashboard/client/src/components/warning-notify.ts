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

import { TinyNotify } from '@opentiny/vue';

export const WarningNotify = (tip: string, err: any) => {
    let message = '';
    if (typeof err === 'string' && err != '') {
        message = err;
    }else {
        err = 'backend error';
        message = 'Unable to access backend service.';
    }
    console.error(tip, 'error:', err);
    TinyNotify({
        type: 'warning',
        message: message,
        position: 'top-right',
        duration: 5000,
    });
};