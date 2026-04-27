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
  <div class="preview">
    <tiny-nav-menu :data="menuData" overflow="hidden">
      <template #logo>
        <a href="#/overview">
          <img src="/favicon.ico" alt="YuanRong logo" class="yr-logo"/>
        </a>
      </template>
      <template #toolbar>
         <p class="version">{{ config.version }}</p>
      </template>
    </tiny-nav-menu>
    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { TinyNavMenu } from '@opentiny/vue';
import Logo from '@/../public/favicon.ico';
import config from '@/config/config.json';

const menuData = ref([
  {
    title: 'Overview',
    url: '/overview',
    id: '1',
  },
  {
    title: 'Cluster',
    url: '/cluster',
    id: '2',
  },
  {
    title: 'Jobs',
    url: '/jobs',
    id: '3',
  },
  {
    title: 'Instances',
    url: '/instances',
    id: '4',
  },
  {
    title: 'Logs',
    url: '/logs',
    id: '5',
  },
]);

const TinyIconTotal = Logo;
</script>

<style scoped>
.preview {
  min-height: 450px;
  font-weight: bold;
}

.tiny-nav-menu {
  --tv-NavMenu-item-selected-underline-bg-color: var(--tv-color-text-active);
  --tv-NavMenu-item-li-margin: 20px;
}

.yr-logo {
  height: 34px;
  margin-top: 8px;
}

.version {
  font-weight: normal;
  font-size: 15px;
}
</style>
