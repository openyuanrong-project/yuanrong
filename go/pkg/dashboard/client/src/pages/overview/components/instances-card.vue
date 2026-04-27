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
      Instances
    </template>
    <template v-slot:card-content>
      <div class="content-box">
        <div>Total: {{ totalNum }}</div>
        <div>Running: {{ runningNum }}</div>
        <div>Exited: {{ exitedNum }}</div>
        <div>Fatal: {{ fatalNum }}</div>
        <div>Others: {{ othersNum }}</div>
        <div class="space"></div>
        <tiny-link class="font-size24" href="#/instances">View All Instances ></tiny-link>
      </div>
    </template>
  </common-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { TinyLink } from '@opentiny/vue';
import { GetInstSummAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';
import { SWR } from '@/utils/swr';

const totalNum = ref(0);
const runningNum = ref(0);
const exitedNum = ref(0);
const fatalNum = ref(0);
const othersNum = ref(0);

onMounted(() => {
  SWR(initStatus);
})

function initStatus() {
  GetInstSummAPI().then((res: GetInstSummAPIRes) => {
    if (!res?.data) {
      throw '';
    }
    totalNum.value = res.data.total;
    runningNum.value = res.data.running;
    exitedNum.value = res.data.exited;
    fatalNum.value = res.data.fatal;
    othersNum.value = totalNum.value - runningNum.value - exitedNum.value - fatalNum.value;
  }).catch((err: any)=>{
    WarningNotify('GetInstSummAPI', err);
  });
}
</script>

<style scoped>
.space {
  height: 36px;
}
</style>
