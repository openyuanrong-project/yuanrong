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

import { createApp } from 'vue';
import { createRouter, createWebHashHistory } from 'vue-router';
import VueVirtualScroller from 'vue-virtual-scroller';
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css';
import '@opentiny/vue-theme/index.css';
import { i18n } from '@/i18n';
import '@/index.css';
import App from '@/pages/layout.vue';

const routes = [
    {
        path:'/overview',
        component: () => import(/* webpackChunkName: "overview" */ '@/pages/overview/overview-layout.vue'),
    },
    {
        path:'/cluster',
        component : () => import(/* webpackChunkName: "cluster" */ '@/pages/cluster/cluster-layout.vue'),
    },
    {
        path:'/jobs',
        component : () => import(/* webpackChunkName: "instances" */ '@/pages/jobs/jobs-chart.vue'),
    },
    {
        path:'/jobs/:jobID',
        component : () =>
            import(/* webpackChunkName: "instanceDetails" */ '@/pages/job-details/job-details-layout.vue'),
    },
    {
        path:'/instances',
        component : () => import(/* webpackChunkName: "instances" */ '@/pages/instances/instances-chart.vue'),
    },
    {
        path:'/instances/:instanceID',
        component : () =>
            import(/* webpackChunkName: "instanceDetails" */ '@/pages/instance-details/instance-details-layout.vue'),
    },
    {
        path:'/logs',
        component : () =>
            import(/* webpackChunkName: "logs" */ '@/pages/log-pages/logs-nodes.vue'),
    },
    {
        path:'/logs/:nodeID',
        component : () =>
            import(/* webpackChunkName: "logsNodeID" */ '@/pages/log-pages/logs-files.vue'),
    },
    {
        path:'/logs/:nodeID/:filename',
        component : () =>
            import(/* webpackChunkName: "logsNodeIDFilename" */ '@/pages/log-pages/logs-content.vue'),
    },
];

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {path: '/', redirect: '/overview'},
        ...routes,
    ],
});

const app = createApp(App);
app.use(i18n);
app.use(router);
app.use(VueVirtualScroller);

app.mount('#app');