#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
import json
import uuid
from pathlib import Path
from typing import Dict

from yr.config import InvokeOptions

WORKING_DIR_KEY = "WORKING_DIR"
CONDA_PREFIX = "CONDA_PREFIX"


def _get_conda_bin_executable(executable_name: str) -> str:
    conda_home = os.environ.get("YR_CONDA_HOME")
    if conda_home:
        return conda_home
    if CONDA_PREFIX in os.environ:
        return os.environ.get(CONDA_PREFIX)
    raise ValueError(
        "please configure YR_CONDA_HOME environment variable which contain a bin subdirectory"
    )


def _check_pip_and_conda(runtime_env: Dict):
    if runtime_env.get("conda") and runtime_env.get("pip"):
        raise ValueError(
            "The 'pip' field and 'conda' field of "
            "runtime_env cannot both be specified.\n"
            f"specified pip field: {runtime_env['pip']}\n"
            f"specified conda field: {runtime_env['conda']}\n"
            "To use pip with conda, please only set the 'conda' "
            "field, and specify your pip dependencies "
            "within the conda YAML config dict"
        )


def _process_pip(opt: InvokeOptions, create_opt: Dict):
    _check_pip_and_conda(opt.runtime_env)
    pip_command = "pip3 install " + " ".join(opt.runtime_env.get("pip"))
    create_opt["POST_START_EXEC"] = pip_command


def _process_conda(opt: InvokeOptions, create_opt: Dict):
    _check_pip_and_conda(opt.runtime_env)
    create_opt[CONDA_PREFIX] = _get_conda_bin_executable("conda")
    conda_config = opt.runtime_env.get("conda")
    if isinstance(conda_config, str):
        yaml_file = Path(conda_config)
        if yaml_file.suffix in (".yaml", ".yml"):
            if not yaml_file.is_file():
                raise ValueError(f"Can't find conda YAML file {yaml_file}.")
            try:
                import yaml
                result = yaml.safe_load(yaml_file.read_text())
                name = result.get("name", str(uuid.uuid4()))
                json_str = json.dumps(result)
                create_opt["CONDA_CONFIG"] = json_str
                conda_command = "conda env create -f env.yaml"
                create_opt["CONDA_COMMAND"] = conda_command
                create_opt["CONDA_DEFAULT_ENV"] = name
            except Exception as e:
                raise ValueError(f"Failed to read conda file {yaml_file}") from e
        else:
            conda_command = f"conda activate {conda_config}"
            create_opt["CONDA_COMMAND"] = conda_command
            create_opt["CONDA_DEFAULT_ENV"] = conda_config
    if isinstance(conda_config, dict):
        try:
            json_str = json.dumps(conda_config)
            name = conda_config.get("name", str(uuid.uuid4()))
            create_opt["CONDA_CONFIG"] = json_str
            conda_command = "conda env create -f env.yaml"
            create_opt["CONDA_COMMAND"] = conda_command
            create_opt["CONDA_DEFAULT_ENV"] = name
        except Exception as e:
            raise ValueError(f"Failed to load conda {conda_config}") from e
    if not isinstance(conda_config, dict) and not isinstance(conda_config, str):
        raise TypeError("runtime_env.get('conda') must be of type dict or str")


def _process_working_dir(opt: InvokeOptions, create_opt: Dict):
    working_dir = opt.runtime_env.get("working_dir")
    if not isinstance(working_dir, str):
        raise TypeError("`working_dir` must be a string, got " f"{type(working_dir)}.")
    create_opt[WORKING_DIR_KEY] = working_dir


def _process_env_vars(opt: InvokeOptions, create_opt: Dict):
    env_vars = opt.runtime_env.get("env_vars")
    if not isinstance(env_vars, dict):
        raise TypeError(
            "runtime_env.get('env_vars') must be of type "
            f"Dict[str, str], got {type(env_vars)}"
        )
    for key, val in env_vars.items():
        if not isinstance(key, str):
            raise TypeError(
                "runtime_env.get('env_vars') must be of type "
                f"Dict[str, str], but the key {key} is of type {type(key)}"
            )
        if not isinstance(val, str):
            raise TypeError(
                "runtime_env.get('env_vars') must be of type "
                f"Dict[str, str], but the value {val} is of type {type(val)}"
            )
    # Instead of modifying opt.env_vars, we will place runtime_env["env_vars"] into create_opt,
    # where the priority of opt.env_vars will be handled in parse_invoke_opts
    create_opt["env_vars"] = env_vars


def _process_shared_dir(opt: InvokeOptions, create_opt: Dict):
    shared_dir = opt.runtime_env.get("shared_dir")
    if not isinstance(shared_dir, dict):
        raise TypeError(
            "runtime_env.get('shared_dir') must be of type "
            f"Dict[str, str], got {type(shared_dir)}")
    if "name" not in shared_dir:
        raise ValueError("runtime_env.get('shared_dir') contain of 'name'")
    name = shared_dir["name"]
    if not isinstance(name, str):
        raise TypeError(
            "runtime_env['shared_dir']['name'] must be of type str"
            f"but the value {name} is of type {type(name)}")
    if "TTL" in shared_dir:
        ttl = shared_dir["TTL"]
    else:
        ttl = 0
    if not isinstance(ttl, int):
        raise TypeError(
            "runtime_env['shared_dir']['TTL'] must be of type int"
            f"but the value {ttl} is of type {type(ttl)}")
    create_opt["DELEGATE_SHARED_DIRECTORY"] = name
    create_opt["DELEGATE_SHARED_DIRECTORY_TTL"] = f"{ttl}"


_runtime_env_processors = {
    "pip": _process_pip,
    "conda": _process_conda,
    "working_dir": _process_working_dir,
    "env_vars": _process_env_vars,
    "shared_dir": _process_shared_dir
}


def parse_runtime_env(opt: InvokeOptions) -> Dict:
    """
    parse runtime env to create options
    """
    create_opt = {}
    if opt.runtime_env is None:
        return create_opt
    if not isinstance(opt.runtime_env, dict):
        raise TypeError("`InvokeOptions.runtime_env` must be a dict, got " f"{type(opt.runtime_env)}.")

    for key in opt.runtime_env.keys():
        if key not in _runtime_env_processors:
            raise ValueError(f"runtime_env.get('{key}') is not supported.")
        _runtime_env_processors[key](opt, create_opt)

    return create_opt
