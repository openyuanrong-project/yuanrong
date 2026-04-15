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
from wheel.bdist_wheel import bdist_wheel as _bdist_wheel
from packaging import tags

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
elif os.getenv("SETUP_TYPE") == "all":
    base_name = os.getenv("YR_PACKAGE_NAME", "openyuanrong")
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_ALL,
        f"{base_name}_all",
        "openyuanrong all package",
    )
    setup_spec.entry_points = {
        "console_scripts": [
            "yr=yr.cli.main:main",
            "yrcli=yr.cli.scripts:main",
        ]
    }
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
    ]
    setup_spec.extras["all"] = (
        setup_spec.extras["default"]
        + setup_spec.extras["cpp"]
        + setup_spec.extras["faas"]
        + setup_spec.extras["dashboard"]
    )
    setup_spec.entry_points = {
        "console_scripts": [
            "yrcli=yr.cli.scripts:main",
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
    # Get python runtime version from environment variable
    python_runtime_version = os.getenv("PYTHON_RUNTIME_VERSION", "python3.11")

    # Get python tag (e.g., py311, py310)
    import sys
    ver = sys.version_info
    python_tag = f"py{ver.major}{ver.minor}"

    # Copy third_party from source tree
    third_party_files_to_include = []
    third_party_source_dir = os.path.join(ROOT_DIR, "yr", "third_party")
    if os.path.exists(third_party_source_dir):
        for root, _, fs in os.walk(third_party_source_dir):
            for i in fs:
                third_party_files_to_include.append(os.path.join(root, i))
        for filename in third_party_files_to_include:
            copy_file(
                os.path.join(ctx.build_lib, "yr/third_party"), filename, third_party_source_dir
            )
    # Copy deploy from output directory
    output_root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    deploy_files_to_include = []
    deploy_dir = os.path.join(output_root_dir, "deploy/process")
    for root, _, fs in os.walk(deploy_dir):
        for i in fs:
            deploy_files_to_include.append(os.path.join(root, i))
    for filename in deploy_files_to_include:
        target_dir = os.path.join(ctx.build_lib, "yr/deploy/process")
        copy_file(target_dir, filename, deploy_dir)

    # Update python runtime version in services.yaml files
    import re
    # Replace version in yr/cli/services.yaml (from source)
    cli_services_src = os.path.join(ROOT_DIR, "yr", "cli", "services.yaml")
    cli_services_dst = os.path.join(ctx.build_lib, "yr", "cli", "services.yaml")
    if os.path.exists(cli_services_src):
        os.makedirs(os.path.dirname(cli_services_dst), exist_ok=True)
        with open(cli_services_src, 'r') as f:
            content = f.read()
        # Replace runtime version
        new_content = re.sub(r'runtime: python3\.\d+', f'runtime: {python_runtime_version}', content)
        # Replace function name 'py:' with python tag (e.g., 'py311:')
        new_content = re.sub(r'^(\s*)py:', f'\\1{python_tag}:', new_content, flags=re.MULTILINE)
        with open(cli_services_dst, 'w') as f:
            f.write(new_content)
    # Replace version in yr/deploy/process/services.yaml (copied from output)
    deploy_services_dst = os.path.join(ctx.build_lib, "yr/deploy/process/services.yaml")
    if os.path.exists(deploy_services_dst):
        with open(deploy_services_dst, 'r') as f:
            content = f.read()
        # Replace runtime version
        new_content = re.sub(r'runtime: python3\.\d+', f'runtime: {python_runtime_version}', content)
        # Replace function name 'py:' with python tag (e.g., 'py311:')
        new_content = re.sub(r'^(\s*)py:', f'\\1{python_tag}:', new_content, flags=re.MULTILINE)
        with open(deploy_services_dst, 'w') as f:
            f.write(new_content)

    # Copy Java SDK jars and pom.xml from build output to yr/java/
    java_sdk_dir = os.path.join(output_root_dir, "runtime/sdk/java")
    if os.path.exists(java_sdk_dir):
        java_files_to_include = []
        for f in os.listdir(java_sdk_dir):
            if f.endswith(".jar") or f == "pom.xml":
                java_files_to_include.append(os.path.join(java_sdk_dir, f))
        for filename in java_files_to_include:
            copy_file(os.path.join(ctx.build_lib, "yr/java"), filename, java_sdk_dir)


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
        run_ext(self)


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


def strip_wheel_tests(wheel_path):
    """Remove test directories from wheel."""
    import zipfile
    with zipfile.ZipFile(wheel_path, 'r') as zf:
        names = zf.namelist()
    tests_files = [n for n in names if n.startswith('yr/tests/') or n.startswith('yr/tests\\')]
    if tests_files:
        with zipfile.ZipFile(wheel_path, 'a') as zf:
            for f in tests_files:
                zf.writestr(f, '')  # Overwrite with empty content


class BdistWheelImpl(_bdist_wheel):
    """bdist wheel impl"""

    def get_tag(self):
        """Build wheels with a supported platform tag."""
        tag = next(tags.sys_tags())
        return tag.interpreter, tag.abi, tag.platform

    def run(self):
        super().run()
        for wheel_name in os.listdir(self.dist_dir):
            if not wheel_name.endswith(".whl"):
                continue
            wheel_path = os.path.join(self.dist_dir, wheel_name)
            strip_wheel_tests(wheel_path)


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
    cmdclass={"build_ext": BuildExtImpl, "develop": DevelopImpl, "bdist_wheel": BdistWheelImpl},
    distclass=BinaryDistribution,
    ext_modules=ext_modules,
    packages=setup_spec.get_packages(),
    install_requires=setup_spec.install_requires,
    include_package_data=True,
    package_data={
        "yr": [
            "includes/*.pxd",
            "includes/*.pxi",
            "*.so.*",
            "*.so",
            "cli/*.toml",
            "cli/*.yaml",
            "cli/*.jinja",
            "java/*.jar",
            "java/pom.xml",
        ],
    },
    exclude_package_data={
        "": ["BUILD", "BUILD.bazel"],
    },
    extras_require=setup_spec.extras,
    entry_points=setup_spec.entry_points,
)
