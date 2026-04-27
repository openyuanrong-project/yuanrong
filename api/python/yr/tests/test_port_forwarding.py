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
import unittest

from yr.config import InvokeOptions, PortForwarding
from yr.port_forwarding import parse_port_forwardings, NETWORK_KEY


class TestPortForwarding(unittest.TestCase):

    def test_single_port_forwarding(self):
        """Single port forwarding should produce correct JSON in createOptions."""
        opt = InvokeOptions()
        opt.port_forwardings = [PortForwarding(port=8080, protocol="TCP")]
        create_opt = parse_port_forwardings(opt)
        self.assertIn(NETWORK_KEY, create_opt)
        parsed = json.loads(create_opt[NETWORK_KEY])
        self.assertEqual(parsed, {"portForwardings": [{"port": 8080, "protocol": "TCP"}]})

    def test_multiple_port_forwardings(self):
        """Multiple port forwardings should all appear in the JSON."""
        opt = InvokeOptions()
        opt.port_forwardings = [
            PortForwarding(port=8080, protocol="TCP"),
            PortForwarding(port=9090, protocol="UDP"),
            PortForwarding(port=3000, protocol="TCP"),
        ]
        create_opt = parse_port_forwardings(opt)
        parsed = json.loads(create_opt[NETWORK_KEY])
        self.assertEqual(len(parsed["portForwardings"]), 3)
        self.assertEqual(parsed["portForwardings"][0], {"port": 8080, "protocol": "TCP"})
        self.assertEqual(parsed["portForwardings"][1], {"port": 9090, "protocol": "UDP"})
        self.assertEqual(parsed["portForwardings"][2], {"port": 3000, "protocol": "TCP"})

    def test_empty_port_forwardings(self):
        """Empty port_forwardings should return empty dict (no network key)."""
        opt = InvokeOptions()
        opt.port_forwardings = []
        create_opt = parse_port_forwardings(opt)
        self.assertEqual(create_opt, {})

    def test_none_port_forwardings(self):
        """Default (None) port_forwardings should return empty dict."""
        opt = InvokeOptions()
        create_opt = parse_port_forwardings(opt)
        self.assertEqual(create_opt, {})

    def test_default_protocol_is_tcp(self):
        """PortForwarding default protocol should be TCP."""
        pf = PortForwarding(port=8080)
        self.assertEqual(pf.protocol, "TCP")

    def test_invalid_port_type(self):
        """Non-int port should raise TypeError."""
        opt = InvokeOptions()
        opt.port_forwardings = [PortForwarding(port="8080", protocol="TCP")]
        with self.assertRaises(TypeError):
            parse_port_forwardings(opt)

    def test_invalid_protocol_type(self):
        """Non-str protocol should raise TypeError."""
        opt = InvokeOptions()
        opt.port_forwardings = [PortForwarding(port=8080, protocol=123)]
        with self.assertRaises(TypeError):
            parse_port_forwardings(opt)

    def test_invalid_port_range(self):
        """Port out of valid range should raise ValueError."""
        opt = InvokeOptions()
        opt.port_forwardings = [PortForwarding(port=0, protocol="TCP")]
        with self.assertRaises(ValueError):
            parse_port_forwardings(opt)

        opt.port_forwardings = [PortForwarding(port=70000, protocol="TCP")]
        with self.assertRaises(ValueError):
            parse_port_forwardings(opt)

    def test_invalid_protocol_value(self):
        """Unsupported protocol should raise ValueError."""
        opt = InvokeOptions()
        opt.port_forwardings = [PortForwarding(port=8080, protocol="SCTP")]
        with self.assertRaises(ValueError):
            parse_port_forwardings(opt)


if __name__ == '__main__':
    unittest.main()
