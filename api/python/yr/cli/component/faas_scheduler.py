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


class FaaSSchedulerLauncher(ComponentLauncher):
    def prestart_hook(self) -> None:
        logger.info(f"{self.name}: prestart hook executing")
        src = self.resolver.rendered_config["values"][self.name]["config_path"]
        dest = Path(self.resolver.rendered_config[self.name]["env"]["INIT_ARGS_FILE_PATH"]).resolve()
        self.patch_init_scheduler_args(src, dest)

    def patch_init_scheduler_args(self, src: Path, dest: Path) -> None:
        src = Path(src).resolve()
        text = src.read_text()

        config = self.resolver.rendered_config
        values = config["values"]
        # attention: etcd address getting from function_proxy component
        etcd_addrs = config["function_proxy"]["args"]["etcd_address"]
        etcd_addr = etcd_addrs.split(",")
        etcd_addrs = '","'.join(etcd_addr)
        etcd_auth_type = values["etcd"].get("auth_type", "Noauth")
        etcd_table_prefix = values["etcd"].get("table_prefix", "")

        ssl_enable = str(values["fs"]["tls"].get("enable", "false")).lower()
        ssl_base_path = values["fs"]["tls"].get("base_path", "")
        etcd_ssl_base_path = values["etcd"]["auth"].get("base_path", "")

        if etcd_auth_type == "TLS":
            etcd_ca = f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('ca_file', '')}"
            etcd_cert = f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('client_cert_file', '')}"
            etcd_key = f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('client_key_file', '')}"
            user_pass_phrase = values["etcd"]["auth"].get("pass_phrase", "")
            pass_phrase = f"{etcd_ssl_base_path}/{user_pass_phrase}" if user_pass_phrase else ""
        else:
            etcd_ca = ""
            etcd_cert = ""
            etcd_key = ""
            pass_phrase = ""

        replacements = {
            "{etcdAddr}": etcd_addrs,
            "{sslEnable}": ssl_enable,
            "{etcdAuthType}": etcd_auth_type,
            "{azPrefix}": etcd_table_prefix,
            "{sslBasePath}": ssl_base_path,
            "{etcdCAFile}": etcd_ca,
            "{etcdCertFile}": etcd_cert,
            "{etcdKeyFile}": etcd_key,
            "{passphraseFile}": pass_phrase,
        }

        for placeholder, value in replacements.items():
            text = text.replace(placeholder, value)

        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_text(text)
