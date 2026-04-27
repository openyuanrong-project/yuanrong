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
      Components
    </template>
    <template v-slot:card-content>
      <div class="margin-top16">
        <tiny-grid :fetch-data="fetchData" ref="treeRef" @filter-change="initChartWithProm" remote-filter
                   :tree-config="{ children: 'children' }" :pager="pagerConfig" height="750">
          <tiny-grid-column field="hostname" title="Hostname" tree-node width="27%" align="center"
                            :filter={} sortable>
            <template #filter="data">
              <input v-model="filters.hostname" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="status" title="Status" align="center" width="7%"></tiny-grid-column>
          <tiny-grid-column field="address" title="IP/Address" align="center" :filter={} width="10%">
            <template #filter="data">
              <input v-model="filters.address" @keyup.enter="data.context.confirmFilter" />
              <button @click="data.context.confirmFilter">Confirm</button>
            </template>
          </tiny-grid-column>
          <tiny-grid-column field="cpu" title="CPU" :renderer="NodeCPURender" align="center" width="9%">
          </tiny-grid-column>
          <tiny-grid-column field="memory" title="Memory(GB)" :renderer="NodeMemRender" align="center" width="12%">
          </tiny-grid-column>
          <tiny-grid-column field="npu" title="NPU" :renderer="NodeNPURender" align="center" width="12%">
          </tiny-grid-column>
          <tiny-grid-column field="disk" title="Disk(GB)" :renderer="NodeDiskRender" align="center" width="12%">
          </tiny-grid-column>
          <tiny-grid-column field="logicalCPU" title="Logical Resources" :renderer="logicalResourcesRender"
                            width="11%"></tiny-grid-column>
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
import { TinyGrid, TinyGridColumn } from '@opentiny/vue';
import { GetCompAPI, GetInstAPI, GetPromQueryAPI } from '@/api/api';
import { pagerConfig } from '@/components/chart-config';
import CommonCard from '@/components/common-card.vue';
import EmptyChart from '@/components/empty-chart.vue';
import { MultiProgressBar, ProgressBar, SimpleProgressBar } from '@/components/progress-bar-template';
import { WarningNotify } from '@/components/warning-notify';
import { CPUConvert, MBToGB } from '@/utils/handleNum';
import { ChartSort } from '@/utils/sort';
import { SWR} from '@/utils/swr';

const NODE_ID_KEY = 'node_id';
const AGENT_ID_KEY = 'agent_id';
const INSTANCE_ID_KEY = 'instance_id';
const NULL_DATA = 'N/A';

const fetchData = ref({ api: getData });
const tableData = ref([]);
const treeRef = ref();
const filters = ref({
  hostname: '',
  address: '',
});

let instances = [];
let instanceCpuUsages = {};
let instanceMemUsages = {};
let instanceNpuData = {};
let nodeCpuUsages = {};
let nodeMemCaps = {};
let nodeMemUsages = {};
let nodeDiskCaps = {};
let nodeDiskUsages = {};
let nodeNpuData = {};
let expandRowMap = new Set();
let expandRows = [];
let originExpandRows = [];

onMounted(() => {
  GetPromQueryAPI('up').catch((err: any)=>{
    WarningNotify('GetPromQueryAPI', err.response?.data);
  })
  SWR(initChartWithProm);
})

async function initChartWithProm() {
  await getInstancePromData('yr_instance_cpu_usage', instanceCpuUsages);
  await getInstancePromData('yr_instance_memory_usage', instanceMemUsages);
  await getInstancePromData('yr_instance_npu', instanceNpuData, true);
  await getInstancesInfo();

  await getNodePromData('yr_node_cpu_usage', nodeCpuUsages);
  await getNodePromData('yr_node_memory_capacity', nodeMemCaps);
  await getNodePromData('yr_node_memory_usage', nodeMemUsages);
  await getNodePromData('yr_node_disk_capacity', nodeDiskCaps);
  await getNodePromData('yr_node_disk_usage', nodeDiskUsages);
  await getNodePromData('yr_node_npu', nodeNpuData, true);
  await initChart();
}

async function getInstancePromData(metricName: string, values: any, isNPU: boolean = false) {
  await GetPromQueryAPI(metricName).then((res: PromData) => {
    for (let piece of res) {
      values[piece.metric[INSTANCE_ID_KEY]] = isNPU ? piece.metric['details'] : piece.value[1];
    }
  }).catch((err: any)=>{
    console.error('GetPromQueryAPI', 'error:', err.response?.data);
  });
}

async function getNodePromData(metricName: string, values: any, isNPU: boolean = false) {
  await GetPromQueryAPI(metricName).then((res: PromData) => {
    for (let piece of res) {
      if (!values[piece.metric[NODE_ID_KEY]]) {
        values[piece.metric[NODE_ID_KEY]] = {};
      }
      values[piece.metric[NODE_ID_KEY]][piece.metric[AGENT_ID_KEY]] = isNPU ? piece.metric['details'] : piece.value[1];
    }
  }).catch((err: any)=>{
    console.error('GetPromQueryAPI', 'error:', err.response?.data);
  });
}

function getNodePromValue(node: any, values: any) {
  for (let agent of node.children) {
    if (values[node.hostname] != undefined && values[node.hostname][agent.hostname] != undefined) {
      return  values[node.hostname][agent.hostname]
    }
  }
}

function initAgentIDInstancesMap() {
  let agentIDInstancesMap = {};
  for (let instance of instances) {
    if (!agentIDInstancesMap[instance.agent_id]) {
      agentIDInstancesMap[instance.agent_id] = [];
    }
    const instanceInfo = {
      hostname: instance.id,
      status: instance.status,
      address: instance.ip,
      required_cpu: instance.required_cpu,
      required_mem: instance.required_mem,
      required_npu: instance.required_npu,
      cpu_usage: instanceCpuUsages[instance.id],
      mem_cap: nodeMemCaps[instance.node_id] ? MBToGB(nodeMemCaps[instance.node_id][instance.agent_id]) : undefined,
      mem_usage: MBToGB(instanceMemUsages[instance.id]),
      npu_data: instanceNpuData[instance.id],
      create_time: instance.create_time,
    };
    agentIDInstancesMap[instance.agent_id].push(instanceInfo);
  }
  return agentIDInstancesMap;
}

function saveRowExpansion() {
  originExpandRows = treeRef.value.getTreeExpandeds();
  expandRowMap = new Set();
  for (let row of originExpandRows) {
    expandRowMap.add(row.hostname);
  }
}

function loadRowExpansion() {
  expandRows = [];
  for (let nodeRow of tableData.value) {
    if (!expandRowMap.has(nodeRow.hostname)) {
      continue
    }
    expandRows.push(nodeRow)
    for (let agentRow of nodeRow.children) {
      if (expandRowMap.has(agentRow.hostname)) {
        expandRows.push(agentRow)
      }
    }
  }
  treeRef.value.setTreeExpansion(originExpandRows, false);
  treeRef.value.setTreeExpansion(expandRows, true);
}

async function initChart() {
  // 保存折叠状态
  saveRowExpansion();
  await GetCompAPI().then((res: GetCompAPIRes) => {
    if (!res?.data?.nodes) {
      throw '';
    }
    const agentIDInstancesMap = initAgentIDInstancesMap();

    tableData.value = res.data.nodes.filter((node: any) =>{
      return node?.hostname?.includes(filters.value.hostname) && node?.address?.includes(filters.value.address);
    });
    tableData.value.sort((a, b) => a.hostname.localeCompare(b.hostname));
    for (let node of tableData.value) {
      node.children = res.data.components[node.hostname];
      node.children.sort((a, b) => a.hostname.localeCompare(b.hostname));
      for (let agent of node.children) {
        agent.children = agentIDInstancesMap[agent.hostname];
        ChartSort(agent.children, 'create_time', 'hostname');
      }
      node.cpu_usage = getNodePromValue(node, nodeCpuUsages);
      node.mem_cap = MBToGB(getNodePromValue(node, nodeMemCaps));
      node.mem_usage = MBToGB(getNodePromValue(node, nodeMemUsages));
      node.npu_data = getNodePromValue(node, nodeNpuData);
      node.disk_cap = getNodePromValue(node, nodeDiskCaps);
      node.disk_usage = getNodePromValue(node, nodeDiskUsages);
    }
    // 恢复折叠状态
    loadRowExpansion();
    treeRef.value.handleFetch();
  }).catch((err: any)=>{
    WarningNotify('GetCompAPI', err);
  });
}

async function getInstancesInfo() {
  await GetInstAPI().then((res: GetInstAPIRes) => {
    if (!res?.data) {
      throw '';
    }
    instances = res.data;
  }).catch((err: any)=>{
    WarningNotify('GetInstAPI', err);
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

function NodeCPURender(h: any, { row }) {
  return row.cpu_usage != undefined ? SimpleProgressBar(h, row.cpu_usage) : NULL_DATA
}

function NodeMemRender(h: any, { row }) {
  return !Number.isNaN(row.mem_cap) && row.mem_cap != undefined ?
      ProgressBar(h, row.mem_usage, row.mem_cap) : NULL_DATA
}

function NodeNPURender(h: any, { row }) {
  if (row.npu_data === undefined) {
    return NULL_DATA
  }
  return MultiProgressBar(h, JSON.parse(row.npu_data))
}

function NodeDiskRender(h: any, { row }) {
  return row.disk_cap != undefined ? ProgressBar(h, row.disk_usage, row.disk_cap) : NULL_DATA
}

function logicalResourcesRender(h: any, { row }) {
  if (!row.cap_cpu) {
    return h('div', [
      h('div', [CPUConvert(row.required_cpu), ' CPU']),
      h('div', [MBToGB(row.required_mem), 'GB Memory']),
      h('div', [row.required_npu, ' NPU']),
    ]);
  }

  const usedCPU = CPUConvert(row.cap_cpu - row.alloc_cpu);
  const totalCPU = CPUConvert(row.cap_cpu);
  const usedMem = MBToGB(row.cap_mem - row.alloc_mem);
  const totalMem = MBToGB(row.cap_mem);
  return h('div', [
    h('div', [usedCPU, '/', totalCPU, ' CPU']),
    h('div', [usedMem, '/', totalMem, 'GB Memory']),
    h('div', [row.alloc_npu, ' NPU']),
  ]);
}
</script>

<style scoped>
</style>