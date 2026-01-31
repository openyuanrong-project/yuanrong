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
  <common-card>
    <template v-slot:card-title>
      Log Files
    </template>
    <template v-slot:card-content>
      <tiny-search class="filename-search" v-model="searchValue" placeholder="Filename"
                   @search="filesFilter" is-enter-search></tiny-search>
      <div class="margin-left20" v-for="file in files" :key="file">
        <tiny-link :href="curHash + '/' + file">{{ file }}</tiny-link>
        <tiny-button @click="DownloadLogHandler(file)" class="no-border" :icon="IconDownload" size="small" />
      </div>
    </template>
  </common-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { TinyButton, TinyLink, TinySearch } from '@opentiny/vue';
import { iconDownload } from '@opentiny/vue-icon';
import { ListLogsAPI } from '@/api/api';
import BreadcrumbComponent from '@/components/breadcrumb-component.vue';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';
import { DownloadLogHandler } from "@/utils/downloadLogHandler";

const curHash = window.location.hash;
const IconDownload = iconDownload();
const searchValue = ref('');
const allFiles = ref([]);
const files = ref([]);

onMounted(() => {
  initScrollerItem();
})

function initScrollerItem() {
  ListLogsAPI("").then((res: ListLogsRes) => {
    if (!res?.data) {
      throw '';
    }
    for (let [key, v] of Object.entries(res.data)) {
      if (key == curHash.split('/')[2]) {
        (v as string[]).sort();
        allFiles.value = v;
        files.value = v;
        return;
      }
    }
  }).catch((err: any)=>{
    WarningNotify('ListLogsAPI', err);
  });
}

function filesFilter(_, v: string){
  files.value = allFiles.value.filter((file) => {
    return file.includes(v);
  });
}
</script>

<style scoped>
.filename-search {
  margin: 10px 0 0 20px;
  width: 510px;
}
</style>