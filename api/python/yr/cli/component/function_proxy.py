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

import json
import logging
from pathlib import Path

from yr.cli.component.base import ComponentLauncher
from yr.cli.const import DEFAULT_LOG_DIR

logger = logging.getLogger(__name__)


class FunctionProxyLauncher(ComponentLauncher):
    def health_check(self) -> bool:
        logger.info(f"{self.name}: performing custom health check")
        args_config = self.resolver.rendered_config[self.component_config.name]["args"]
        node_id = self.resolver.rendered_config[self.name]["args"][
            "node_id"
        ]  # this should be already set in prepare_config
        pid = self.component_config.pid
        endpoint = self.resolver.rendered_config[self.component_config.name]["health_check"]["endpoint"]
        logger.info(
            f"{self.name}: health check endpoint: {endpoint}, node_id: {node_id}, pid: {pid}",
        )
        base_path = args_config.get("ssl_base_path", "")
        ca_path = args_config.get("ssl_root_file", "")
        cert_file = args_config.get("ssl_cert_file", "")
        key_file = args_config.get("ssl_key_file", "")
        if base_path and ca_path and cert_file and key_file:
            tls_cacert = Path(base_path) / ca_path
            tls_cert = Path(base_path) / cert_file
            tls_key = Path(base_path) / key_file
        else:
            tls_cacert = None
            tls_cert = None
            tls_key = None
        return self._check_http_or_https_health(
            node_id=node_id,
            pid=str(pid),
            tls_cacert=tls_cacert,
            tls_cert=tls_cert,
            tls_key=tls_key,
        )

    def prestart_hook(self) -> None:
        logger.info(f"{self.name}: prestart hook executing")
        log_config = self.resolver.rendered_config[self.component_config.name]["args"]["log_config"]
        fs_log_path = json.loads(log_config).get("filepath", f"{DEFAULT_LOG_DIR}/function_system")
        Path(fs_log_path).mkdir(parents=True, exist_ok=True)
