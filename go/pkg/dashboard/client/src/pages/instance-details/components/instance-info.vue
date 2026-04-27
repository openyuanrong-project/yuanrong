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
      InstanceInfos
    </template>
    <template v-slot:card-content>
      <div class="margin-top10">
        <tiny-layout>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">ID</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">Status</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">JobID</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.id }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.status }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.job_id }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">PID</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">IP</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">NodeID</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.pid }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.ip }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.node_id }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">ParentID</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">CreateTime</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">RequiredCPU</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.parent_id }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.create_time }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.required_cpu }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">RequiredMemory(MB)</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">RequiredGPU</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">RequiredNPU</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.required_mem }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.required_gpu }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.required_npu }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">Restarted</div>
            </tiny-col>
            <tiny-col :span="8">
              <div class="font-size18">ExitDetail</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ instanceInfo.restarted }}</div>
            </tiny-col>
            <tiny-col :span="8">
              <div class="font-size14 auto-tip" v-auto-tip>{{ instanceInfo.exit_detail }}</div>
            </tiny-col>
          </tiny-row>
        </tiny-layout>
      </div>
    </template>
  </common-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { TinyCol, TinyLayout, TinyRow } from '@opentiny/vue';
import { AutoTip } from '@opentiny/vue-directive';
import { GetInstInstIDAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';
import { DayFormat } from '@/utils/dayFormat';
import { CPUConvert } from '@/utils/handleNum';
import { SWR } from '@/utils/swr';

const VAutoTip = AutoTip;
const instanceInfo = ref({});

onMounted(() => {
  SWR(initInstanceInfo);
})

function initInstanceInfo() {
  const instanceID = window.location.hash.split('/')[2];
  GetInstInstIDAPI(instanceID).then((res: InstInfo) => {
    if (!res?.id) {
      throw 'Instance not found, failed to get instance info, instance ID: ' + instanceID;
    }
    instanceInfo.value = res;
    instanceInfo.value.create_time = DayFormat(res.create_time);
    instanceInfo.value.required_cpu = CPUConvert(res.required_cpu);
    if (res.restarted === -1) {
      instanceInfo.value.restarted = 0;
    }
  }).catch((err: any)=>{
    WarningNotify('GetInstInstIDAPI', err);
  });
}
</script>

<style scoped>
.tiny-layout > :nth-child(even) {
  margin-bottom: 12px;
}

.tiny-row {
  margin-bottom: 4px;
}

.auto-tip {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
