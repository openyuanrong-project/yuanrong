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
import tempfile
from pathlib import Path
from unittest import TestCase, main
from unittest.mock import Mock, patch, MagicMock

from yr.cli.component.faas_frontend import FaaSFrontendLauncher
from yr.cli.component.faas_scheduler import FaaSSchedulerLauncher

logger = logging.getLogger(__name__)


class MockComponentConfig:
    def __init__(self, name="faas_frontend"):
        self.name = name
        self.enabled = True
        self.args = {}
        self.env_vars = {}


class TestFaaSFrontendLauncher(TestCase):
    def setUp(self):
        self.launcher = FaaSFrontendLauncher(name="faas_frontend", resolver=Mock())
        self.launcher.component_config = MockComponentConfig("faas_frontend")

    def test_prestart_hook(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            src_file = Path(tmpdir) / "config.template"
            dest_file = Path(tmpdir) / "config.ini"
            
            src_file.write_text("{etcdAddr}\n{faas_frontend_http_ip}\n{faas_frontend_http_port}")
            
            mock_resolver = Mock()
            mock_resolver.rendered_config = {
                "values": {
                    "faas_frontend": {
                        "ip": "0.0.0.0",
                        "port": 8888,
                        "scc_enable": "false",
                    },
                    "etcd": {
                        "auth_type": "Noauth",
                        "table_prefix": "",
                        "auth": {},
                    },
                    "fs": {
                        "tls": {
                            "enable": "false",
                            "base_path": "",
                        }
                    }
                },
                "function_proxy": {
                    "args": {
                        "etcd_address": "127.0.0.1:2379"
                    }
                },
                "faas_frontend": {
                    "args": {},
                    "env": {
                        "INIT_ARGS_FILE_PATH": str(dest_file)
                    }
                }
            }
            mock_resolver.rendered_config["values"]["faas_frontend"]["config_path"] = str(src_file)
            self.launcher.resolver = mock_resolver
            
            self.launcher.prestart_hook()
            
            self.assertTrue(dest_file.exists())
            content = dest_file.read_text()
            self.assertIn("127.0.0.1:2379", content)
            self.assertIn("0.0.0.0", content)
            self.assertIn("8888", content)

    def test_patch_init_frontend_args_with_tls(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            src_file = Path(tmpdir) / "config.template"
            dest_file = Path(tmpdir) / "config.ini"
            
            src_file.write_text("{etcdAddr}\n{etcdAuthType}\n{etcdCAFile}\n{etcdCertFile}\n{etcdKeyFile}")
            
            mock_resolver = Mock()
            mock_resolver.rendered_config = {
                "values": {
                    "faas_frontend": {
                        "ip": "0.0.0.0",
                        "port": 8888,
                        "scc_enable": "false",
                        "config_path": str(src_file),
                    },
                    "etcd": {
                        "auth_type": "TLS",
                        "table_prefix": "",
                        "auth": {
                            "base_path": "/etc/ssl",
                            "ca_file": "ca.crt",
                            "client_cert_file": "client.crt",
                            "client_key_file": "client.key",
                        },
                    },
                    "fs": {
                        "tls": {
                            "enable": "true",
                            "base_path": "/etc/ssl",
                        }
                    }
                },
                "function_proxy": {
                    "args": {
                        "etcd_address": "127.0.0.1:2379"
                    }
                },
                "faas_frontend": {
                    "args": {},
                    "env": {
                        "INIT_ARGS_FILE_PATH": str(dest_file)
                    }
                }
            }
            self.launcher.resolver = mock_resolver
            
            self.launcher.patch_init_frontend_args(src_file, dest_file)
            
            content = dest_file.read_text()
            self.assertIn("TLS", content)
            self.assertIn("/etc/ssl/ca.crt", content)
            self.assertIn("/etc/ssl/client.crt", content)
            self.assertIn("/etc/ssl/client.key", content)


class TestFaaSSchedulerLauncher(TestCase):
    def setUp(self):
        self.launcher = FaaSSchedulerLauncher(name="faas_scheduler", resolver=Mock())
        self.launcher.component_config = MockComponentConfig("faas_scheduler")

    def test_prestart_hook(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            src_file = Path(tmpdir) / "config.template"
            dest_file = Path(tmpdir) / "config.ini"
            
            src_file.write_text("{etcdAddr}\n{sslEnable}\n{etcdAuthType}")
            
            mock_resolver = Mock()
            mock_resolver.rendered_config = {
                "values": {
                    "etcd": {
                        "auth_type": "Noauth",
                        "table_prefix": "",
                        "auth": {},
                    },
                    "fs": {
                        "tls": {
                            "enable": "false",
                            "base_path": "",
                        }
                    },
                    "faas_scheduler": {
                        "config_path": str(src_file),
                    }
                },
                "function_proxy": {
                    "args": {
                        "etcd_address": "127.0.0.1:2379,127.0.0.1:2380"
                    }
                },
                "faas_scheduler": {
                    "args": {},
                    "env": {
                        "INIT_ARGS_FILE_PATH": str(dest_file)
                    }
                }
            }
            self.launcher.resolver = mock_resolver
            
            self.launcher.prestart_hook()
            
            self.assertTrue(dest_file.exists())
            content = dest_file.read_text()
            self.assertIn("127.0.0.1:2379", content)
            self.assertIn("127.0.0.1:2380", content)
            self.assertIn("Noauth", content)

    def test_patch_init_scheduler_args_with_tls(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            src_file = Path(tmpdir) / "config.template"
            dest_file = Path(tmpdir) / "config.ini"
            
            src_file.write_text("{etcdAddr}\n{etcdAuthType}\n{etcdCAFile}\n{etcdCertFile}\n{etcdKeyFile}")
            
            mock_resolver = Mock()
            mock_resolver.rendered_config = {
                "values": {
                    "etcd": {
                        "auth_type": "TLS",
                        "table_prefix": "",
                        "auth": {
                            "base_path": "/etc/ssl",
                            "ca_file": "ca.crt",
                            "client_cert_file": "client.crt",
                            "client_key_file": "client.key",
                        },
                    },
                    "fs": {
                        "tls": {
                            "enable": "true",
                            "base_path": "/etc/ssl",
                        }
                    },
                    "faas_scheduler": {
                        "config_path": str(src_file),
                    }
                },
                "function_proxy": {
                    "args": {
                        "etcd_address": "127.0.0.1:2379"
                    }
                },
                "faas_scheduler": {
                    "args": {},
                    "env": {
                        "INIT_ARGS_FILE_PATH": str(dest_file)
                    }
                }
            }
            self.launcher.resolver = mock_resolver
            
            self.launcher.patch_init_scheduler_args(src_file, dest_file)
            
            content = dest_file.read_text()
            self.assertIn("TLS", content)
            self.assertIn("/etc/ssl/ca.crt", content)


if __name__ == "__main__":
    main()
