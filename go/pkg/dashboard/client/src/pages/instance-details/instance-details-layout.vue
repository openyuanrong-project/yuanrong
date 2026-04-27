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

<template>
  <breadcrumb-component />
  <instance-info />
  <div v-for="filename in filenames" :key="filename">
    <log-content-template :filename="filename"/>
  </div>
  <div v-if="filenames.length == 0">
    <empty-log-card />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ListLogsAPI } from '@/api/api';
import BreadcrumbComponent from '@/components/breadcrumb-component.vue';
import LogContentTemplate from '@/components/log-content-template.vue';
import { WarningNotify } from '@/components/warning-notify';
import EmptyLogCard from './components/empty-log-card.vue';
import InstanceInfo from './components/instance-info.vue';

const filenames = ref([]);

onMounted(() => {
  initScrollerItem();
})

function initScrollerItem() {
  const id = window.location.hash.split('/')[2];
  ListLogsAPI(id).then((res: ListLogsRes) => {
    if (!res?.data) {
      throw '';
    }
    const logs = []
    for (let [_, value] of Object.entries(res.data)) {
      logs.push(...value);
    }
    logs.sort().reverse();
    filenames.value = logs;
  }).catch((err: any)=>{
    WarningNotify('ListLogsAPI', err);
  });
}
</script>

<style scoped>
</style>
