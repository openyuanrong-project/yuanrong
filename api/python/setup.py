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

from enum import Enum
import os
import shutil
import warnings

import setuptools
from setuptools.command.build_ext import build_ext
from setuptools.command.develop import develop
from setuptools import Extension

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
    OPENYUANRONG_SDK = 2
    OPENYUANRONG_CPP_SDK = 3
    OPENYUANRONG_ALL = 4
    OPENYUANRONG_RUNTIME = 5
    OPENYUANRONG_DASHBOARD = 6
    OPENYUANRONG_FAAS = 7


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
            return setuptools.find_packages(include=("yr.inner", "yr.inner.*"))
        elif self.setup_type == SetupType.OPENYUANRONG_RUNTIME:
            return setuptools.find_packages(
                include=("yr.runtime", "yr.runtime.*", "yr.faas", "yr.faas.*")
            )
        elif self.setup_type == SetupType.OPENYUANRONG_SDK:
            return setuptools.find_packages(
                exclude=("yr.tests", "yr.tests.*", "yr.inner", "yr.inner.*")
            )
        else:
            return []


if os.getenv("SETUP_TYPE") == "sdk":
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_SDK,
        f"{base_name}_sdk",
        "openyuanrong python sdk",
    )
    setup_spec.install_requires = [
        "cloudpickle==2.2.1",
        "msgpack==1.0.5",
        "protobuf==4.25.5",
        "cython==3.0.10",
        "pyyaml==6.0.2",
        "click==8.1.8",
        "requests==2.32.5",
        "websockets==15.0.1",
    ]
    setup_spec.entry_points = {
        "console_scripts": [
            "yrcli=yr.cli.scripts:main",
        ]
    }
elif os.getenv("SETUP_TYPE") == "dashboard":
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_DASHBOARD,
        f"{base_name}_dashboard",
        "openyuanrong dashboard",
    )
elif os.getenv("SETUP_TYPE") == "faas":
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_FAAS,
        f"{base_name}_faas",
        "openyuanrong faas",
    )
elif os.getenv("SETUP_TYPE") == "runtime":
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_RUNTIME,
        f"{base_name}_runtime",
        "openyuanrong runtime",
    )
elif os.getenv("SETUP_TYPE") == "sdk_cpp":
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_CPP_SDK,
        f"{base_name}_cpp_sdk",
        "openyuanrong cpp sdk",
    )
elif os.getenv("SETUP_TYPE") == "all":
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_ALL,
        f"{base_name}_all",
        "openyuanrong all package",
    )
    setup_spec.entry_points = {
        "console_scripts": [
            "yr=yr.inner.scripts:run_yr",
            "yrcli=yr.cli.scripts:main",
        ]
    }
else:
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG, base_name, "openyuanrong package"
    )
    setup_spec.install_requires = [
        f"{base_name}_sdk==" + setup_spec.version,
    ]
    setup_spec.extras["cpp"] = [f"{base_name}_cpp_sdk==" + setup_spec.version]
    setup_spec.entry_points = {
        "console_scripts": [
            "yr=yr.inner.scripts:run_yr",
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


def copy_openyuanrong(build_lib):
    """copy openyuanrong"""
    keyword_to_exclude = [
        "datasystem/sdk",
        "deploy/k8s",
        "functionsystem/bin/domain_scheduler",
        "functionsystem/bin/iam_server",
        "functionsystem/bin/runtime_manager",
        "functionsystem/sym",
        "pattern_faas/faasmanager",
        "runtime/sdk",
    ]
    file_to_exclude = [
        "faasfrontend",
        "faasfrontend.zip",
        "faasscheduler",
        "faasscheduler.zip",
    ]
    files_to_include = []
    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    for root, _, fs in os.walk(root_dir):
        if contains_keyword(root, keyword_to_exclude):
            continue
        for i in fs:
            if compare_keyword(i, file_to_exclude):
                continue
            files_to_include.append(os.path.join(root, i))
    for filename in files_to_include:
        copy_file(os.path.join(build_lib, "yr/inner"), filename, root_dir)


def copy_openyuanrong_cpp_sdk(build_lib):
    """copy openyuanrong"""
    import re

    python_runtime_version = os.getenv("PYTHON_RUNTIME_VERSION", "python3.9")
    files_to_include = []
    for root, _, fs in os.walk("./yr"):
        for i in fs:
            if "so" in i:
                files_to_include.append(os.path.join(root, i))
    for filename in files_to_include:
        copy_file(build_lib, filename, ROOT_DIR)

    # Update python runtime version in services.yaml files
    # Replace version in yr/cli/services.yaml (from source)
    cli_services_src = os.path.join(ROOT_DIR, "yr", "cli", "services.yaml")
    cli_services_dst = os.path.join(build_lib, "yr", "cli", "services.yaml")
    if os.path.exists(cli_services_src):
        os.makedirs(os.path.dirname(cli_services_dst), exist_ok=True)
        with open(cli_services_src, 'r') as f:
            content = f.read()
        new_content = re.sub(r'runtime: python3\.\d+', f'runtime: {python_runtime_version}', content)
        with open(cli_services_dst, 'w') as f:
            f.write(new_content)
    # Replace version in yr/deploy/process/services.yaml (copied from output)
    deploy_services_dst = os.path.join(build_lib, "yr/deploy/process/services.yaml")
    if os.path.exists(deploy_services_dst):
        with open(deploy_services_dst, 'r') as f:
            content = f.read()
        new_content = re.sub(r'runtime: python3\.\d+', f'runtime: {python_runtime_version}', content)
        with open(deploy_services_dst, 'w') as f:
            f.write(new_content)


def copy_openyuanrong_runtime(build_lib):
    """copy openyuanrong runtime"""
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
        copy_file(os.path.join(build_lib, "yr/runtime"), filename, runtime_dir)


def copy_openyuanrong_faas(build_lib):
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
        copy_file(os.path.join(build_lib, "yr/faas"), filename, faas_dir)


def copy_openyuanrong_dashboard(build_lib):
    """copy openyuanrong dashboard"""
    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    dashboard_files_to_include = []
    dashboard_dir = os.path.join(root_dir, "dashboard")
    for root, _, fs in os.walk(dashboard_dir):
        for i in fs:
            dashboard_files_to_include.append(os.path.join(root, i))
    for filename in dashboard_files_to_include:
        copy_file(os.path.join(build_lib, "yr/dashboard"), filename, dashboard_dir)


def run_ext(build_lib):
    """run ext"""
    if setup_spec.setup_type == SetupType.OPENYUANRONG:
        copy_openyuanrong(build_lib)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_CPP_SDK:
        copy_openyuanrong_cpp_sdk(build_lib)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_ALL:
        copy_openyuanrong(build_lib)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_RUNTIME:
        copy_openyuanrong_runtime(build_lib)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_DASHBOARD:
        copy_openyuanrong_dashboard(build_lib)
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_FAAS:
        copy_openyuanrong_faas(build_lib)


class BuildExtImpl(build_ext):
    """build ext impl"""

    def run(self):
        run_ext(self.build_lib)


class DevelopImpl(develop):
    """develop impl for editable install"""

    def run(self):
        super().run()
        run_ext(ROOT_DIR)


class BinaryDistribution(setuptools.Distribution):
    """binary distribution"""

    def has_ext_modules(self):
        """has ext modules"""
        return True


warnings.filterwarnings("ignore", category=setuptools.SetuptoolsDeprecationWarning)

ext_modules = []
if setup_spec.setup_type in [SetupType.OPENYUANRONG, SetupType.OPENYUANRONG_CPP_SDK, SetupType.OPENYUANRONG_ALL, SetupType.OPENYUANRONG_RUNTIME, SetupType.OPENYUANRONG_DASHBOARD, SetupType.OPENYUANRONG_FAAS]:
    ext_modules = [Extension("yr._dummy", sources=[])]

setuptools.setup(
    name=setup_spec.name,
    version=setup_spec.version,
    author="openyuanrong",
    classifiers=[
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
    cmdclass={"build_ext": BuildExtImpl, "develop": DevelopImpl},
    distclass=BinaryDistribution,
    ext_modules=ext_modules,
    packages=setup_spec.get_packages(),
    install_requires=setup_spec.install_requires,
    include_package_data=True,
    package_data={
        "yr": ["includes/*.pxd", "includes/*.pxi", "*.so.*", "*.so"],
    },
    exclude_package_data={
        "": ["BUILD", "BUILD.bazel"],
    },
    extras_require=setup_spec.extras,
    entry_points=setup_spec.entry_points,
)
