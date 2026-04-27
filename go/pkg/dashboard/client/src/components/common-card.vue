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
  <div class="card-wrap">
    <div class="card-box">
      <div class="card-title">
        <slot name="card-title"></slot>
      </div>
      <div class="card-content">
        <slot name="card-content"></slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.card-wrap {
  padding: var(--card--space);
}

.card-box {
  border: 1px solid #e0e0e0;
  border-radius: 15px;
  font-weight: normal;
  padding: 22px 20px;
}

.card-title {
  font-size: 24px;
}

.card-content {
  font-size: 20px;
}
</style>
