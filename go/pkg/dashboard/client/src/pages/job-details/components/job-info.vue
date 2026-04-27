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
      JobInfos
    </template>
    <template v-slot:card-content>
      <div class="margin-top10">
        <tiny-layout>
          <tiny-row>
            <tiny-col :span="12">
              <div class="font-size18">Entrypoint</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="12">
              <div class="font-size14 auto-tip" v-auto-tip>{{ jobInfo.entrypoint }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="12">
              <div class="font-size18">RuntimeEnv</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="12">
              <div class="font-size14 auto-tip" v-auto-tip>{{ jobInfo.runtime_env }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">SubmissionID</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">Status</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">StartTime</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ jobInfo.submission_id }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ jobInfo.status }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ jobInfo.start_time }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">EndTime</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">Message</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">ErrorType</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ jobInfo.end_time }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14 auto-tip" v-auto-tip>{{ jobInfo.message }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14 auto-tip" v-auto-tip>{{ jobInfo.error_type }}</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size18">DriverNodeID</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">DriverNodeIP</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size18">DriverPID</div>
            </tiny-col>
          </tiny-row>
          <tiny-row>
            <tiny-col :span="4">
              <div class="font-size14">{{ jobInfo.driver_node_id }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ jobInfo.driver_node_ip }}</div>
            </tiny-col>
            <tiny-col :span="4">
              <div class="font-size14">{{ jobInfo.driver_node_pid }}</div>
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
import { GetJobInfoAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';
import { DayFormat } from '@/utils/dayFormat';
import { SWR } from '@/utils/swr';

const VAutoTip = AutoTip;
const jobInfo = ref({});

onMounted(() => {
  SWR(initJobInfo);
})

function initJobInfo() {
  const submissionID = window.location.hash.split('/')[2];
  GetJobInfoAPI(submissionID).then((res: JobInfo) => {
    if (!res?.submission_id) {
      throw '';
    }
    jobInfo.value = res;
    jobInfo.value.start_time = DayFormat(res.start_time);
    jobInfo.value.end_time = DayFormat(res.end_time);
    jobInfo.value.driver_node_ip = res.driver_info.node_ip_address;
    jobInfo.value.driver_node_pid = res.driver_info.pid;
  }).catch((err: any)=>{
    WarningNotify('GetJobInfoAPI', err?.response?.data);
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
