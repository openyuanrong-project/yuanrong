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

import os
import tempfile
import unittest
import logging
from unittest.mock import patch

from yr.common import utils

logger = logging.getLogger(__name__)


class TestLoadEnvFromFile(unittest.TestCase):
    """Test cases for load_env_from_file function."""

    def setUp(self):
        """Set up test fixtures."""
        # Save original environment variables that might be modified
        self.original_env = os.environ.copy()
        # Clear test environment variables
        test_keys = ["TEST_KEY1", "TEST_KEY2", "TEST_KEY3", "TEST_NUM", "TEST_BOOL", "TEST_QUOTED"]
        for key in test_keys:
            if key in os.environ:
                del os.environ[key]

    def tearDown(self):
        """Clean up after tests."""
        # Restore original environment
        os.environ.clear()
        os.environ.update(self.original_env)

    @patch("yr.log.get_logger")
    def test_load_env_from_file(self, mock_get_logger):
        mock_get_logger.return_value = logger
        
        """Test loading basic environment variables from .env file."""
        env_content = """
                      TEST_KEY1=value1
                      TEST_KEY2=value2
                      TEST_KEY3=value with spaces
                      TEST_QUOTED="quoted value"
                      """

        with tempfile.NamedTemporaryFile(mode='w', suffix='.env', delete=False) as f:
            f.write(env_content)
            env_file_path = f.name

        try:
            utils.load_env_from_file(env_file_path)

            self.assertEqual(os.environ.get("TEST_KEY1"), "value1")
            self.assertEqual(os.environ.get("TEST_KEY2"), "value2")
            self.assertEqual(os.environ.get("TEST_KEY3"), "value with spaces")
            self.assertEqual(os.environ.get("TEST_QUOTED"), "quoted value")
        finally:
            os.unlink(env_file_path)

    @patch("yr.log.get_logger")
    def test_load_env_from_file_with_comments(self, mock_get_logger):
        mock_get_logger.return_value = logger
        
        """Test loading environment variables with comments."""
        env_content = """
                      # This is a comment
                      TEST_KEY1=value1
                      # Another comment
                      TEST_KEY2=value2
                      """

        with tempfile.NamedTemporaryFile(mode='w', suffix='.env', delete=False) as f:
            f.write(env_content)
            env_file_path = f.name

        try:
            utils.load_env_from_file(env_file_path)

            self.assertEqual(os.environ.get("TEST_KEY1"), "value1")
            self.assertEqual(os.environ.get("TEST_KEY2"), "value2")
        finally:
            os.unlink(env_file_path)

    @patch("yr.log.get_logger")
    def test_load_env_from_file_empty_lines(self, mock_get_logger):
        mock_get_logger.return_value = logger

        """Test loading environment variables with empty lines."""
        env_content = """
                      TEST_KEY1=value1

                      TEST_KEY2=value2
                      """

        with tempfile.NamedTemporaryFile(mode='w', suffix='.env', delete=False) as f:
            f.write(env_content)
            env_file_path = f.name

        try:
            utils.load_env_from_file(env_file_path)

            self.assertEqual(os.environ.get("TEST_KEY1"), "value1")
            self.assertEqual(os.environ.get("TEST_KEY2"), "value2")
        finally:
            os.unlink(env_file_path)


if __name__ == "__main__":
    unittest.main()
