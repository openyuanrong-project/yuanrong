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

"""setup"""

import os
import shutil
import warnings
from enum import Enum

import setuptools
from setuptools.command.build_ext import build_ext

ROOT_DIR = os.path.dirname(__file__)


def get_version():
    """get version"""
    version = os.getenv("BUILD_VERSION", None)
    if version is None or len(version) == 0:
        return open(os.path.join(ROOT_DIR, "../../VERSION")).read().strip()
    return version


class SetupType(Enum):
    """setup type enum"""

    OPENYUANRONG = 1
    OPENYUANRONG_RUNTIME = 2
    OPENYUANRONG_CPP_SDK = 3
    OPENYUANRONG_DASHBOARD = 4
    OPENYUANRONG_FAAS = 5


class SetupSpec:
    """setup spec"""

    def __init__(self, setup_type: SetupType, name: str, description: str):
        self.setup_type = setup_type
        self.name = name
        self.description = description
        self.version = get_version()
        self.install_requires = []
        self.extras = {}
        self.entry_points = {}

    def get_packages(self):
        if self.setup_type == SetupType.OPENYUANRONG:
            return setuptools.find_packages(
                exclude=(
                    "yr.tests",
                    "yr.tests.*",
                    "yr.runtime",
                    "yr.runtime.*",
                    "yr.faas",
                    "yr.faas.*",
                )
            )
        elif self.setup_type == SetupType.OPENYUANRONG_RUNTIME:
            return setuptools.find_packages(
                include=("yr.runtime", "yr.runtime.*", "yr.faas", "yr.faas.*")
            )
        else:
            return []


if os.getenv("SETUP_TYPE") == "runtime":
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_RUNTIME,
        "openyuanrong_runtime",
        "openyuanrong runtime package",
    )
    setup_spec.install_requires = []
elif os.getenv("SETUP_TYPE") == "sdk_cpp":
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_CPP_SDK,
        "openyuanrong_cpp_sdk",
        "openyuanrong cpp sdk",
    )
elif os.getenv("SETUP_TYPE") == "dashboard":
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_DASHBOARD,
        "openyuanrong_dashboard",
        "openyuanrong dashboard",
    )
elif os.getenv("SETUP_TYPE") == "faas":
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_FAAS,
        "openyuanrong_faas",
        "openyuanrong faas",
    )
else:
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG, "openyuanrong", "openyuanrong package"
    )
    setup_spec.install_requires = [
        "cloudpickle==2.2.1",
        "msgpack==1.0.5",
        "protobuf==4.25.5",
        "cython==3.0.10",
        "pyyaml==6.0.2",
        "click==8.1.8",
        "tomli_w==1.2.0",
        "Jinja2==3.1.6",
    ]
    setup_spec.extras["cpp"] = ["openyuanrong_cpp_sdk==" + setup_spec.version]
    setup_spec.extras["dashboard"] = ["openyuanrong_dashboard==" + setup_spec.version]
    setup_spec.extras["faas"] = ["openyuanrong_faas==" + setup_spec.version]
    setup_spec.extras["default"] = [
        "openyuanrong_runtime==" + setup_spec.version,
        "openyuanrong_functionsystem==" + setup_spec.version,
        "openyuanrong_datasystem==" + setup_spec.version,
        "openyuanrong_dashboard==" + setup_spec.version,
    ]
    setup_spec.extras["all"] = (
        setup_spec.extras["default"]
        + setup_spec.extras["cpp"]
        + setup_spec.extras["faas"]
        + setup_spec.extras["dashboard"]
    )
    setup_spec.entry_points = {
        "console_scripts": [
            "yrcli=yr.cli.scripts:run_yr",
            "yr=yr.cli.main:main",
        ]
    }


def copy_file(target, filename, root):
    """copy file"""
    source = os.path.relpath(filename, root)
    dst = os.path.join(target, source)
    os.makedirs(os.path.dirname(dst), exist_ok=True)
    shutil.copy(filename, dst, follow_symlinks=True)


def contains_keyword(text, keywords):
    return any(kw in text for kw in keywords)


def compare_keyword(text, keywords):
    return any(text == kw for kw in keywords)


def copy_openyuanrong_runtime(ctx):
    """copy openyuanrong"""

    runtime_files_to_include = []
    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    runtime_dir = os.path.join(root_dir, "runtime")
    for root, _, fs in os.walk(runtime_dir):
        if contains_keyword(
            root, ["runtime/sdk", "runtime/service/python", "runtime/service/cpp"]
        ):
            continue
        for i in fs:
            runtime_files_to_include.append(os.path.join(root, i))
    for filename in runtime_files_to_include:
        copy_file(os.path.join(ctx.build_lib, "yr/runtime"), filename, runtime_dir)


def copy_openyuanrong_faas(ctx):
    """copy openyuanrong faas"""
    file_to_exclude = [
        "faasfrontend",
        "faasfrontend.zip",
        "faasscheduler",
        "faasscheduler.zip",
    ]
    faas_files_to_include = []
    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    faas_dir = os.path.join(root_dir, "pattern/pattern_faas")
    for root, _, fs in os.walk(faas_dir):
        if contains_keyword(root, ["faasmanager"]):
            continue
        for i in fs:
            if compare_keyword(i, file_to_exclude):
                continue
            faas_files_to_include.append(os.path.join(root, i))
    for filename in faas_files_to_include:
        copy_file(os.path.join(ctx.build_lib, "yr/faas"), filename, faas_dir)


def copy_openyuanrong_dashboard(ctx):
    """copy openyuanrong dashboard"""
    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    dashboard_files_to_include = []
    dashboard_dir = os.path.join(root_dir, "dashboard")
    for root, _, fs in os.walk(dashboard_dir):
        for i in fs:
            dashboard_files_to_include.append(os.path.join(root, i))
    for filename in dashboard_files_to_include:
        copy_file(os.path.join(ctx.build_lib, "yr/dashboard"), filename, dashboard_dir)


def copy_openyuanrong_cpp_sdk(ctx):
    """copy openyuanrong cpp sdk"""
    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    cpp_sdk_files_to_include = []
    cpp_sdk_dir = os.path.join(root_dir, "runtime/sdk/cpp")
    for root, _, fs in os.walk(cpp_sdk_dir):
        for i in fs:
            cpp_sdk_files_to_include.append(os.path.join(root, i))
    for filename in cpp_sdk_files_to_include:
        copy_file(os.path.join(ctx.build_lib, "yr/cpp"), filename, cpp_sdk_dir)
    cpp_runtime_files_to_include = []
    runtime_dir = os.path.join(root_dir, "runtime")
    for root, _, fs in os.walk(runtime_dir):
        if contains_keyword(root, ["runtime/service/cpp"]):
            for i in fs:
                cpp_runtime_files_to_include.append(os.path.join(root, i))
    for filename in cpp_runtime_files_to_include:
        copy_file(os.path.join(ctx.build_lib, "yr/runtime"), filename, runtime_dir)


def copy_openyuanrong(ctx):
    """copy openyuanrong"""
    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    third_party_files_to_include = []
    third_party_dir = os.path.join(root_dir, "third_party")
    for root, _, fs in os.walk(third_party_dir):
        for i in fs:
            third_party_files_to_include.append(os.path.join(root, i))
    for filename in third_party_files_to_include:
        copy_file(
            os.path.join(ctx.build_lib, "yr/third_party"), filename, third_party_dir
        )
    deploy_files_to_include = []
    deploy_dir = os.path.join(root_dir, "deploy/process")
    for root, _, fs in os.walk(deploy_dir):
        for i in fs:
            deploy_files_to_include.append(os.path.join(root, i))
    for filename in deploy_files_to_include:
        copy_file(
            os.path.join(ctx.build_lib, "yr/deploy/process"), filename, deploy_dir
        )


def run_ext(ctx):
    """run ext"""
    if setup_spec.setup_type == SetupType.OPENYUANRONG_RUNTIME:
        copy_openyuanrong_runtime(ctx)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_CPP_SDK:
        copy_openyuanrong_cpp_sdk(ctx)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_DASHBOARD:
        copy_openyuanrong_dashboard(ctx)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_FAAS:
        copy_openyuanrong_faas(ctx)
    else:
        copy_openyuanrong(ctx)


class BuildExtImpl(build_ext):
    """build ext impl"""

    def run(self):
        return run_ext(self)


class BinaryDistribution(setuptools.Distribution):
    """binary distribution"""

    def has_ext_modules(self):
        """has ext modules"""
        return True


warnings.filterwarnings("ignore", category=setuptools.SetuptoolsDeprecationWarning)
setuptools.setup(
    name=setup_spec.name,
    version=setup_spec.version,
    author="openyuanrong",
    classifiers=[
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
    cmdclass={"build_ext": BuildExtImpl},
    distclass=BinaryDistribution,
    packages=setup_spec.get_packages(),
    install_requires=setup_spec.install_requires,
    include_package_data=True,
    package_data={
        "yr": [
            "includes/*.pxd",
            "includes/*.pxi",
            "*.so.*",
            "*.so",
            "cli/values.toml",
            "cli/config.toml.jinja",
            "cli/services.yaml",
            "config/python-runtime-log.json",
        ],
    },
    exclude_package_data={
        "": ["BUILD", "BUILD.bazel"],
    },
    extras_require=setup_spec.extras,
    entry_points=setup_spec.entry_points,
)
