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


class MetaServiceLauncher(ComponentLauncher):
    def prestart_hook(self) -> None:
        logger.info(f"{self.name}: prestart hook executing")
        config_src = self.resolver.rendered_config["values"][self.name]["config_path"]
        config_dest = Path(self.resolver.rendered_config[self.name]["args"]["config_path"]).resolve()
        self.patch_meta_service_config(config_src, config_dest)

        log_src = self.resolver.rendered_config["values"][self.name]["log_config_path"]
        log_dest = Path(self.resolver.rendered_config[self.name]["args"]["log_config_path"]).resolve()
        self.patch_meta_service_log(log_src, log_dest)

    def patch_meta_service_config(self, src: Path, dest: Path) -> None:
        src = Path(src).resolve()
        text = src.read_text()

        config = self.resolver.rendered_config
        values = config["values"]
        service_args = config[self.component_config.name]["args"]
        service_values = values["meta_service"]
        # attention: etcd address getting from function_proxy component
        etcd_addrs = config["function_proxy"]["args"]["etcd_address"]
        etcd_addr = etcd_addrs.split(",")
        etcd_addrs = '","'.join(etcd_addr)
        ip_address = service_values["ip"]
        port = service_values["port"]
        etcd_auth_type = values["etcd"].get("auth_type", "Noauth")
        etcd_table_prefix = values["etcd"].get("table_prefix", "")

        ssl_enable = str(values["fs"]["tls"].get("enable", "false")).lower()
        scc_enable = str(service_values.get("scc_enable", "false")).lower()
        ssl_base_path = values["fs"]["tls"].get("base_path", "")
        scc_base_path = service_args.get("scc_base_path", "")
        etcd_ssl_base_path = values["etcd"]["auth"].get("base_path", "")

        if ssl_enable == "true":
            root_ca = f"{ssl_base_path}/{values['fs']['tls'].get('ca_file', '')}"
            module_cert = f"{ssl_base_path}/{values['fs']['tls'].get('cert_file', '')}"
            module_key = f"{ssl_base_path}/{values['fs']['tls'].get('key_file', '')}"
            pwd_file = f"{ssl_base_path}/{values['fs']['tls'].get('pwd_file', '')}"
            if etcd_ssl_base_path:
                etcd_ca = f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('ca_file', '')}"
                etcd_cert = f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('client_cert_file', '')}"
                etcd_key = f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('client_key_file', '')}"
                pass_phrase = f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('pass_phrase', '')}"
        else:
            etcd_ca = ""
            etcd_cert = ""
            etcd_key = ""
            pass_phrase = ""
            root_ca = ""
            module_cert = ""
            module_key = ""
            pwd_file = ""

        replacements = {
            "{ip}": ip_address,
            "{port}": str(port),
            "{etcdAddr}": etcd_addrs,
            "{sslEnable}": ssl_enable,
            "{azPrefix}": etcd_table_prefix,
            "{etcdAuthType}": etcd_auth_type,
            "{clusters}": service_values.get("clusters", ""),
            "{sccEnable}": scc_enable,
            "{sccBasePath}": scc_base_path,
            # if ssl enable
            "{rootCAFile}": root_ca,
            "{moduleCertFile}": module_cert,
            "{moduleKeyFile}": module_key,
            "{pwdFile}": pwd_file,
            # if etcd ssl enable
            "{etcdCAFile}": etcd_ca,
            "{etcdCertFile}": etcd_cert,
            "{etcdKeyFile}": etcd_key,
            "{passphraseFile}": pass_phrase,
            "{sslDecryptTool}": service_values.get(
                "ssl_decrypt_tool",
                "SCC",
            ),
        }

        for placeholder, value in replacements.items():
            text = text.replace(placeholder, value)

        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_text(text)

    def patch_meta_service_log(self, src: Path, dest: Path) -> None:
        src = Path(src).resolve()
        text = src.read_text()

        config = self.resolver.rendered_config
        values = config["values"]
        log_path = values["fs"]["log"]["path"]
        log_level = values["fs"]["log"]["level"]

        replacements = {
            "{logConfigPath}": log_path,
            "{logLevel}": log_level,
        }

        for placeholder, value in replacements.items():
            text = text.replace(placeholder, value)

        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_text(text)
