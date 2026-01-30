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
      <div class="font-size20 no-block">
        Log: {{ filename }}
      </div>
      <tiny-button @click="DownloadLogHandler(filename)" class="no-border" :icon="IconDownload" />
    </template>
    <template v-slot:card-content>
      <div class="label">Start Line:</div>
      <tiny-input class="line-input" type="number" v-model="startLine" size="mini" :min="1"></tiny-input>
      <div class="label margin-left20">End Line:</div>
      <tiny-input class="line-input" type="number" v-model="endLine" size="mini" :min="1"></tiny-input>
      <tiny-button class="margin-left20" :icon="IconSearch" circle size="mini" @click="initScrollerItem"></tiny-button>
      <div class="label margin-left20 color-blue" :style="tipStyle">[Tip]This log displays a maximum of 5000 lines.</div>
      <DynamicScroller
          :items="items"
          :min-item-size="1"
          :style="{ height: scrollerHeight + 'px' }"
          class="scroller"
          v-slot="{ item, index, active }"
      >
        <DynamicScrollerItem
            :item="item"
            :active="active"
            :size-dependencies="[item.content]"
            :data-index="index"
        >
          <div class="item">
            <div class="icon">{{ item.id }}</div>
            {{ item.line }}
          </div>
        </DynamicScrollerItem>
      </DynamicScroller>
    </template>
  </common-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { TinyInput, TinyButton } from '@opentiny/vue';
import { iconDownload, iconSearch } from '@opentiny/vue-icon';
import { GetLogByFilenameAPI } from '@/api/api';
import CommonCard from '@/components/common-card.vue';
import { WarningNotify } from '@/components/warning-notify';
import { DownloadLogHandler } from "@/utils/downloadLogHandler";
import { LogSWR } from '@/utils/swr';

const NUM_ERROR = 'The line should be greater than 0';
const MAX_LINES = 5000;
const IconSearch = iconSearch();
const IconDownload = iconDownload();
const startLine = ref(1);
const endLine = ref(MAX_LINES);
const scrollerHeight = ref(300);
const items = ref([{
  id: 1,
  line: '',
}]);
const redTipStyle = { color: 'red' };
const tipStyle = ref({});
const { filename, scrollerH } = defineProps(['filename', 'scrollerH']);

onMounted(() => {
  if (scrollerH) {
    scrollerHeight.value = scrollerH;
  }
  LogSWR(initScrollerItem);
})

function initScrollerItem() {
  if (startLine.value <= 0 || endLine.value <= 0) {
    WarningNotify('line num', NUM_ERROR);
    return ;
  }
  if (endLine.value - startLine.value > MAX_LINES) {
    tipStyle.value = redTipStyle;
    return ;
  } else {
    tipStyle.value = {};
  }
  GetLogByFilenameAPI(filename, startLine.value - 1, endLine.value).then((res: string) => {
    const content = res.toString().trimEnd().split('\n');
    if (content.length >= 1 && content[0] === '<!doctype html>') {
      throw '';
    }
    items.value = Array.from({length: content.length}, (_, i) => ({
      id: i + Number(startLine.value),
      line: content[i],
    }))
  }).catch((err: any)=>{
    WarningNotify('GetLogByFilenameAPI', err?.response?.data?.message);
  });
}
</script>

<style>
.scroller {
  margin-top: 10px;
}

.item {
  display: inline-block;
  box-sizing: border-box;
  font-size: 12px;
  font-weight: 400;
  font-family: var(--tv-font-family);
}

.vue-recycle-scroller__item-view {
  line-height: 16px;
}

.icon {
  display: inline-block;
  width: 30px;
  user-select: none;
  color: #999999;
}

.line-input {
  width: 80px;
}

.label {
  display: inline-block;
  font-size: 14px;
  margin-right: 10px;
}
</style>