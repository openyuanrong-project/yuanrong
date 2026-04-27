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
  <common-card>
    <template v-slot:card-title>
      Logs
    </template>
    <template v-slot:card-content>
      <div class="margin-top10 font-size20">Select a node to view logs:</div>
      <div class="margin-left20" v-for="node in nodes" :key="node">
        <tiny-link :href="'#/logs/' + node">
          {{ node + '(IP:' +  (nodesObj[node] ? nodesObj[node] : '') + ')'}}
        </tiny-link>
      </div>
    </template>
  </common-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { TinyLink } from '@opentiny/vue';
import { GetCompAPI, ListLogsAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';

const nodes = ref([]);
const nodesObj = ref({});

onMounted(async () => {
  await initNodesObj();
  await initScrollerItem();
})

async function initNodesObj() {
  await GetCompAPI().then((res: GetCompAPIRes) => {
    if (!res?.data?.nodes) {
      throw '';
    }
    const nodesData = res.data.nodes;
    for (let node of nodesData) {
      nodesObj.value[node['hostname']] = node['address'];
    }
  }).catch((err: any)=>{
    WarningNotify('Log GetCompAPI', err);
  });
}

async function initScrollerItem() {
  await ListLogsAPI("").then((res: ListLogsRes) => {
    if (!res?.data) {
      throw '';
    }
    for (let [key, _] of Object.entries(res.data)) {
      nodes.value.push(key);
    }
    nodes.value.sort();
  }).catch((err: any)=>{
    WarningNotify('ListLogsAPI', err);
  });
}
</script>

<style scoped>
</style>