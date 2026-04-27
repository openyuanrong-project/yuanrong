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
          <tiny-grid-column field="submission_id" title="SubmissionID" width="22%" align="center"
                            :filter={} :renderer="{component: TinyLink, attrs: toJobDetails}">
            <template #filter="data">
              <input v-model="allFilters.submission_id" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="entrypoint" title="Entrypoint" align="center" show-overflow="tooltip">
          </tiny-grid-column>
          <tiny-grid-column field="status" title="Status" align="center" :filter="statusFilter"></tiny-grid-column>
          <tiny-grid-column field="message" title="Message" align="center"></tiny-grid-column>
          <tiny-grid-column field="start_time" title="StartTime" align="center" sortable></tiny-grid-column>
          <tiny-grid-column field="end_time" title="EndTime" align="center" sortable></tiny-grid-column>
          <tiny-grid-column field="driver_info.pid" title="DriverPID" align="center"></tiny-grid-column>
          <tiny-grid-column field="log" title="Log" align="center"
                            :renderer="{component: TinyLink, attrs: toJobDetails}"></tiny-grid-column>
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
import { ListJobsAPI } from '@/api/api';
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
  submission_id: '',
  selected_status: [],
});

onMounted(() => {
  SWR(initChart);
})

function initChart() {
  ListJobsAPI().then((res: GetListJobsAPIRes) => {
    if (!res || !Array.isArray(res)) {
      throw '';
    }
    tableData.value = res.filter((job: any)=>{
      if (!job?.submission_id?.includes(allFilters.value.submission_id)) {
        return false;
      }
      if (allFilters.value.selected_status.length === 0) {
        return true;
      }
      for (let status of allFilters.value.selected_status) {
        if (job?.status?.includes(status)) {
          return true;
        }
      }
      return false;
    });
    for (let job of tableData.value) {
      job.start_time = DayFormat(job.start_time);
      job.end_time = DayFormat(job.end_time);
      job.log = 'log';
    }
    ChartSort(tableData.value, 'start_time', 'id');
    gridRef.value.handleFetch();
  }).catch((err: any)=>{
    WarningNotify('ListJobsAPI', err);
  });
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

function toJobDetails(data: any) {
  return {href: '#/jobs/' + data.row.submission_id};
}
</script>

<style scoped>
</style>