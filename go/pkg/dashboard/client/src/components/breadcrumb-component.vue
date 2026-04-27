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
  <tiny-breadcrumb>
    <tiny-breadcrumb-item v-for="(path, i) in pathArr" :key="path">
      <a :href="'#/' + pathArr.slice(0, i+1).join('/')"> {{ path }} </a>
    </tiny-breadcrumb-item>
    <tiny-breadcrumb-item> {{ id }} </tiny-breadcrumb-item>
  </tiny-breadcrumb>
</template>

<script setup lang="ts">
import { TinyBreadcrumb, TinyBreadcrumbItem } from '@opentiny/vue';

const pathArr = window.location.hash.split('/');
const id = pathArr.pop();
pathArr.shift();
</script>

<style scoped>
.tiny-breadcrumb {
  --tv-Breadcrumb-font-size: 16px;
  margin-top: 12px;
  margin-left: 32px;
}
</style>


