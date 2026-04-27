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
import hashlib
import os
import re as _re
import shutil
import subprocess
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
    OPENYUANRONG_FULL = 8


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
        if self.setup_type in (SetupType.OPENYUANRONG, SetupType.OPENYUANRONG_FULL):
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
elif os.getenv("SETUP_TYPE") == "full":
    setup_spec = SetupSpec(
        SetupType.OPENYUANRONG_FULL,
        "openyuanrong-full",
        "openyuanrong full package (all-in-one)",
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
    setup_spec.entry_points = {
        "console_scripts": [
            "yr=yr.inner.scripts:run_yr",
            "yrcli=yr.cli.scripts:main",
        ]
    }
else:
    setup_spec = SetupSpec(SetupType.OPENYUANRONG, "openyuanrong", "openyuanrong package")
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


def _get_soname(filepath):
    """Read SONAME of a shared library via readelf. Returns None if not found."""
    try:
        result = subprocess.run(
            ["readelf", "-d", filepath], capture_output=True, text=True, timeout=10
        )
        for line in result.stdout.splitlines():
            if "(SONAME)" in line:
                m = _re.search(r"\[([^\]]+)\]", line)
                if m:
                    return m.group(1)
    except Exception:
        pass
    return None


def _md5_file(filepath):
    """Compute MD5 digest of a file."""
    h = hashlib.md5()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


def _so_specificity(filepath):
    """Return sort key for .so version specificity (more version parts = more specific)."""
    name = os.path.basename(filepath)
    if ".so." in name:
        ver_str = name.split(".so.", 1)[1]
        parts = ver_str.split(".")
        return (len(parts), [int(p) if p.isdigit() else 0 for p in parts])
    return (0, [])


def _collect_so_files(directory):
    """Collect all .so* regular files directly in directory (non-recursive)."""
    result = []
    if not os.path.isdir(directory):
        return result
    for name in os.listdir(directory):
        if ".so" in name:
            path = os.path.join(directory, name)
            if os.path.isfile(path):
                result.append(path)
    return result


def _dedup_so_in_dir(directory):
    """Remove redundant .so files within a directory.

    Files sharing the same SONAME and MD5 are identical copies.
    Only the most version-specific filename is kept (e.g. libfoo.so.1.2.3
    is kept over libfoo.so.1 and libfoo.so).
    Returns the number of files removed.
    """
    groups = {}
    for f in _collect_so_files(directory):
        soname = _get_soname(f)
        if soname is None:
            continue
        md5 = _md5_file(f)
        groups.setdefault((soname, md5), []).append(f)

    removed = 0
    for (soname, _md5), files in groups.items():
        if len(files) <= 1:
            continue
        # Prefer the file whose name equals the SONAME: that is what the dynamic
        # linker resolves at runtime (NEEDED entries store the SONAME string and
        # the loader looks for a file with that exact name).
        to_keep = next((f for f in files if os.path.basename(f) == soname), None)
        if to_keep is None:
            files.sort(key=_so_specificity)
            to_keep = files[-1]
        for f in files:
            if f != to_keep:
                print(f"  [dedup-indir] {os.path.basename(f)} -> kept {os.path.basename(to_keep)}")
                os.remove(f)
                removed += 1
    return removed


def _get_needed(filepath):
    """Return the set of NEEDED sonames declared in a binary/library."""
    needed = set()
    try:
        result = subprocess.run(
            ["readelf", "-d", filepath], capture_output=True, text=True, timeout=10
        )
        for line in result.stdout.splitlines():
            if "(NEEDED)" in line:
                m = _re.search(r"\[([^\]]+)\]", line)
                if m:
                    needed.add(m.group(1))
    except Exception:
        pass
    return needed


def _dedup_so_cross_dirs(primary_dir, secondary_dir):
    """Remove .so files from secondary_dir that are identical to ones in primary_dir.

    Identity is determined by (SONAME, MD5).  Files unique to secondary_dir
    are always preserved.

    Safety: files that are NEEDED (directly or transitively) by the unique files
    remaining in secondary_dir are also preserved, because secondary_dir is a
    self-contained library directory (RUNPATH=$ORIGIN) and removing a transitive
    dependency would break the runtime.

    Returns the number of files removed.
    """
    primary_index = set()
    for f in _collect_so_files(primary_dir):
        soname = _get_soname(f)
        if soname is None:
            continue
        primary_index.add((soname, _md5_file(f)))

    # Classify all files in secondary_dir
    sec_meta = {}  # filepath -> (soname, md5)
    for f in _collect_so_files(secondary_dir):
        soname = _get_soname(f)
        if soname is None:
            continue
        sec_meta[f] = (soname, _md5_file(f))

    candidates = {f for f, (soname, md5) in sec_meta.items() if (soname, md5) in primary_index}

    # Seed required_sonames from ALL non-candidate files in secondary_dir.
    # This must include .so files without a SONAME (e.g. libcpplibruntime.so)
    # which are go-unique but are invisible to sec_meta, as well as plain
    # executables (e.g. goruntime).  Any file that isn't a removal candidate
    # stays in secondary_dir, so its dynamic dependencies must be preserved.
    all_sec_files = [
        os.path.join(secondary_dir, n)
        for n in os.listdir(secondary_dir)
        if os.path.isfile(os.path.join(secondary_dir, n))
    ]
    required_sonames = set()
    for f in all_sec_files:
        if f not in candidates:
            required_sonames.update(_get_needed(f))

    changed = True
    while changed:
        changed = False
        for f in list(candidates):
            soname, _ = sec_meta[f]
            if soname in required_sonames:
                candidates.discard(f)
                required_sonames.update(_get_needed(f))
                changed = True

    removed = 0
    for f in candidates:
        print(
            f"  [dedup-cross] removing {os.path.basename(f)} from {os.path.relpath(secondary_dir)}"
        )
        os.remove(f)
        removed += 1
    return removed


def _remove_build_artifacts(root_dir):
    """Remove .a static libraries and cmake/pkgconfig build dirs under root_dir.

    Returns the number of items removed.
    """
    removed = 0
    for dirpath, dirnames, filenames in os.walk(root_dir, topdown=True):
        for d in list(dirnames):
            if d in ("cmake", "pkgconfig"):
                full = os.path.join(dirpath, d)
                shutil.rmtree(full)
                dirnames.remove(d)
                print(f"  [artifacts] removed dir {os.path.relpath(full)}")
                removed += 1
        for fname in filenames:
            if fname.endswith(".a"):
                full = os.path.join(dirpath, fname)
                os.remove(full)
                print(f"  [artifacts] removed {os.path.relpath(full)}")
                removed += 1
    return removed


def optimize_wheel_files(build_lib, setup_type):
    """Shrink wheel by deduplicating .so files and removing build-only artifacts.

    For OPENYUANRONG_RUNTIME:
      - Dedup within java/lib and go/bin separately.
      - Cross-dedup: remove go/bin files that are identical to java/lib files.
      - Remove .a, cmake/, pkgconfig/ under yr/runtime/.

    For OPENYUANRONG_CPP_SDK:
      - Dedup within yr/cpp/lib and yr/runtime/service/cpp/lib separately.
      - Remove .a, cmake/, pkgconfig/ under yr/cpp/ and yr/runtime/service/cpp/.
    """
    if setup_type == SetupType.OPENYUANRONG_RUNTIME:
        java_lib = os.path.join(build_lib, "yr/runtime/service/java/lib")
        go_bin = os.path.join(build_lib, "yr/runtime/service/go/bin")
        print("Optimizing openyuanrong_runtime wheel files...")
        n = _dedup_so_in_dir(java_lib)
        print(f"  java/lib intra-dedup: removed {n} files")
        n = _dedup_so_in_dir(go_bin)
        print(f"  go/bin intra-dedup: removed {n} files")
        n = _dedup_so_cross_dirs(java_lib, go_bin)
        print(f"  go/bin vs java/lib cross-dedup: removed {n} files")
        n = _remove_build_artifacts(os.path.join(build_lib, "yr/runtime"))
        print(f"  build artifacts removed: {n} items")

    elif setup_type == SetupType.OPENYUANRONG_CPP_SDK:
        cpp_lib = os.path.join(build_lib, "yr/cpp/lib")
        cpp_svc_lib = os.path.join(build_lib, "yr/runtime/service/cpp/lib")
        print("Optimizing openyuanrong_cpp_sdk wheel files...")
        n = _dedup_so_in_dir(cpp_lib)
        print(f"  cpp/lib intra-dedup: removed {n} files")
        n = _dedup_so_in_dir(cpp_svc_lib)
        print(f"  service/cpp/lib intra-dedup: removed {n} files")
        n = _remove_build_artifacts(os.path.join(build_lib, "yr/cpp"))
        n += _remove_build_artifacts(os.path.join(build_lib, "yr/runtime/service/cpp"))
        print(f"  build artifacts removed: {n} items")

    elif setup_type == SetupType.OPENYUANRONG_FULL:
        print("Optimizing openyuanrong-full wheel files...")
        inner = os.path.join(build_lib, "yr/inner")
        total_removed = 0

        # Only do intra-directory dedup (same SONAME + MD5 within one directory).
        # Cross-directory dedup is NOT safe because executables in parent/sibling
        # directories depend on .so files in lib/ subdirs without RUNPATH=$ORIGIN,
        # and the loader resolves them via LD_LIBRARY_PATH set by deploy scripts.
        for dirpath, _, _ in os.walk(inner):
            dirname = os.path.basename(dirpath)
            if dirname in ("lib", "bin"):
                n = _dedup_so_in_dir(dirpath)
                if n:
                    print(f"  intra-dedup {os.path.relpath(dirpath, inner)}: removed {n} files")
                total_removed += n

        n = _remove_build_artifacts(os.path.join(inner, "runtime"))
        n += _remove_build_artifacts(os.path.join(inner, "datasystem"))
        total_removed += n
        print(f"  build artifacts removed: {n} items")
        print(f"  total files removed: {total_removed}")


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
        if contains_keyword(root, ["runtime/sdk", "runtime/service/python", "runtime/service/cpp"]):
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
        with open(cli_services_src, "r") as f:
            content = f.read()
        # Replace runtime version
        new_content = re.sub(
            r"runtime: python3\.\d+", f"runtime: {python_runtime_version}", content
        )
        # Replace function name 'py:' with python tag (e.g., 'py311:')
        new_content = re.sub(r"^(\s*)py:", f"\\1{python_tag}:", new_content, flags=re.MULTILINE)
        with open(cli_services_dst, "w") as f:
            f.write(new_content)
    # Replace version in yr/deploy/process/services.yaml (copied from output)
    deploy_services_dst = os.path.join(ctx.build_lib, "yr/deploy/process/services.yaml")
    if os.path.exists(deploy_services_dst):
        with open(deploy_services_dst, "r") as f:
            content = f.read()
        # Replace runtime version
        new_content = re.sub(
            r"runtime: python3\.\d+", f"runtime: {python_runtime_version}", content
        )
        # Replace function name 'py:' with python tag (e.g., 'py311:')
        new_content = re.sub(r"^(\s*)py:", f"\\1{python_tag}:", new_content, flags=re.MULTILINE)
        with open(deploy_services_dst, "w") as f:
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


def copy_openyuanrong_full(ctx):
    """Copy all components for the openyuanrong-full monolithic package.

    Bundles runtime, functionsystem, datasystem, faas, deploy, third_party
    etc. into yr/inner/ so the wheel is self-contained.

    Unlike the base openyuanrong package, this does NOT call
    copy_openyuanrong() because yr/third_party and yr/deploy would duplicate
    content already placed in yr/inner/ by this function.  The dashboard
    directory is also skipped since dashboard binaries already live in
    functionsystem/bin/.
    """
    # Copy all output directories into yr/inner/ EXCEPT dashboard/
    # (dashboard binaries are already included in functionsystem/bin/)
    _FULL_EXCLUDE_COMPONENTS = {"dashboard", "VERSION"}

    # Also skip datasystem/sdk — the Go/C++ SDK is huge (~800 MB) and not needed
    # for deployment; only datasystem/service is required at runtime.
    _FULL_EXCLUDE_SUBDIRS = {
        "datasystem/sdk",
        "runtime/service/python/yr/third_party",
    }
    _FULL_EXCLUDE_FILES = {
        "runtime/service/python/yr/libdatasystem_worker.so",
        "runtime/service/java/lib/libdatasystem_worker.so",
        "runtime/service/cpp/lib/libdatasystem_worker.so",
        "pattern/pattern_faas/faasfrontend/faasfrontend.zip",
        "pattern/pattern_faas/faasfrontend/faasfrontend",
        "pattern/pattern_faas/faasscheduler/faasscheduler",
        "pattern/pattern_faas/faasscheduler/faasscheduler.zip",
        "pattern/pattern_faas/faasmanager/faasmanager.zip",
    }

    root_dir = os.path.join(ROOT_DIR, "../../output/openyuanrong")
    inner_target = os.path.join(ctx.build_lib, "yr/inner")
    for component in os.listdir(root_dir):
        if component in _FULL_EXCLUDE_COMPONENTS:
            continue
        component_dir = os.path.join(root_dir, component)
        if not os.path.isdir(component_dir):
            continue
        files_to_include = []
        for root, _, fs in os.walk(component_dir):
            # Skip excluded subdirectories (e.g. datasystem/sdk)
            rel = os.path.relpath(root, root_dir)
            if any(rel == excl or rel.startswith(excl + "/") for excl in _FULL_EXCLUDE_SUBDIRS):
                continue
            for f in fs:
                file_rel = os.path.relpath(os.path.join(root, f), root_dir)
                if file_rel in _FULL_EXCLUDE_FILES:
                    continue
                files_to_include.append(os.path.join(root, f))
        for filename in files_to_include:
            copy_file(inner_target, filename, root_dir)


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
    elif setup_spec.setup_type == SetupType.OPENYUANRONG_FULL:
        copy_openyuanrong_full(ctx)
    else:
        copy_openyuanrong(ctx)


class BuildExtImpl(build_ext):
    """build ext impl"""

    def run(self):
        run_ext(self)
        # For the full package, remove files from the base yr/ layer that are
        # redundant (already provided by yr/inner/ at runtime).
        if setup_spec.setup_type == SetupType.OPENYUANRONG_FULL:
            _full_exclude_base_files = [
                "yr/libdatasystem_worker.so",
            ]
            for f in _full_exclude_base_files:
                path = os.path.join(self.build_lib, f)
                if os.path.exists(path):
                    print(f"  [full] removing redundant base file: {f}")
                    os.remove(path)
        optimize_wheel_files(self.build_lib, setup_spec.setup_type)


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

    with zipfile.ZipFile(wheel_path, "r") as zf:
        names = zf.namelist()
    tests_files = [n for n in names if n.startswith("yr/tests/") or n.startswith("yr/tests\\")]
    if tests_files:
        with zipfile.ZipFile(wheel_path, "a") as zf:
            for f in tests_files:
                zf.writestr(f, "")  # Overwrite with empty content


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
if setup_spec.setup_type in [
    SetupType.OPENYUANRONG,
    SetupType.OPENYUANRONG_CPP_SDK,
    SetupType.OPENYUANRONG_ALL,
    SetupType.OPENYUANRONG_RUNTIME,
    SetupType.OPENYUANRONG_DASHBOARD,
    SetupType.OPENYUANRONG_FAAS,
]:
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
