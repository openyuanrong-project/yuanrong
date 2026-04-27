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
      Logical Resources
    </template>
    <template v-slot:card-content>
      <div style="height: 270px">
        <div class="gauge-legend">Logical CPU</div>
        <div class="gauge-legend gauge-legend2">Logical Memory(GB)</div>
        <tiny-chart-gauge class="gauge-item" :options="cpuOptions"></tiny-chart-gauge>
        <tiny-chart-gauge class="gauge-item" :options="memOptions"></tiny-chart-gauge>
      </div>
    </template>
  </common-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { TinyHuichartsGauge as TinyChartGauge } from '@opentiny/vue-huicharts';
import { GetRsrcSummAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';
import { CPUConvert, Float2, MBToGB } from '@/utils/handleNum';
import { SWR } from '@/utils/swr';

const cpuPercent = ref(0);
const memPercent = ref(0);
const resData = ref({
  cap_cpu: 0,
  cap_mem: 0,
  used_cpu: 0,
  used_mem: 0,
});

onMounted(() => {
    SWR(initRsrc);
})

function initRsrc() {
  GetRsrcSummAPI().then((res: GetRsrcSummAPIRes) => {
    if (!res?.data) {
      throw '';
    }
    const data = res.data;
    resData.value.cap_cpu = CPUConvert(data.cap_cpu);
    resData.value.cap_mem = MBToGB(data.cap_mem);
    resData.value.used_cpu = CPUConvert(data.cap_cpu-data.alloc_cpu);
    resData.value.used_mem = MBToGB(data.cap_mem-data.alloc_mem);
    cpuPercent.value = Float2(resData.value.used_cpu / resData.value.cap_cpu * 100);
    memPercent.value = Float2(resData.value.used_mem / resData.value.cap_mem * 100);
  }).catch((err: any)=>{
    WarningNotify('GetRsrcSummAPI', err);
  });
}

const commonStyle = {
  position: {
    center: ['60%', '32%'],
    radius: '56%',
  },
  splitLine: {
    show: false,
  },
  tooltip: {
    show: false,
  },
};

const formatterStyle = {
  formatterStyle: {
    num: {
      fontSize: 30,
      color: '#000',
    },
    value: {
      fontSize: 26,
      color: '#000',
      padding: [30, 0, 0, 0],
    },
  },
};

const cpuOptions = ref({
  ...commonStyle,
  data: [
    {
      value: cpuPercent,
    },
  ],
  text: {
    formatter(value: string) {
      return '{num|' + resData.value.used_cpu + ' / ' + resData.value.cap_cpu + '}\n{value|' + value + '%}';
    },
    ...formatterStyle,
  },
});

const memOptions = ref({
  ...commonStyle,
  position: {
    center: ['40%', '32%'],
    radius: '56%',
  },
  data: [
    {
      value: memPercent,
    },
  ],
  text: {
    formatter(value: string) {
      return '{num|' + resData.value.used_mem + ' / ' + resData.value.cap_mem + '}\n{value|' + value + '%}';
    },
    ...formatterStyle,
  },
});
</script>

<style scoped>
.gauge-legend {
  display: inline-block;
  width: 21%;
  margin-left: 16%;
}

.gauge-item {
  display: inline-block;
  width: 50%;
}
</style>
