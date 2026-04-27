#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import logging
from pathlib import Path

from yr.cli.component.base import ComponentLauncher

logger = logging.getLogger(__name__)


class DsWorkerLauncher(ComponentLauncher):
    def health_check(self) -> bool:
        logger.info(f"{self.name}: performing custom health check")
        return self._check_file_health()

    # create rocksdb, socket and log dir before start
    def prestart_hook(self) -> None:
        logger.info(f"{self.name}: prestart hook executing")
        # these should be already set in prepare_config
        rocksdb_dir = self.resolver.rendered_config[self.name]["args"]["rocksdb_store_dir"]
        socket_dir = self.resolver.rendered_config[self.name]["args"]["unix_domain_socket_dir"]
        logs_dir = self.resolver.rendered_config[self.name]["args"]["log_dir"]
        dirs = [rocksdb_dir, socket_dir, logs_dir]
        if self.resolver.rendered_config[self.name].get("spill", False):
            spill_dir = self.resolver.rendered_config[self.name]["args"].get("spill_directory")
            if spill_dir:
                dirs.append(spill_dir)
        for dir_path in dirs:
            Path(dir_path).mkdir(parents=True, exist_ok=True)

        logger.info(f"Created directories for {self.name}: {dirs}")
