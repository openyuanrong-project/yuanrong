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
import ssl
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

from yr.cli.config import ConfigResolver
from yr.cli.const import StartMode
from yr.cli.utils import service_discovery_from_function_master, services_to_overrides

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class ServiceDiscoveryOptions:
    function_master_addr: str
    function_system_ca_cert: Optional[Path]
    function_system_client_cert: Optional[Path]
    function_system_client_key: Optional[Path]


def resolve_overrides_from_function_master(
    config_path: Path,
    cli_dir: Path,
    mode: StartMode,
    overrides: tuple[str, ...],
    function_master_addr: str,
) -> tuple[str, ...]:
    if mode != StartMode.AGENT:
        raise ValueError(
            "function_master_address is only applicable in agent mode. Ignoring it since we're in master mode."
        )

    function_system_ca_cert, function_system_client_cert, function_system_client_key = (
        _fill_service_discovery_tls_paths_from_config(
            config_path=config_path,
            cli_dir=cli_dir,
            mode=mode,
            overrides=overrides,
            function_master_addr=function_master_addr,
        )
    )

    discovery_opts = ServiceDiscoveryOptions(
        function_master_addr=function_master_addr,
        function_system_ca_cert=function_system_ca_cert,
        function_system_client_cert=function_system_client_cert,
        function_system_client_key=function_system_client_key,
    )
    return overrides + _resolve_service_discovery_overrides(discovery_opts)


def _resolve_tls_path(path_value: Optional[str], base_path: Optional[str]) -> Optional[Path]:
    if not isinstance(path_value, str):
        return None
    value = path_value.strip()
    if not value:
        return None
    path = Path(value)
    base = base_path.strip() if isinstance(base_path, str) else ""
    if not path.is_absolute() and base:
        path = Path(base) / path
    return path


def _fill_service_discovery_tls_paths_from_config(
    config_path: Path,
    cli_dir: Path,
    mode: StartMode,
    overrides: tuple[str, ...],
    function_master_addr: str,
) -> tuple[Optional[Path], Optional[Path], Optional[Path]]:
    if not function_master_addr.lower().startswith("https://"):
        return None, None, None

    try:
        cfg = ConfigResolver(config_path, cli_dir, mode=mode, overrides=overrides)
    except Exception as e:
        raise ValueError(f"Failed to load config for resolving TLS paths for service discovery: {e}")
    tls_cfg = cfg.rendered_config.get("values", {}).get("fs", {}).get("tls", {})
    if not isinstance(tls_cfg, dict):
        return None, None, None

    base_path = tls_cfg.get("base_path")
    resolved_ca = _resolve_tls_path(tls_cfg.get("ca_file"), base_path)
    resolved_cert = _resolve_tls_path(tls_cfg.get("cert_file"), base_path)
    resolved_key = _resolve_tls_path(tls_cfg.get("key_file"), base_path)

    missing_fields: list[str] = []
    if not resolved_ca or not resolved_ca.exists():
        missing_fields.append("values.fs.tls.ca_file")
    if not resolved_cert or not resolved_cert.exists():
        missing_fields.append("values.fs.tls.cert_file")
    if not resolved_key or not resolved_key.exists():
        missing_fields.append("values.fs.tls.key_file")
    if missing_fields:
        raise ValueError(
            "HTTPS function_master_address requires non-empty TLS cert settings in merged config (-s/--config): "
            + ", ".join(missing_fields)
        )

    return resolved_ca, resolved_cert, resolved_key


def _build_service_discovery_tls_context(
    opts: ServiceDiscoveryOptions,
) -> Optional[ssl.SSLContext]:
    if not opts.function_master_addr.lower().startswith("https://"):
        return None
    if not opts.function_system_client_cert or not opts.function_system_client_key:
        raise ValueError(
            "HTTPS function_master_address requires values.fs.tls.cert_file and "
            "values.fs.tls.key_file from merged config (-s/--config) for mTLS service discovery.",
        )

    ssl_context = ssl.create_default_context(cafile=opts.function_system_ca_cert)
    ssl_context.check_hostname = False
    ssl_context.verify_mode = ssl.CERT_REQUIRED
    ssl_context.load_cert_chain(
        certfile=opts.function_system_client_cert,
        keyfile=opts.function_system_client_key,
    )
    return ssl_context


def _resolve_service_discovery_overrides(
    opts: ServiceDiscoveryOptions,
) -> tuple[str, ...]:
    ssl_context = _build_service_discovery_tls_context(opts)
    services = service_discovery_from_function_master(
        opts.function_master_addr,
        ssl_context=ssl_context,
    )
    logger.debug(f"Derived config overrides from service discovery: {services}")
    resolved = tuple(services_to_overrides(services))
    logger.debug(f"Derived override list from service discovery: {resolved}")
    return resolved
