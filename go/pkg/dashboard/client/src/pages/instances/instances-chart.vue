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
    <template v-slot:card-content>
      <div class="margin-top16">
        <tiny-grid :fetch-data="fetchData" @filter-change="filterChangeEvent" remote-filter
                   ref="gridRef" :pager="pagerConfig">
          <tiny-grid-column field="id" title="ID" width="19%" align="center"
                            :filter={} :renderer="{component: TinyLink, attrs: toInstanceDetails}">
            <template #filter="data">
              <input v-model="allFilters.id" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="status" title="Status" align="center" :filter="statusFilter" width="6%">
          </tiny-grid-column>
          <tiny-grid-column field="job_id" title="JobID" align="center" show-overflow="tooltip" width="7%"
                            :filter={}>
            <template #filter="data">
              <input v-model="allFilters.job_id" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="pid" title="PID" align="center" width="6%" :filter={}>
            <template #filter="data">
              <input v-model="allFilters.pid" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="ip" title="IP" align="center" width="10%" :filter={}>
            <template #filter="data">
              <input v-model="allFilters.ip" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="node_id" title="NodeID" align="center" show-overflow="tooltip" width="7%"
                            :filter={}>
            <template #filter="data">
              <input v-model="allFilters.node_id" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="parent_id" title="ParentID" align="center" show-overflow="tooltip" width="7%"
                            :filter={}>
            <template #filter="data">
              <input v-model="allFilters.parent_id" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="create_time" title="CreateTime" align="center" sortable width="10%">
          </tiny-grid-column>
          <tiny-grid-column field="required_cpu" title="Required CPU" align="center" width="6%"></tiny-grid-column>
          <tiny-grid-column field="required_mem" title="Required Memory(MB)" align="center" width="7%">
          </tiny-grid-column>
          <tiny-grid-column field="required_gpu" title="Required GPU" align="center" width="6%"></tiny-grid-column>
          <tiny-grid-column field="required_npu" title="Required NPU" align="center" width="6%"></tiny-grid-column>
          <tiny-grid-column field="restarted" title="Restarted" align="center" width="6%"></tiny-grid-column>
          <tiny-grid-column field="exit_detail" title="ExitDetail" align="center" show-overflow="tooltip" width="7%">
          </tiny-grid-column>
          <tiny-grid-column field="log" title="Log" align="center" width="4%"
                            :renderer="{component: TinyLink, attrs: toInstanceDetails}"></tiny-grid-column>
          <template #empty>
            <EmptyChart />
          </template>
        </tiny-grid>
      </div>
    </template>
  </common-card>
</template>

<script setup lang="tsx">
import { ref, onMounted } from 'vue';
import { TinyGrid, TinyGridColumn, TinyLink } from '@opentiny/vue';
import { GetInstAPI, GetInstParentIDAPI } from '@/api/api';
import { pagerConfig, statusFilter } from '@/components/chart-config';
import CommonCard from '@/components/common-card.vue';
import EmptyChart from '@/components/empty-chart.vue';
import { WarningNotify } from '@/components/warning-notify';
import { DayFormat } from '@/utils/dayFormat';
import { ChartSort } from '@/utils/sort';
import { SWR } from '@/utils/swr';

const fetchData = ref({ api: getData });
const tableData = ref([]);
const gridRef = ref();
const allFilters = ref({
  id: '',
  selected_status: [],
  job_id: '',
  pid: '',
  ip: '',
  node_id: '',
  parent_id: '',
});

onMounted(() => {
  SWR(initChart);
})

function initChart() {
  const pathArr = window.location.hash.split('/');
  if (pathArr.length > 2 && pathArr[1] == 'jobs') {
    GetInstParentIDAPI(pathArr[2]).then((res: InstInfo[]) => {
      if (!res || !Array.isArray(res)) {
        throw '';
      }
      tableData.value = res;
      handleChartData();
    }).catch((err: any)=>{
      WarningNotify('GetInstParentIDAPI', err);
    });
  }else {
    GetInstAPI().then((res: GetInstAPIRes) => {
      if (!res?.data) {
        throw '';
      }
      tableData.value = res.data;
      handleChartData();
    }).catch((err: any)=>{
      WarningNotify('GetInstAPI', err);
    });
  }
}

function handleChartData() {
  tableData.value = tableData.value.filter((instance: any)=>{
    let isSelectedStatus = false;
    if (allFilters.value.selected_status.length === 0) {
      isSelectedStatus = true;
    }
    for (let status of allFilters.value.selected_status) {
      if (instance?.status?.includes(status)) {
        isSelectedStatus = true;
        break;
      }
    }
    return isSelectedStatus && instance?.id?.includes(allFilters.value.id)
        && instance?.job_id?.includes(allFilters.value.job_id) && instance?.pid?.includes(allFilters.value.pid)
        && instance?.ip?.includes(allFilters.value.ip) && instance?.node_id?.includes(allFilters.value.node_id)
        && instance?.parent_id?.includes(allFilters.value.parent_id);
  })

  for (let instance of tableData.value) {
    instance.create_time = DayFormat(instance.create_time);
    instance.required_cpu /= 1000;
    instance.log = 'log';
    if (instance.restarted == -1) {
      instance.restarted = '';
    }
  }
  ChartSort(tableData.value, 'create_time', 'id');
  gridRef.value.handleFetch();
}

function getData({ page }) {
  const { currentPage, pageSize } = page;
  const offset = (currentPage - 1) * pageSize;
  return Promise.resolve({
    result: tableData.value.slice(offset, offset + pageSize),
    page: { total: tableData.value.length },
  });
}

function filterChangeEvent({ filters }) {
  allFilters.value.selected_status = filters?.status?.value ? filters.status.value : [];
  initChart();
}

function toInstanceDetails(data: any) {
  return {href: '#/instances/' + data.row.id};
}
</script>

<style scoped>
</style>