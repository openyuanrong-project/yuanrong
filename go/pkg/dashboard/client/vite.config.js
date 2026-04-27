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

import { fileURLToPath } from 'node:url';
import path from 'path';
import vue from '@vitejs/plugin-vue';
import { mergeConfig } from 'vite';
import { configDefaults, defineConfig } from 'vitest/config';
import viteConfig from './vite.config';
import importPlugin from '@opentiny/vue-vite-import';

export default mergeConfig(
    viteConfig,
    defineConfig({
      plugins: [
        vue(),
        importPlugin(
            [
              {
                libraryName: '@opentiny/vue'
              },
              {
                libraryName: `@opentiny/vue-icon`,
                customName: (name) => {
                  return `@opentiny/vue-icon/lib/${name.replace(/^icon-/, '')}.js`
                }
              }
            ],
            'pc'
        )
      ],
      resolve: {
        alias: {
          '@': path.resolve(__dirname, './src'),
        },
      },
      test: {
        environment: 'jsdom',
        exclude: [...configDefaults.exclude, 'e2e/*'],
        root: fileURLToPath(new URL('./', import.meta.url)),
        transformMode: {
          web: [/\.[jt]sx$/]
        },
        coverage: {
          provider: 'v8',
          reporter: ['text', 'json', 'html'],
        },
        deps: {
          inline: [
            '@opentiny/vue',
            '@opentiny/vue-renderless',
            '@opentiny/vue-common',
            '@opentiny/vue-icon',
            '@opentiny/vue-theme',
          ]
        },
        threads: false,
        environmentOptions: {
          jsdom: {
            resources: 'usable'
          }
        }
      }
    })
)