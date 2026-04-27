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

"""Unit tests for sandbox gateway URL helpers (sandbox / port-forwarding feature)."""

from unittest import TestCase, main

from yr.sandbox.sandbox import _build_gateway_url, _sanitize_instance_id


class TestSandboxGateway(TestCase):
    def test_sanitize_instance_id_at_sign(self):
        self.assertEqual(_sanitize_instance_id("a@b"), "a-at-b")
        self.assertEqual(_sanitize_instance_id("x@y@z"), "x-at-y-at-z")

    def test_sanitize_instance_id_special_chars(self):
        self.assertEqual(_sanitize_instance_id("a/b_c.d"), "a-b-c-d")

    def test_sanitize_instance_id_truncates(self):
        long_id = "x" * 250
        out = _sanitize_instance_id(long_id)
        self.assertEqual(len(out), 200)

    def test_build_gateway_url(self):
        url = _build_gateway_url("id@1", 8080, "gw.example.com")
        self.assertEqual(url, "https://gw.example.com/id-at-1/8080")

    def test_build_gateway_url_with_path(self):
        url = _build_gateway_url("inst", 9000, "h", path="api")
        self.assertEqual(url, "https://h/inst/9000/api")
        url2 = _build_gateway_url("inst", 9000, "h", path="/ready")
        self.assertEqual(url2, "https://h/inst/9000/ready")


if __name__ == "__main__":
    main()
