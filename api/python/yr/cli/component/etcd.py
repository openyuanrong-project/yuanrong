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
import subprocess
from pathlib import Path

from yr.cli.component.base import ComponentLauncher

logger = logging.getLogger(__name__)


class ETCDLauncher(ComponentLauncher):
    def prestart_hook(self) -> None:
        logger.info(f"{self.name}: prestart hook executing")
        data_dir = self.resolver.rendered_config[self.component_config.name]["args"]["data-dir"]
        Path(data_dir).mkdir(parents=True, exist_ok=True)

    def health_check(self) -> bool:
        etcdctl_path = self.resolver.rendered_config["etcd"]["health_check"]["etcdctl_path"]
        endpoint = self.resolver.rendered_config["etcd"]["health_check"]["endpoint"]
        if etcdctl_path == "":
            logger.error("etcd: etcdctl_path is not set for health check.")
            return False
        try:
            args = [str(etcdctl_path)]
            if (
                self.resolver.rendered_config["etcd"]["health_check"]["auth_type"] == "TLS"
                and self.resolver.rendered_config["etcd"]["health_check"]["client_cert_file"] != ""
                and self.resolver.rendered_config["etcd"]["health_check"]["client_key_file"] != ""
                and self.resolver.rendered_config["etcd"]["args"]["trusted-ca-file"] != ""
            ):
                args.extend(
                    [
                        "--cacert",
                        self.resolver.rendered_config["etcd"]["args"]["trusted-ca-file"],
                        "--cert",
                        self.resolver.rendered_config["etcd"]["health_check"]["client_cert_file"],
                        "--key",
                        self.resolver.rendered_config["etcd"]["health_check"]["client_key_file"],
                    ],
                )
            args.extend(["--endpoints", endpoint, "endpoint", "health"])

            logger.info(f"{self.name}: health check command: {' '.join(args)}")
            result = subprocess.run(
                args,
                check=False,
                capture_output=True,
                timeout=2,
            )
            if result.returncode != 0:
                logger.warning(
                    f"{self.name}: health check failed: {result.stderr.decode().strip()}",
                )
                return False
        except subprocess.TimeoutExpired:
            logger.warning(f"{self.name}: health check timed out")
            return False
        except Exception:
            logger.exception("Exception during etcd health check")
            return False
        else:
            return True

    def poststart_hook(self) -> None:
        logger.info(f"{self.name}: poststart hook executing")
        endpoint = self.resolver.rendered_config["etcd"]["args"]["listen-client-urls"]
        if self.resolver.rendered_config["etcd"]["health_check"]["auth_type"] == "Noauth":
            etcdctl_path = self.resolver.rendered_config["etcd"]["health_check"]["etcdctl_path"]
            if etcdctl_path == "":
                logger.error("etcd: etcdctl_path is not set for disabling auth.")
                return
            args = [str(etcdctl_path), "--endpoints", endpoint, "auth", "disable"]
            logger.info(f"etcd: disabling etcd auth with command: {' '.join(args)}")

            try:
                result = subprocess.run(
                    args,
                    check=False,
                    capture_output=True,
                    timeout=5,
                )
                if result.returncode != 0:
                    logger.warning(
                        f"Failed to disable etcd auth: {result.stderr.decode().strip()}",
                    )
                else:
                    logger.info("etcd: auth disabled successfully.")
            except Exception:
                logger.exception("etcd: exception while disabling etcd auth")
