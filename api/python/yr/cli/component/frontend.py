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


class FrontendLauncher(ComponentLauncher):
    def prestart_hook(self) -> None:
        logger.info(f"{self.name}: prestart hook executing")
        src = self.resolver.rendered_config[self.name]["src_init_config_path"]
        dest = Path(self.resolver.rendered_config[self.name]["env"]["INIT_ARGS_FILE_PATH"]).resolve()
        self.patch_init_frontend_args(src, dest)

    def patch_init_frontend_args(self, src: Path, dest: Path) -> None:
        src = Path(src).resolve()
        text = src.read_text()

        config = self.resolver.rendered_config
        values = config["values"]
        faas_args = config[self.component_config.name]["args"]
        faas_config = config[self.name]
        # attention: etcd address getting from function_proxy component
        etcd_addrs = config["function_proxy"]["args"]["etcd_address"]
        etcd_addr = etcd_addrs.split(",")
        etcd_addrs = '","'.join(etcd_addr)
        ip_address = faas_config["ip"]
        port = faas_config["port"]
        etcd_auth_type = values["etcd"].get("auth_type", "Noauth")
        etcd_table_prefix = values["etcd"].get("table_prefix", "")

        ssl_enable = str(values["fs"]["tls"].get("enable", "false")).lower()
        frontend_ssl_enable = str(faas_config.get("ssl_enable", "false")).lower()
        client_auth_type = faas_config.get("client_auth_type", "RequireAndVerifyClientCert")
        enable_func_token_auth = str(faas_config.get("enable_function_token_auth", "false")).lower()
        scc_enable = str(faas_config.get("scc_enable", "false")).lower()
        ssl_base_path = values["fs"]["tls"].get("base_path", "")
        scc_base_path = faas_args.get("scc_base_path", "")
        etcd_ssl_base_path = values["etcd"]["auth"].get("base_path", "")

        meta_service_config = config.get("meta_service", {})
        meta_service_address = (
            f"{meta_service_config['ip']}:{meta_service_config['port']}"
            if meta_service_config.get("ip")
            else ""
        )
        iam_server_vals = values.get("iam_server", {})
        iam_server_address = (
            f"{iam_server_vals['ip']}:{iam_server_vals['port']}"
            if iam_server_vals.get("ip")
            else ""
        )

        auth_config = values.get("auth", {})
        auth_provider = auth_config.get("provider", "casdoor")
        if auth_provider == "casdoor":
            casdoor = auth_config.get("casdoor", {})
            auth_enabled = str(casdoor.get("enabled", False)).lower()
            auth_public_url = casdoor.get("public_endpoint") or casdoor.get("endpoint", "")
            auth_internal_url = casdoor.get("endpoint", "")
            auth_realm = casdoor.get("organization", "openyuanrong.org")
            auth_client_id = casdoor.get("client_id", "")
            auth_client_secret = casdoor.get("client_secret", "")
        else:
            keycloak = auth_config.get("keycloak", {})
            auth_enabled = str(keycloak.get("enabled", False)).lower()
            auth_public_url = keycloak.get("url", "")
            auth_internal_url = keycloak.get("internal_url") or keycloak.get("url", "")
            auth_realm = keycloak.get("realm", "yuanrong")
            auth_client_id = keycloak.get("client_id", "frontend")
            auth_client_secret = keycloak.get("client_secret", "")

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
        pass_phrase = (
            f"{etcd_ssl_base_path}/{values['etcd']['auth'].get('pass_phrase', '')}"
            if values["etcd"]["auth"].get("pass_phrase") and etcd_auth_type == "TLS" and etcd_ssl_base_path
            else ""
        )

        # Replace literal string before dict-based substitutions so it applies unconditionally
        text = text.replace("RequireAndVerifyClientCert", client_auth_type)

        replacements = {
            "{etcdAddr}": etcd_addrs,
            "{faas_frontend_http_ip}": ip_address,
            "{faas_frontend_http_port}": str(port),
            "{sslEnable}": ssl_enable,
            "{frontendSslEnable}": frontend_ssl_enable,
            "{enable_func_token_auth}": enable_func_token_auth,
            "{iam_server_address}": iam_server_address,
            "{meta_service_address}": meta_service_address,
            "{sccEnable}": scc_enable,
            "{etcdAuthType}": etcd_auth_type,
            "{azPrefix}": etcd_table_prefix,
            "{sslBasePath}": ssl_base_path,
            "{sccBasePath}": scc_base_path,
            "{etcdCAFile}": etcd_ca,
            "{etcdCertFile}": etcd_cert,
            "{etcdKeyFile}": etcd_key,
            "{passphraseFile}": pass_phrase,
            # Generic auth placeholders (casdoor or keycloak depending on provider)
            "{auth_enabled}": auth_enabled,
            "{auth_public_url}": auth_public_url,
            "{auth_internal_url}": auth_internal_url,
            "{auth_realm}": auth_realm,
            "{auth_client_id}": auth_client_id,
            "{auth_client_secret}": auth_client_secret,
            # Legacy keycloak placeholders — same values as generic auth
            "{keycloak_enabled}": auth_enabled,
            "{keycloak_url}": auth_public_url,
            "{keycloak_internal_url}": auth_internal_url,
            "{keycloak_realm}": auth_realm,
            "{keycloak_client_id}": auth_client_id,
            "{keycloak_client_secret}": auth_client_secret,
        }

        for placeholder, value in replacements.items():
            text = text.replace(placeholder, value)

        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_text(text)
