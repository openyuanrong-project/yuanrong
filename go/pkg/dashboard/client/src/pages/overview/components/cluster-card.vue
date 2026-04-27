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
      Cluster Status
    </template>
    <template v-slot:card-content>
      <div class="content-box">
        <div>Total: {{ proxyNum }} node</div>
        <div>Alive: {{ proxyNum }} node</div>
        <div class="space"></div>
        <tiny-link class="font-size24" href="#/cluster">View All Cluster Status ></tiny-link>
      </div>
    </template>
  </common-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { TinyLink } from '@opentiny/vue';
import { GetRsrcSummAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';
import { SWR } from '@/utils/swr';

const proxyNum = ref(0);

onMounted(() => {
  SWR(initStatus);
})

function initStatus() {
  GetRsrcSummAPI().then((res: GetRsrcSummAPIRes) => {
    proxyNum.value = res?.data?.proxy_num ? res.data.proxy_num : 0;
  }).catch((err: any)=>{
    WarningNotify('GetRsrcSummAPI', err);
  });
}
</script>

<style>
.content-box {
  margin-top: 60px;
}

.content-box>div {
  line-height: 28px;
}
</style>

<style scoped>
.space {
  height: 120px;
}
</style>