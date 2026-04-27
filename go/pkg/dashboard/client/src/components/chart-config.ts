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

import { ref } from 'vue';

export const pagerConfig = ref({
    attrs: {
        currentPage: 1,
        pageSize: 10,
        pageSizes: [10, 20, 50],
        total: 0,
        align: 'right',
        layout: 'total, sizes, prev, pager, next, jumper',
    }
});

export const statusFilter = ref({
    layout: 'simple',
    multi: true,
    simpleFilter: {
        selectAll: true
    },
});