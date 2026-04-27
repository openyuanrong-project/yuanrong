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
  <div class="content">
    <tiny-layout>
      <resources-card />
      <tiny-row>
        <tiny-col :span="6">
          <div class="col card-space-right">
            <cluster-card />
          </div>
        </tiny-col>
        <tiny-col :span="6">
          <div class="col">
            <instances-card />
          </div>
        </tiny-col>
      </tiny-row>
    </tiny-layout>
  </div>
</template>

<script setup lang="ts">
import { TinyLayout, TinyRow, TinyCol } from '@opentiny/vue';
import ClusterCard from './components/cluster-card.vue';
import InstancesCard from './components/instances-card.vue';
import ResourcesCard from './components/resources-card.vue';
</script>

<style scoped>
</style>
