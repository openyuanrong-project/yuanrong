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


class DashboardLauncher(ComponentLauncher):
    def prestart_hook(self) -> None:
        logger.info(f"{self.name}: prestart hook executing")
        src = self.resolver.rendered_config[self.name]["src_config_path"]
        dest = Path(self.resolver.rendered_config[self.name]["args"]["config_path"]).resolve()
        self.patch_dashboard_config(src, dest)

    def patch_dashboard_config(self, src: Path, dest: Path) -> None:
        src = Path(src).resolve()
        text = src.read_text()

        config = self.resolver.rendered_config
        values = config["values"]
        dashboard_values = config["values"]["dashboard"]

        ip_address = dashboard_values["ip"]
        port = dashboard_values["port"]
        static_path = values["yr_package_path"] + "/functionsystem/bin/client/dist"
        grpc_ip = ip_address
        grpc_port = dashboard_values["grpc_port"]
        function_master_addr = f"{values['function_master']['ip']}:{values['function_master']['global_scheduler_port']}"
        frontend_addr = (
            f"{config['frontend']['ip']}:{config['frontend']['port']}"
            if config["mode"]["master"].get("frontend", False) or config["mode"]["agent"].get("frontend", False)
            else ""
        )
        prometheus_addr = dashboard_values["prometheus"]["address"]

        # attention: etcd address getting from function_proxy component
        etcd_addrs = config["function_proxy"]["args"]["etcd_address"]
        etcd_addr = etcd_addrs.split(",")
        etcd_addrs = '","'.join(etcd_addr)
        etcd_auth_type = values["etcd"].get("auth_type", "Noauth")
        etcd_table_prefix = values["etcd"].get("table_prefix", "")
        etcd_ssl_base_path = values["etcd"]["auth"].get("base_path", "")
        etcd_ca = (
            f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('ca_file', '')}"
            if values["etcd"]["auth"].get("ca_file") and etcd_auth_type == "TLS" and etcd_ssl_base_path
            else ""
        )
        etcd_cert = (
            f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('client_cert_file', '')}"
            if values["etcd"]["auth"].get("client_cert_file") and etcd_auth_type == "TLS" and etcd_ssl_base_path
            else ""
        )
        etcd_key = (
            f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('client_key_file', '')}"
            if values["etcd"]["auth"].get("client_key_file") and etcd_auth_type == "TLS" and etcd_ssl_base_path
            else ""
        )

        fs_tls_enable = str(values["fs"]["tls"].get("enable", "false")).lower()
        fs_tls_base_path = values["fs"]["tls"].get("base_path", "")
        fs_ca = (
            f"{fs_tls_base_path}/{values['fs']['tls'].get('ca_file', '')}"
            if values["fs"]["tls"].get("ca_file") and fs_tls_enable == "true"
            else ""
        )
        fs_cert = (
            f"{fs_tls_base_path}/{values['fs']['tls'].get('cert_file', '')}"
            if values["fs"]["tls"].get("cert_file") and fs_tls_enable == "true"
            else ""
        )
        fs_key = (
            f"{fs_tls_base_path}/{values['fs']['tls'].get('key_file', '')}"
            if values["fs"]["tls"].get("key_file") and fs_tls_enable == "true"
            else ""
        )

        prometheus_auth = dashboard_values["prometheus"]["auth"]
        prometheus_auth_enable = str(prometheus_auth.get("enable", "false")).lower()
        base_path = prometheus_auth.get("base_path", "")
        prometheus_ca = (
            f"{base_path}/{prometheus_auth.get('ca_file', '')}"
            if prometheus_auth.get("ca_file") and prometheus_auth_enable == "true"
            else ""
        )
        prometheus_cert = (
            f"{base_path}/{prometheus_auth.get('cert_file', '')}"
            if prometheus_auth.get("ca_file") and prometheus_auth_enable == "true"
            else ""
        )
        prometheus_key = (
            f"{base_path}/{prometheus_auth.get('key_file', '')}"
            if prometheus_auth.get("ca_file") and prometheus_auth_enable == "true"
            else ""
        )

        dashboard_ssl_enable = str(dashboard_values["auth"].get("enable", "false")).lower()
        dashboard_cert = dashboard_values["auth"].get("cert_file", "") if dashboard_ssl_enable == "true" else ""
        dashboard_key = dashboard_values["auth"].get("key_file", "") if dashboard_ssl_enable == "true" else ""

        replacements = {
            "{ip}": ip_address,
            "{port}": str(port),
            "{staticPath}": static_path,
            "{grpcIP}": grpc_ip,
            "{grpcPort}": str(grpc_port),
            "{functionMasterAddr}": function_master_addr,
            "{frontendAddr}": frontend_addr,
            "{prometheusAddr}": prometheus_addr,
            "{etcdAddr}": etcd_addrs,
            "{etcdSsl}": fs_tls_enable,
            "{etcdAuthType}": etcd_auth_type,
            "{azPrefix}": etcd_table_prefix,
            "{etcdCAFile}": etcd_ca,
            "{etcdCertFile}": etcd_cert,
            "{etcdKeyFile}": etcd_key,
            "{functionSystemSsl}": fs_tls_enable,
            "{functionSystemCAFile}": fs_ca,
            "{functionSystemCertFile}": fs_cert,
            "{functionSystemKeyFile}": fs_key,
            "{prometheusSsl}": prometheus_auth_enable,
            "{prometheusCAFile}": prometheus_ca,
            "{prometheusCertFile}": prometheus_cert,
            "{prometheusKeyFile}": prometheus_key,
            "{dashboardSsl}": dashboard_ssl_enable,
            "{dashboardCertFile}": dashboard_cert,
            "{dashboardKeyFile}": dashboard_key,
        }

        for placeholder, value in replacements.items():
            text = text.replace(placeholder, value)

        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_text(text)
