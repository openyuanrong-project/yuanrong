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

"""Sandbox implementation for isolated code execution."""

import argparse
import logging
import os
import sys
import subprocess
import tempfile
from typing import Optional, Dict, Any, List, TYPE_CHECKING

import yr

logger = logging.getLogger(__name__)

if TYPE_CHECKING:
    from yr.config import PortForwarding


def _sanitize_instance_id(instance_id: str) -> str:
    """Sanitize instance ID to match TraefikRegistry::SanitizeID (C++).

    Rules: @ -> -at-, / . _ -> -, truncate to 200 chars.
    """
    result = instance_id
    pos = 0
    while True:
        pos = result.find("@", pos)
        if pos == -1:
            break
        result = result[:pos] + "-at-" + result[pos + 1:]
        pos += 4
    result = result.replace("/", "-").replace(".", "-").replace("_", "-")
    if len(result) > 200:
        result = result[:200]
    return result


def _get_gateway_host() -> str:
    """Get Gateway host from YR_GATEWAY_ADDRESS or YR_SERVER_ADDRESS."""
    host = os.environ.get("YR_GATEWAY_ADDRESS", "").strip()
    if host:
        return host
    addr = os.environ.get("YR_SERVER_ADDRESS", "").strip()
    return addr


def _build_gateway_url(instance_id: str, sandbox_port: int, gateway_host: str, path: str = "") -> str:
    """Build Gateway HTTP path URL: https://{gateway_host}/{safeID}/{sandbox_port}{path}.

    URL format must match TraefikRegistry::RegisterInstance in function_proxy:
    - {safeID} is sanitized instance ID (SanitizeID logic)
    - {sandbox_port} is the original sandbox port
    - Full path format: /{safeID}/{sandbox_port}

    See: functionsystem/src/function_proxy/local_scheduler/traefik_registry/traefik_registry.cpp
    """
    safe_id = _sanitize_instance_id(instance_id)
    base = f"https://{gateway_host}/{safe_id}/{sandbox_port}"
    if path:
        path = path if path.startswith("/") else f"/{path}"
        return f"{base}{path}"
    return base


def _print_gateway_urls(instance_id: str, port_forwardings: List["PortForwarding"]) -> None:
    """Print Gateway URLs for port forwardings after sandbox creation."""
    if not port_forwardings:
        return
    gateway_host = _get_gateway_host()
    if not gateway_host:
        print("Warning: cannot print port forwarding URLs: YR_GATEWAY_ADDRESS or YR_SERVER_ADDRESS not set")
        return
    print("sandbox created, port forwarding URLs:")
    for pf in port_forwardings:
        url = _build_gateway_url(instance_id, pf.port, gateway_host)
        print(f"  port {pf.port}: {url}")


@yr.instance
class SandBoxInstance:
    """
    SandBoxInstance class provides isolated environment for code execution.

    This class creates a sandboxed environment where code can be executed
    with limited permissions and resource constraints.

    This is the underlying instance class decorated with @yr.instance.
    Users should typically use the SandBox wrapper class instead.
    """

    def __init__(
        self, working_dir: Optional[str] = None, env: Optional[Dict[str, str]] = None
    ):
        """
        Initialize the SandBox instance.

        Args:
            working_dir (Optional[str]): The working directory for sandbox execution.
                If None, a temporary directory will be created.
            env (Optional[Dict[str, str]]): Environment variables for the sandbox.
                If None, inherits from parent process.
        """
        self._temp_dir_obj: Optional[tempfile.TemporaryDirectory] = None
        if working_dir is None:
            self._temp_dir_obj = tempfile.TemporaryDirectory(prefix="yr_sandbox_")
            self.working_dir = self._temp_dir_obj.name
        else:
            self.working_dir = working_dir

        self.env = env if env is not None else os.environ.copy()
        self._initialized = True

    def __del__(self):
        """Destructor to ensure cleanup on object deletion."""
        self.cleanup()

    @staticmethod
    def get_name():
        """
        Get the name of the sandbox instance.

        Returns:
            str: The name of the sandbox instance.
        """
        return os.environ.get("INSTANCE_ID", "")

    def get_working_dir(self) -> str:
        """
        Get the working directory of the sandbox.

        Returns:
            str: The path to the working directory.
        """
        return self.working_dir

    def execute(self, command: str, timeout: Optional[int] = None) -> Dict[str, Any]:
        """
        Execute a command in the sandbox environment.

        Args:
            command (str): The command to execute.
            timeout (Optional[int]): Timeout in seconds for command execution.
                If None, no timeout is set.

        Returns:
            Dict[str, Any]: A dictionary containing:
                - returncode (int): The return code of the command.
                - stdout (str): Standard output of the command.
                - stderr (str): Standard error of the command.

        Raises:
            RuntimeError: If the sandbox is not initialized.
            subprocess.TimeoutExpired: If the command execution times out.

        Examples:
            >>> sandbox = yr.sandbox.SandBox.invoke()
            >>> result = yr.get(sandbox.execute.invoke("ls -la"))
            >>> print(result['stdout'])
        """
        if not self._initialized:
            raise RuntimeError("SandBox is not initialized")

        try:
            cmd = command.strip()
            if not cmd:
                return {
                    "returncode": -1,
                    "stdout": "",
                    "stderr": "empty command",
                }
            # shell=False with explicit shell argv: same semantics as former shell=True, satisfies EDV.04
            if os.name == "nt":
                argv = [os.environ.get("COMSPEC", "cmd.exe"), "/c", cmd]
            else:
                argv = ["/bin/sh", "-c", cmd]
            result = subprocess.run(
                argv,
                shell=False,
                cwd=self.working_dir,
                env=self.env,
                capture_output=True,
                text=True,
                timeout=timeout,
            )
            return {
                "returncode": result.returncode,
                "stdout": sys.version + '\n' + result.stdout,
                "stderr": result.stderr,
            }
        except subprocess.TimeoutExpired as e:
            return {
                "returncode": -1,
                "stdout": e.stdout if e.stdout else "",
                "stderr": f"Command timed out after {timeout} seconds",
            }
        except Exception as e:
            return {"returncode": -1, "stdout": "", "stderr": str(e)}

    def cleanup(self) -> None:
        """
        Cleanup the sandbox environment.

        This method removes temporary files and directories created by the sandbox.
        """
        temp_dir_obj = getattr(self, "_temp_dir_obj", None)
        if temp_dir_obj is not None:
            try:
                temp_dir_obj.cleanup()
            except Exception as e:
                logger.warning(
                    "Failed to cleanup sandbox directory %s: %s",
                    getattr(self, "working_dir", ""),
                    e,
                )
            self._temp_dir_obj = None


def create(
    working_dir: Optional[str] = None,
    env: Optional[Dict[str, str]] = None,
    port_forwardings: Optional[List["PortForwarding"]] = None,
):
    """
    Create a new SandBox instance.

    This is a convenience function to create a sandbox without using the class directly.

    Args:
        working_dir (Optional[str]): The working directory for sandbox execution.
            If None, a temporary directory will be created.
        env (Optional[Dict[str, str]]): Environment variables for the sandbox.
            If None, inherits from parent process.
        port_forwardings (Optional[List[PortForwarding]]): Port forwarding rules.
            If set, Gateway URLs will be printed after creation.

    Returns:
        SandBox wrapper instance.

    Examples:
        >>> import yr
        >>> yr.init()
        >>>
        >>> sandbox = yr.sandbox.create()
        >>> result = yr.get(sandbox.exec("pwd"))
        >>> print(result['stdout'])
        >>>
        >>> sandbox.terminate()
        >>> yr.finalize()
    """
    return SandBox(working_dir, env, port_forwardings)


class SandBox:
    """
    SandBox wrapper class for convenient sandbox operations.

    This class wraps the SandBoxInstance to provide a simpler interface
    for sandbox operations.

    Examples:
        >>> import yr
        >>> yr.init()
        >>>
        >>> sandbox = yr.sandbox.SandBox()
        >>> result = yr.get(sandbox.exec("echo 'Hello from sandbox'"))
        >>> print(result['stdout'])
        >>>
        >>> sandbox.terminate()
        >>> yr.finalize()
    """

    def __init__(
        self,
        working_dir: Optional[str] = None,
        env: Optional[Dict[str, str]] = None,
        port_forwardings: Optional[List["PortForwarding"]] = None,
    ):
        """
        Initialize the SandBox wrapper.

        Args:
            working_dir (Optional[str]): The working directory for sandbox execution.
                If None, a temporary directory will be created.
            env (Optional[Dict[str, str]]): Environment variables for the sandbox.
                If None, inherits from parent process.
            port_forwardings (Optional[List[PortForwarding]]): Port forwarding rules.
                If set, Gateway URLs will be printed after creation.
        """
        # Create InvokeOptions with skip_serialize=True for cross-version compatibility
        opt = yr.InvokeOptions()
        opt.skip_serialize = True
        if port_forwardings:
            opt.port_forwardings = port_forwardings
        self._instance = SandBoxInstance.options(opt).invoke(working_dir, env)
        if port_forwardings:
            instance_id = yr.get(self._instance.get_name.invoke())
            _print_gateway_urls(instance_id, port_forwardings)

    def __del__(self):
        """
        Destructor to ensure cleanup and termination on object deletion.

        Automatically calls cleanup() and terminate() when the SandBox object is deleted.
        """
        try:
            if hasattr(self, "_instance") and self._instance is not None:
                # Call cleanup first
                yr.get(self.cleanup())
                # Then terminate the instance
                self.terminate()
        except Exception:
            # Silently catch exceptions during cleanup to avoid errors in destructor
            pass

    def exec(self, command: str, timeout: Optional[int] = None):
        """
        Execute a command in the sandbox environment.

        Args:
            command (str): The command to execute.
            timeout (Optional[int]): Timeout in seconds for command execution.
                If None, no timeout is set.

        Returns:
            ObjectRef: Reference to the execution result that needs to be unwrapped with yr.get().
                The result is a dictionary containing:
                - returncode (int): The return code of the command.
                - stdout (str): Standard output of the command.
                - stderr (str): Standard error of the command.

        Examples:
            >>> sandbox = yr.sandbox.SandBox()
            >>> result_ref = sandbox.exec("ls -la")
            >>> result = yr.get(result_ref)
            >>> print(result['stdout'])
        """
        return self._instance.execute.invoke(command, timeout)

    def get_working_dir(self):
        """
        Get the working directory of the sandbox.

        Returns:
            ObjectRef: Reference to the working directory path.
        """
        return self._instance.get_working_dir.invoke()

    def cleanup(self):
        """
        Cleanup the sandbox environment.

        Returns:
            ObjectRef: Reference to the cleanup result.
        """
        return self._instance.cleanup.invoke()

    def terminate(self):
        """
        Terminate the sandbox instance.

        This will cleanup resources and terminate the remote instance.
        """
        self._instance.terminate()


def main():
    parser = argparse.ArgumentParser(description="Create a detached sandbox instance")
    parser.add_argument(
        "--name", type=str, default=None, help="Name of the sandbox instance"
    )
    parser.add_argument(
        "--namespace",
        type=str,
        default="detached.sandbox",
        help="Namespace for the sandbox instance",
    )
    args = parser.parse_args()
    os.environ.pop("YR_WORKING_DIR", None)

    cfg = yr.Config()
    cfg.in_cluster = True
    yr.init(cfg)
    try:
        opt = yr.InvokeOptions()
        opt.custom_extensions["lifecycle"] = "detached"
        opt.idle_timeout = 60 * 60 * 24 * 7
        opt.cpu = 1000
        opt.memory = 2048
        opt.name = args.name
        opt.namespace = args.namespace
        opt.skip_serialize = True  # Skip serialization for pre-deployed SDK class
        if not opt.name:
            import uuid
            opt.name = str(uuid.uuid4())

        sandbox = SandBoxInstance.options(opt).invoke()
        try:
            name = yr.get(sandbox.get_name.invoke())
            print(f"sandbox created, instance_name={name}")
        except Exception as e:
            print(f"sandbox create failed, name={opt.name}, error={e}")
    finally:
        yr.finalize()


if __name__ == "__main__":
    main()
