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

from __future__ import annotations

import json
import logging
import os
import select
import signal
import ssl
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request
from collections import deque
from pathlib import Path
from typing import Optional

import tomli_w

try:
    import tomllib
except ImportError:
    from pip._vendor import tomli as tomllib

from yr.cli.component.base import ComponentConfig, ComponentLauncher
from yr.cli.component.registry import LAUNCHER_CLASSES, PREPEND_CHAR_OVERRIDES, get_depends_on_overrides
from yr.cli.config import ConfigResolver
from yr.cli.const import (
    DEFAULT_DEPLOY_DIR,
    DEFAULT_LOG_DIR,
    DEFAULT_MASTER_INFO_PATH,
    SESSION_JSON_PATH,
    SESSION_LATEST_PATH,
    StartMode,
)
from yr.cli.utils import wait_pid_exit

logger = logging.getLogger(__name__)
print_logger = logging.getLogger("print")

# Sections from values.toml that should be propagated into the
# generated "join cluster" config snippet for agents.
JOIN_VALUES_INCLUDE_SECTIONS: list[str] = [
    "meta_store",
    "fs",
    "etcd",
    "ds_master",
    "function_master",
]

# Per-section fields that should be excluded when exporting join
# config. Keys are section names and values are sets of field names to
# drop from that section.
JOIN_VALUES_EXCLUDE_FIELDS: dict[str, set[str]] = {
    # Agents do not need the binary path for ds_master.
    "ds_master": {"bin_path"},
    "etcd": {"bin_path"},
    "function_master": {"bin_path"},
    "fs": {"log.path"},
}


def _parse_addr(addr: str) -> tuple[str, str]:
    """Parse an address string into (host, port).

    Handles values that may contain a scheme prefix like http://.
    """
    hostport = addr.split("//")[-1]
    host, port = hostport.rsplit(":", 1)
    return host, port


def _atomic_write_json(path: Path, data: dict[str, any]) -> None:
    """Atomically write JSON to path (write temp + fsync + replace)."""
    temp_file = Path(str(path) + ".tmp")
    try:
        with temp_file.open("w") as f:
            json.dump(data, f, indent=2)
            f.write("\n")
            f.flush()
            os.fsync(f.fileno())
        temp_file.replace(path)
    except Exception:
        if temp_file.exists():
            temp_file.unlink()
        raise


def delete_by_path(d: dict[str, any], path: str) -> None:
    """Delete a nested key from dict by dot path.

    Example: path='log.path' -> del d['log']['path'] if exists
    """
    parts = path.split(".")
    cur: any = d
    for key in parts[:-1]:
        if not isinstance(cur, dict):
            return
        cur = cur.get(key)
        if cur is None:
            return
    if isinstance(cur, dict):
        cur.pop(parts[-1], None)


def _flatten_dot_keys(prefix: str, obj: any) -> list[tuple[str, any]]:
    """Flatten nested dicts into dot-notation keys.

    Example:
        _flatten_dot_keys('values', {'etcd': {'port': 123}})
        -> [('values.etcd.port', 123)]
    """
    items: list[tuple[str, any]] = []
    if isinstance(obj, dict):
        for k, v in obj.items():
            key = f"{prefix}.{k}" if prefix else str(k)
            items.extend(_flatten_dot_keys(key, v))
    else:
        items.append((prefix, obj))
    return items


def _to_toml_literal(value: any) -> str:
    """Turn a Python value into a single-line TOML literal suitable for KEY=VALUE."""
    tmp = tomli_w.dumps({"v": value})
    one_line = " ".join(line.strip() for line in tmp.splitlines() if line.strip())
    prefix = "v = "
    if one_line.startswith(prefix):
        return one_line[len(prefix):]  # fmt: skip
    return one_line


def _filter_join_values(current_cfg: dict[str, any], default_cfg: dict[str, any]) -> dict[str, any]:
    """Filter values.toml for join instructions (include sections + exclude fields)."""
    filtered_values: dict[str, any] = {}
    for section in JOIN_VALUES_INCLUDE_SECTIONS:
        section_cfg = current_cfg.get(section, {}) or {}
        if not isinstance(section_cfg, dict):
            filtered_values[section] = section_cfg
            continue

        section_dict: dict[str, any] = dict(section_cfg)
        excludes = JOIN_VALUES_EXCLUDE_FIELDS.get(section, set())
        for path in excludes:
            if "." in path:
                delete_by_path(section_dict, path)
            else:
                section_dict.pop(path, None)
        filtered_values[section] = section_dict

    def get_custom_values_only(custom: dict[str, any], default: dict[str, any]) -> dict[str, any]:
        custom_only = {}
        for key, custom_val in custom.items():
            default_val = default.get(key)
            if default_val is None or type(custom_val) is not type(default_val):
                custom_only[key] = custom_val
                continue
            if isinstance(custom_val, dict) and isinstance(default_val, dict):
                sub_result = get_custom_values_only(custom_val, default_val)
                if sub_result:
                    custom_only[key] = sub_result
            elif custom_val != default_val:
                custom_only[key] = custom_val

        return custom_only

    return get_custom_values_only(filtered_values, default_cfg)


def _format_join_help(filtered_values: dict[str, any]) -> str:
    join_values = {"values": filtered_values}

    join_cmd_parts: list[str] = ["yr start"]
    for key, value in _flatten_dot_keys("values", filtered_values):
        literal = _to_toml_literal(value)
        join_cmd_parts.append(f"-s '{key}={literal}'")
    join_cmd = " ".join(join_cmd_parts)

    return f"""To join an existing cluster, execute the following commands in your shell on worker nodes:

{join_cmd}

OR

mkdir -p /etc/yuanrong/ && cat << EOF > /etc/yuanrong/config.toml && yr start
{tomli_w.dumps(join_values)}
EOF
"""


class SessionManager:
    def __init__(
        self,
        session_id: str,
        mode: StartMode,
        session_file: Optional[str] = None,
    ) -> None:
        self.session_file = Path(session_file or SESSION_JSON_PATH)
        self.session_data = {
            "components": {},
            "start_time": None,
            "session_id": session_id,
            "cluster_info": {},
        }
        self.mode = mode

    def save_session(self, components: dict[str, ComponentConfig], config: dict[str, any]) -> None:
        logger.info(f"Saving session data to {self.session_file}...")
        self.session_data["start_time"] = time.ctime()

        for name, comp in components.items():
            if comp.pid:
                self.session_data["components"][name] = {
                    "pid": comp.pid,
                    "status": comp.status.value,
                    "start_time": comp.start_time,
                    "cmd": comp.args,
                    "env_vars": comp.env_vars,
                    "restart_count": comp.restart_count,
                }

        etcd_ip, etcd_port = _parse_addr(config["etcd"]["args"]["listen-client-urls"])
        _, etcd_peer_port = _parse_addr(config["etcd"]["args"]["listen-peer-urls"])
        ds_master_ip, ds_master_port = _parse_addr(config["ds_master"]["args"]["master_address"])
        fm_ip, fm_port = _parse_addr(config["function_master"]["args"]["ip"])
        _, fp_port = _parse_addr(config["function_proxy"]["args"]["address"])
        _, ds_worker_port = _parse_addr(config["ds_worker"]["args"]["worker_address"])
        frontend_port = (
            config["values"]["frontend"].get("port") if config["mode"][self.mode.value].get("frontend") else None
        )
        self.session_data["mode"] = self.mode.value
        self.session_data["cluster_info"] = {
            "for-join": {
                "function_master.ip": fm_ip,
                "function_master.port": fm_port,
                "etcd.ip": etcd_ip,
                "etcd.port": etcd_port,
                "etcd.peer_port": etcd_peer_port,
                "etcd.addresses": config["values"]["etcd"]["address"],
                "ds_master.ip": ds_master_ip,
                "ds_master.port": ds_master_port,
                "function_proxy.port": fp_port,
                "function_proxy.grpc_port": config["function_proxy"]["args"]["grpc_listen_port"],
                "ds_worker.port": ds_worker_port,
                "agent.ip": config["function_agent"]["args"]["ip"],
                "frontend.port": frontend_port,
            },
            "daemon": {
                "pid": os.getpid(),
            },
        }

        _atomic_write_json(self.session_file, self.session_data)

    def load_session(self) -> dict[str, any]:
        if self.session_file.exists():
            with self.session_file.open() as f:
                return json.load(f)
        return {}

    def clear_session(self) -> None:
        if self.session_file.exists():
            self.session_file.unlink()

    def is_session_active(self) -> bool:
        if not self.session_file.exists():
            logger.error(f"No active session found in {self.session_file}.")
            return False
        session = self.load_session()
        if not session or "session_id" not in session:
            return False
        for comp_info in session.get("components", {}).values():
            pid = comp_info.get("pid")
            if pid:
                try:
                    os.kill(pid, 0)
                except OSError:
                    logger.warning(f"Process {comp_info.get('name', 'unknown')} with PID {pid} not running.")
                    continue
                else:
                    return True
        return False


class SystemLauncher:
    def __init__(
        self,
        config_path: Path,
        cli_dir: Path,
        mode: Optional[StartMode] = None,
        overrides: Optional[tuple[str, ...]] = None,
        session_file: Optional[str] = None,
        render: bool = True,
    ):
        self.mode = mode
        self.resolver = ConfigResolver(config_path, cli_dir, mode, overrides, render)
        self.session_manager = SessionManager(
            self.resolver.runtime_context["time"],
            mode,
            self.resolver.runtime_context["deploy_path"] / "session.json" if mode else session_file,
        )
        self.components: dict[str, ComponentLauncher] = {}
        self.processes: dict[str, subprocess.Popen] = {}
        self._stopped: bool = False
        self._had_uncaught_exception: bool = False
        self._monitor_thread: Optional[threading.Thread] = None
        self._monitor_interval: int = 3  # seconds between health checks

        self._startup_complete: bool = False
        self._startup_cancel_wfd: Optional[int] = None

        self._register_component_launchers()

    # -------------------------------------------------------------------------
    # status helpers
    # -------------------------------------------------------------------------

    @staticmethod
    def _extract_arg_value(args: list[str], key: str) -> Optional[str]:
        prefix = f"{key}="
        for arg in args:
            if arg == key:
                return ""
            if arg.startswith(prefix):
                return arg[len(prefix):]  # fmt: skip
        return None

    @staticmethod
    def _resolve_ssl_path(value: Optional[str], base_path: Optional[str]) -> Optional[str]:
        if value is None:
            return None
        value = value.strip()
        if not value:
            return None
        path = Path(value)
        if not path.is_absolute() and base_path:
            path = Path(base_path) / path
        return str(path)

    @staticmethod
    def _build_tls_client_context(
        ca_path: Optional[str],
        cert_path: str,
        key_path: str,
    ) -> Optional[ssl.SSLContext]:
        try:
            context = ssl.create_default_context(cafile=ca_path)
            context.check_hostname = False
            context.verify_mode = ssl.CERT_REQUIRED
            context.load_cert_chain(certfile=cert_path, keyfile=key_path)
        except OSError as exc:
            logger.warning("Failed to build TLS context for function_master: %s", exc)
            return None
        else:
            return context

    @staticmethod
    def _extract_scalars(section: any) -> dict[str, int]:
        if not isinstance(section, dict):
            return {}
        resources = section.get("resources")
        if not isinstance(resources, dict):
            return {}
        out: dict[str, int] = {}
        for name, entry in resources.items():
            if isinstance(entry, dict):
                val = (entry.get("scalar") or {}).get("value")
                if isinstance(val, (int, float)) and not isinstance(val, bool):
                    out[str(name)] = int(val)
        return out

    @staticmethod
    def _fmt_resources(values: dict[str, int]) -> str:
        return " ".join(f"{k}={v}" for k, v in sorted(values.items()))

    @staticmethod
    def _signal_startup_result(startup_result_wfd: int, success: bool) -> None:
        """Daemon-side: signal startup result to parent.

        Parent uses this to fail fast when the daemon exits/panics during startup
        before a valid session file is written.
        """
        try:
            os.write(startup_result_wfd, b"1" if success else b"0")
        except OSError:
            logger.debug("Failed to signal startup result to parent")
        finally:
            try:
                os.close(startup_result_wfd)
            except OSError:
                logger.debug("Failed to signal startup result to parent")

    @staticmethod
    def _fetch_url(
        url: str,
        timeout: float = 2.0,
        ssl_context: Optional[ssl.SSLContext] = None,
        as_json: bool = False,
    ) -> Optional[any]:
        try:
            req = urllib.request.Request(url, headers={"Accept": "application/json"})  # noqa: S310
            with urllib.request.urlopen(req, timeout=timeout, context=ssl_context) as resp:  # noqa: S310
                content = resp.read().decode("utf-8", errors="replace")
                return json.loads(content) if as_json else content.strip()
        except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as exc:
            logger.warning("  Failed to fetch %s: %s", url, exc)
            return None

    @staticmethod
    def _print_resource_info(unit: dict[str, any]) -> None:
        uid, owner = unit.get("id"), unit.get("ownerId")
        if uid or owner:
            print_logger.info(f"  Resource: id={uid} owner={owner}")

    def load_components(self) -> None:
        components_config = self.resolver.rendered_config["mode"].get(self.mode.value, {})

        for comp_name, enable in components_config.items():
            if enable:
                launcher_class: Optional[type[ComponentLauncher]] = self.launcher_classes.get(comp_name)
                if launcher_class is None:
                    logger.error(f"Unknown component: {comp_name}")
                    sys.exit(1)

                launcher = launcher_class(comp_name, self.resolver)
                self._apply_component_overrides(comp_name, launcher)
                self.components[comp_name] = launcher

    def start_all(self) -> bool:
        logger.info("Starting system components...")
        self._prepare_environment()

        # Create a startup-cancel pipe. Parent keeps write-end; daemon keeps read-end.
        # Parent can send a single byte on Ctrl-C; daemon stops if startup isn't complete yet.
        cancel_rfd, cancel_wfd = os.pipe()
        # Daemon -> parent: one byte indicates startup success/failure.
        # If daemon panics/exits before writing session.json, parent will see EOF
        # and fail fast (instead of waiting until timeout).
        startup_result_rfd, startup_result_wfd = os.pipe()
        daemon_pid = self._daemonize()
        if daemon_pid > 0:
            # Parent process
            os.close(startup_result_wfd)
            try:
                return self._run_parent_wait_ready(
                    daemon_pid=daemon_pid,
                    cancel_rfd=cancel_rfd,
                    cancel_wfd=cancel_wfd,
                    startup_result_rfd=startup_result_rfd,
                )
            finally:
                try:
                    os.close(startup_result_rfd)
                except OSError:
                    logger.debug("Failed to close startup result read fd")

        # Child (daemon) process continues here
        os.close(startup_result_rfd)
        return self._run_daemon_start(
            cancel_rfd=cancel_rfd,
            cancel_wfd=cancel_wfd,
            startup_result_wfd=startup_result_wfd,
        )

    def stop_all(self, force: bool = False):
        logger.info("Stopping system components...")

        for launcher in self.components.values():
            launcher.terminate(force=force)
        self.session_manager.clear_session()
        logger.info("✅ All components stopped.")

    def stop_daemon_from_session(self, force: bool = False):
        if not Path(self.session_manager.session_file).exists():
            logger.error("Error: No active session found.")
            return False
        session = self.session_manager.load_session()
        if not session or "cluster_info" not in session or "daemon" not in session["cluster_info"]:
            logger.error("Error: Invalid session file.")
            return False

        daemon_info = session["cluster_info"]["daemon"]
        daemon_pid = daemon_info.get("pid")
        if not daemon_pid:
            logger.error("Error: No daemon PID found in session file.")
            return False

        try:
            if force:
                os.kill(daemon_pid, signal.SIGKILL)
                logger.info(f"Killed daemon process (PID: {daemon_pid})")
            else:
                os.kill(daemon_pid, signal.SIGTERM)
                timeout_s = 40
                logger.info(f"Sent SIGTERM to daemon process (PID: {daemon_pid}) with timeout {timeout_s} seconds")
                exited = wait_pid_exit(daemon_pid, time.time() + timeout_s)
                if exited:
                    logger.info(f"Daemon process (PID: {daemon_pid}) has exited.")
                else:
                    logger.warning(
                        f'Daemon process (PID: {daemon_pid}) did not exit in time. Try calling "yr stop --force"',
                    )
        except ProcessLookupError:
            logger.error(f"Daemon process (PID: {daemon_pid}) not found")
        except Exception as e:
            logger.error(f"Error stopping daemon process: {e}")

        return True

    def stop_components_from_session(self, force: bool = False):
        if not self.session_manager.session_file.exists():
            logger.error("Error: No active session found.")
            return False
        session = self.session_manager.load_session()
        if not session or "components" not in session:
            logger.error("Error: Invalid session file.")
            return False

        logger.info(f"Stopping components from session file {self.session_manager.session_file}...")

        for comp_name, comp_info in reversed(list(session["components"].items())):
            pid = comp_info.get("pid")
            if pid:
                try:
                    if force:
                        os.kill(pid, signal.SIGKILL)
                        logger.info(f"Killed {comp_name} (PID: {pid})")
                    else:
                        os.kill(pid, signal.SIGTERM)
                        logger.info(f"Terminated {comp_name} (PID: {pid})")
                except ProcessLookupError:
                    logger.warning(f"Process {comp_name} (PID: {pid}) not found")
                except Exception:
                    logger.exception(f"Error stopping {comp_name}")

        deadline = time.time() + 90  # wait up to 90 seconds for graceful exit
        logger.info("Waiting for components to exit gracefully with timeout 90 seconds...")
        for comp_name, comp_info in reversed(list(session["components"].items())):
            pid = comp_info.get("pid")
            if pid:
                exited = wait_pid_exit(pid, deadline)
                if exited:
                    logger.info(f"{comp_name} (PID: {pid}) has exited.")
                else:
                    logger.warning(f"{comp_name} (PID: {pid}) did not gracefully exit in time.")
                    try:
                        os.kill(pid, signal.SIGKILL)
                    except ProcessLookupError:
                        logger.warning(f"Process {comp_name} (PID: {pid}) already exited.")
                    logger.info(f"Killed {comp_name} (PID: {pid}) after timeout.")

        self.session_manager.clear_session()
        logger.info("✅ All components stopped.")
        return True

    def health(self) -> bool:
        healthy = True
        if self.session_manager.is_session_active():
            session = self.session_manager.load_session()
            print_logger.info("System Status: RUNNING")
            print_logger.info(f"Started at: {session.get('start_time')}")
            print_logger.info(f"Session ID: {session.get('session_id')}")
            print_logger.info("Components:")
            for comp_name, comp_info in session.get("components", {}).items():
                pid = comp_info.get("pid")
                status = "RUNNING"
                if pid:
                    try:
                        os.kill(pid, 0)
                    except ProcessLookupError:
                        healthy = False
                        status = "STOPPED"
                print_logger.info(f"  {comp_name:<18} {status} (PID: {pid})")
        else:
            print_logger.info("System Status: STOPPED")
            healthy = False
        return healthy

    def status(self) -> bool:
        """Display cluster status by querying the function_master scheduler."""
        if not self.session_manager.session_file.exists():
            logger.info("System Status: STOPPED")
            logger.info("No session file found at %s", self.session_manager.session_file)
            return False

        session = self.session_manager.load_session()
        if not session:
            logger.info("System Status: STOPPED")
            logger.info("Session file exists but is empty/invalid: %s", self.session_manager.session_file)
            return False

        if session.get("mode") != StartMode.MASTER.value:
            logger.error("Cluster status is only available in 'master' mode.")
            return False
        cluster_info = session.get("cluster_info", {}).get("for-join", {})
        scheduler_ip = cluster_info.get("function_master.ip")
        scheduler_port = cluster_info.get("function_master.port")

        if not (scheduler_ip and scheduler_port):
            return True

        tls_context = self._get_function_master_tls_context_from_session(session)
        scheme = "https" if tls_context else "http"
        base_url = f"{scheme}://{scheduler_ip}:{scheduler_port}/global-scheduler"

        self._print_cluster_summary(base_url, ssl_context=tls_context)
        return True

    def _register_component_launchers(self):
        self.launcher_classes = LAUNCHER_CLASSES
        self.prepend_char_overrides = PREPEND_CHAR_OVERRIDES
        self.depends_on_overrides = get_depends_on_overrides(self.mode)

    def _register_exit_hooks(self, for_daemon: bool) -> None:
        """Register signal handlers and exception hooks.

        Args:
            for_daemon: If True, register handlers appropriate for daemon process.
                       If False, register handlers for pre-daemon process.
        """

        # Handle termination signals like Ctrl-C / kill TERM.
        def _signal_handler(signum, frame):  # type: ignore[unused-argument]
            logger.warning(f"Received signal {signum}, stopping all components...")
            if for_daemon:
                # In daemon process, signal monitor thread to exit
                self._stopped = True
                # Also remove session file to prevent restart attempts
                self.session_manager.clear_session()
            else:
                # In pre-daemon process
                self._on_exit(force=True)

        for sig in (signal.SIGINT, signal.SIGTERM):
            try:
                signal.signal(sig, _signal_handler)
            except Exception:
                logger.error(f"Unable to register handler for signal {sig}")

        # Handle uncaught exceptions as "panic" and stop components.
        if for_daemon:
            # Only set excepthook in pre-daemon process
            # (daemon stdout/stderr are redirected, so exceptions are logged differently)
            previous_hook = sys.excepthook

            def _excepthook(exc_type, exc_value, traceback):
                self._had_uncaught_exception = True
                logger.error("Uncaught exception, stopping components...", exc_info=(exc_type, exc_value, traceback))
                self._on_exit(force=True)
                previous_hook(exc_type, exc_value, traceback)

            sys.excepthook = _excepthook

    def _on_exit(self, force: bool = False) -> None:
        if self._stopped:
            return
        self._stopped = True
        try:
            self.stop_all(force=force)
        except Exception:
            logger.exception("Error while stopping components on exit")

    def _prepare_environment(self) -> None:
        deploy_path = self.resolver.runtime_context["deploy_path"]

        deploy_path.mkdir(parents=True, exist_ok=True)
        symlink_path = Path(SESSION_LATEST_PATH)
        if symlink_path.is_symlink() or symlink_path.exists():
            symlink_path.unlink()
        # /tmp/yr/session_latest -> real deploy path. we only need to focus on dealing with /tmp/yr/session_latest
        symlink_path.symlink_to(deploy_path)
        logger.info(f"Created symlink {symlink_path} -> {deploy_path}")
        Path(DEFAULT_LOG_DIR).mkdir(parents=True, exist_ok=True)
        Path(DEFAULT_DEPLOY_DIR).mkdir(parents=True, exist_ok=True)
        os.chdir(DEFAULT_DEPLOY_DIR)
        self.resolver.runtime_context["deploy_path"] = deploy_path

    def _apply_component_overrides(self, comp_name: str, launcher: ComponentLauncher) -> None:
        """Apply per-component configuration tweaks after a launcher is constructed."""
        override_char = self.prepend_char_overrides.get(comp_name)
        # Only override when the component is still using the default value,
        # so any explicit custom config is preserved.
        if override_char and launcher.component_config.prepend_char == "--":
            launcher.component_config.prepend_char = override_char
        depends_on_override = self.depends_on_overrides.get(comp_name)
        if depends_on_override is not None:
            launcher.component_config.depends_on = depends_on_override

    def _start_component(self, component_name: str) -> Optional[subprocess.Popen]:
        launcher = self.components[component_name]
        process = launcher.launch()
        self.processes[component_name] = process

        return process

    def _start_components_in_order(self, components_order: list[str]) -> bool:
        for comp_name in components_order:
            logger.debug(f"Starting {comp_name}...")
            process = self._start_component(comp_name)
            if not process:
                logger.error(f"Failed to start {comp_name}")
                return False
            if not self.components[comp_name].wait_until_healthy():
                logger.debug(f"{comp_name} failed health check after start.")
                return False
        return True

    def _finalize_startup(self) -> None:
        self.session_manager.save_session(
            {name: launcher.component_config for name, launcher in self.components.items()},
            self.resolver.rendered_config,
        )
        if self.mode == StartMode.MASTER:
            write_old_current_master_info(self.session_manager.session_data)

        # From this point, parent exiting is expected and should not stop the daemon.
        self._startup_complete = True

        logger.info("✅ All components are healthy!")
        logger.info(f"Session saved to: {self.session_manager.session_file}")

    def _run_parent_wait_ready(
        self,
        daemon_pid: int,
        cancel_rfd: int,
        cancel_wfd: int,
        startup_result_rfd: int,
    ) -> bool:
        os.close(cancel_rfd)
        self._startup_cancel_wfd = cancel_wfd
        try:
            return self._wait_for_daemon_ready(daemon_pid, startup_result_rfd=startup_result_rfd)
        finally:
            if self._startup_cancel_wfd is not None:
                try:
                    os.close(self._startup_cancel_wfd)
                except OSError:
                    logger.debug("Failed to close startup cancel write fd")
                self._startup_cancel_wfd = None

    def _run_daemon_start(self, cancel_rfd: int, cancel_wfd: int, startup_result_wfd: int) -> bool:
        os.close(cancel_wfd)
        self._start_startup_cancel_watcher(cancel_rfd)

        try:
            components_order = self._get_start_order()
            if not self._start_components_in_order(components_order):
                logger.debug("❌ Some components failed health checks.")
                self.stop_all(force=True)
                self._signal_startup_result(startup_result_wfd, success=False)
                return False
            logger.info("✅ All components started and passed health checks.")

            self._finalize_startup()
            self._signal_startup_result(startup_result_wfd, success=True)

            self._start_monitor_daemon()
            if self._monitor_thread:
                self._monitor_thread.join()
            return True
        except Exception:
            self._signal_startup_result(startup_result_wfd, success=False)
            raise

    def _start_startup_cancel_watcher(self, cancel_rfd: int) -> None:
        """Daemon-side watcher: if parent cancels during startup, stop launching.

        The parent may exit normally after startup is complete; in that case this watcher
        must not kill the daemon.
        """

        t = threading.Thread(
            target=self._startup_cancel_watch_loop,
            args=(cancel_rfd,),
            daemon=True,
            name="startup-cancel-watcher",
        )
        t.start()

    def _startup_cancel_watch_loop(self, cancel_rfd: int) -> None:
        try:
            while True:
                readable, _, _ = select.select([cancel_rfd], [], [], 1.0)
                if not readable:
                    if self._startup_complete:
                        return
                    continue
                data = os.read(cancel_rfd, 1)
                # data == b'' means parent closed the pipe.
                # Treat as cancel only if startup hasn't completed.
                if data in (b"", b"X"):
                    if not self._startup_complete:
                        logger.warning("Startup canceled by parent; stopping components...")
                        self._stopped = True
                        try:
                            self.stop_all(force=True)
                        finally:
                            sys.exit(1)
                    return
        finally:
            try:
                os.close(cancel_rfd)
            except OSError:
                logger.debug("Failed to close startup cancel read fd")

    def _wait_for_daemon_ready(
        self,
        daemon_pid: int,
        timeout: int = 120,
        startup_result_rfd: Optional[int] = None,
    ) -> bool:
        """Parent process waits for daemon to signal readiness via session file.

        Args:
            daemon_pid: PID of the intermediate child (will exit after second fork)
            timeout: Maximum time to wait for readiness in seconds

        Returns:
            True if system is ready, False if timeout
        """
        try:
            os.waitpid(daemon_pid, 0)
        except ChildProcessError:
            pass  # Already exited

        logger.info("Waiting for cluster to be ready...")
        deadline = time.time() + timeout
        check_interval = 0.5
        got_startup_success_signal = False

        try:
            while time.time() < deadline:
                if self.session_manager.session_file.exists():
                    session = self.session_manager.load_session()
                    if session and "cluster_info" in session:
                        logger.info("✅ All components are healthy!")
                        self._print_status_from_session(session)
                        print_logger.info(
                            "❕Check logs in %s/logs for details.",
                            self.resolver.runtime_context["deploy_path"],
                        )
                        return True
                    logger.warning("Session file found but invalid, this should not happen...")

                if startup_result_rfd is not None:
                    readable, _, _ = select.select([startup_result_rfd], [], [], 0)
                    if readable:
                        data = os.read(startup_result_rfd, 1)
                        if data == b"1":
                            got_startup_success_signal = True
                        elif data in (b"0", b"") and not self.session_manager.session_file.exists():
                            logger.error("❌ Daemon exited or reported startup failure")
                            logger.info(f"❕Check {self.resolver.runtime_context['deploy_path']}/logs for details.")
                            return False

                time.sleep(check_interval)
        except KeyboardInterrupt:
            # Ctrl-C in the parent should cancel daemon startup if not complete yet.
            logger.warning("Interrupted by user; canceling startup...")
            if self._startup_cancel_wfd is not None:
                try:
                    os.write(self._startup_cancel_wfd, b"X")
                except OSError:
                    logger.debug("Failed to send startup cancel signal to daemon")
            return False

        if got_startup_success_signal:
            logger.error(f"❌ Timeout waiting for session to become ready after daemon reported success ({timeout}s)")
        else:
            logger.error(f"❌ Timeout waiting for system to be ready after {timeout}s")
        return False

    def _print_status_from_session(self, session: dict[str, any]) -> None:
        """Print component status from session data (for parent process)."""
        logger.info("Component Status:")
        logger.info("-" * 60)
        logger.info(f"{'Component':<15} {'PID':<8} {'Status':<12}")
        logger.info("-" * 60)

        for comp_name, comp_info in session.get("components", {}).items():
            pid = comp_info.get("pid", "N/A")
            status = comp_info.get("status", "unknown")
            logger.info(f"{comp_name:<15} {pid:<8} {status:<12}")
        logger.info("-" * 60)
        if self.mode == StartMode.MASTER:
            self._print_join_info()

    def _start_monitor_daemon(self) -> None:
        """Start a background daemon thread that monitors component health and restarts failed ones."""
        if self._monitor_thread is not None and self._monitor_thread.is_alive():
            logger.info("Monitor daemon already running")
            return

        self._monitor_thread = threading.Thread(target=self._monitor_loop, daemon=True, name="component-monitor")
        self._monitor_thread.start()

    def _monitor_loop(self) -> None:
        logger.info("🔍 Monitor daemon started")
        while not self._stopped:
            time.sleep(self._monitor_interval)
            if self._stopped:
                logger.info("Monitor daemon stopping due to signal")
                break

            update_session = False
            for comp_name, launcher in list(self.components.items()):
                process = launcher.component_config.process
                if process is None:
                    continue
                exit_code = process.poll()
                if exit_code is None:
                    continue

                logger.warning(
                    f"Component {comp_name} (PID: {launcher.component_config.pid}) "
                    f"exited with code {exit_code}. Restarting...",
                )
                try:
                    new_process = launcher.restart()
                    if new_process:
                        self.processes[comp_name] = new_process
                        launcher.component_config.restart_count += 1
                        update_session = True
                        logger.info(f"✅ Successfully restarted {comp_name} (new PID: {launcher.component_config.pid})")
                    else:
                        logger.error(f"❌ Failed to restart {comp_name}")
                except Exception:
                    logger.exception(f"❌ Error restarting {comp_name}")

            if update_session:
                self.session_manager.save_session(
                    {name: lnch.component_config for name, lnch in self.components.items()},
                    self.resolver.rendered_config,
                )

        logger.info("🔍 Monitor daemon stopped")
        # Clean shutdown: stop all components
        self.stop_all(force=False)

    def _daemonize(self) -> int:
        """Fork the process to run in background.

        Returns:
            In parent: the daemon PID (> 0)
            In child (daemon): 0
        """
        deploy_path = self.resolver.runtime_context["deploy_path"]

        try:
            pid = os.fork()
            if pid > 0:
                # Parent process - return daemon pid, don't exit yet
                # Parent will wait for readiness signal
                return pid
        except OSError as e:
            logger.error(f"First fork failed: {e}")
            sys.exit(1)

        # Child process continues - decouple from parent environment
        os.setsid()
        os.umask(0)

        try:
            pid = os.fork()
            if pid > 0:
                sys.exit(0)
        except OSError as e:
            logger.error(f"Second fork failed: {e}")
            sys.exit(1)

        sys.stdout.flush()
        sys.stderr.flush()
        log_file = deploy_path / "logs" / "yr_daemon.log"
        with log_file.open("a") as f:
            os.dup2(f.fileno(), sys.stdout.fileno())
            os.dup2(f.fileno(), sys.stderr.fileno())

        with Path("/dev/null").open("r") as f:
            os.dup2(f.fileno(), sys.stdin.fileno())
        self._register_exit_hooks(for_daemon=True)

        return 0  # Child returns 0

    def _get_start_order(self) -> list[str]:
        """Determine component start order based on dependency graph (topological sort).

        Each launcher.component_config.depends_on contains a list of component names
        that must start before this component.
        """
        graph: dict[str, set[str]] = {name: set() for name in self.components}
        in_degree: dict[str, int] = dict.fromkeys(self.components.keys(), 0)

        for name, launcher in self.components.items():
            depends_on = launcher.component_config.depends_on or []
            for dep in depends_on:
                if dep not in self.components:
                    logger.error(f"Component '{name}' depends on unknown component '{dep}'")
                    sys.exit(1)
                if name not in graph[dep]:
                    graph[dep].add(name)
                    in_degree[name] += 1

        queue: deque[str] = deque([name for name, deg in in_degree.items() if deg == 0])
        order: list[str] = []

        while queue:
            cur = queue.popleft()
            order.append(cur)
            for nxt in graph[cur]:
                in_degree[nxt] -= 1
                if in_degree[nxt] == 0:
                    queue.append(nxt)

        if len(order) != len(self.components):
            cycle_nodes = [name for name, deg in in_degree.items() if deg > 0]
            logger.error(f"Detected cyclic or unresolved dependencies among components: {cycle_nodes}")
            sys.exit(1)

        logger.debug(f"Component start order (topological): {order}")
        return order

    def _print_join_info(self) -> None:
        values_cfg: dict[str, any] = self.resolver.rendered_config["values"]
        default_values_cfg_path: dict[str, any] = (
            Path(self.resolver.runtime_context["yr_package_path"]) / "cli" / "values.toml"
        )
        default_values_cfg = tomllib.loads(default_values_cfg_path.read_text())["values"]
        filtered_values = _filter_join_values(values_cfg, default_values_cfg)
        print_logger.info(_format_join_help(filtered_values))

    def _get_tls_client_paths_from_args(self, args: list[str]) -> Optional[tuple[Optional[str], str, str]]:
        ssl_enable = self._extract_arg_value(args, "--ssl_enable")
        if ssl_enable is None:
            return None
        if ssl_enable and ssl_enable.strip().lower() not in {"1", "true", "yes", "y", "on"}:
            return None
        base_path = self._extract_arg_value(args, "--ssl_base_path")
        ca_path = self._resolve_ssl_path(self._extract_arg_value(args, "--ssl_root_file"), base_path)
        cert_path = self._resolve_ssl_path(self._extract_arg_value(args, "--ssl_cert_file"), base_path)
        key_path = self._resolve_ssl_path(self._extract_arg_value(args, "--ssl_key_file"), base_path)
        if not cert_path or not key_path:
            return None
        return ca_path, cert_path, key_path

    def _get_function_master_tls_context_from_session(self, session: dict[str, any]) -> Optional[ssl.SSLContext]:
        args = session["components"]["function_master"]["cmd"]  # for valid master session, function_master should exist
        tls_paths = self._get_tls_client_paths_from_args(args)
        if tls_paths is None:
            return None
        ca_path, cert_path, key_path = tls_paths
        return self._build_tls_client_context(ca_path, cert_path, key_path)

    def _print_cluster_summary(self, base_url: str, ssl_context: Optional[ssl.SSLContext] = None) -> None:
        """Fetch and display cluster resource status from the global scheduler."""
        print_logger.info("Cluster Status:")

        # Display agent count
        agent_count = self._fetch_url(f"{base_url}/queryagentcount", ssl_context=ssl_context)
        if agent_count is not None:
            print_logger.info(f"  ReadyAgentsCount: {agent_count}")

        # Fetch and display resource information
        data = self._fetch_url(f"{base_url}/resources", ssl_context=ssl_context, as_json=True)
        if not isinstance(data, dict):
            return

        resource_unit = data.get("resource")
        if not isinstance(resource_unit, dict):
            return

        self._print_resource_info(resource_unit)
        self._print_resource_metrics(resource_unit)
        self._print_fragments(resource_unit)

    def _print_resource_metrics(self, unit: dict[str, any]) -> None:
        for key in ("capacity", "allocatable", "actualUse"):
            scalars = self._extract_scalars(unit.get(key))
            if scalars:
                print_logger.info(f"    {key}: {self._fmt_resources(scalars)}")

    def _print_fragments(self, unit: dict[str, any]) -> None:
        """Display fragment-level resource breakdown."""
        fragments = unit.get("fragment")
        if not isinstance(fragments, dict) or not fragments:
            return

        print_logger.info("    Fragments:")
        for frag_id in sorted(fragments):
            fragment = fragments[frag_id]
            if not isinstance(fragment, dict):
                continue

            # Collect non-empty resource metrics for this fragment
            metrics = []
            for metric_type in ("capacity", "allocatable", "actualUse"):
                scalars = self._extract_scalars(fragment.get(metric_type))
                if scalars:
                    metrics.append(f"{metric_type}: {self._fmt_resources(scalars)}")

            if metrics:
                print_logger.info(f"      {frag_id}:")
                for metric in metrics:
                    print_logger.info(f"        {metric}")


def write_old_current_master_info(session_data: dict[str, any]) -> None:
    master_info = Path(DEFAULT_MASTER_INFO_PATH)
    with master_info.open("w") as f:
        master_info = (
            "local_ip:{},master_ip:{},etcd_ip:{},etcd_port:{},etcd_peer_port:{},"
            "global_scheduler_port:{},ds_master_port:{},bus-proxy:{},bus:{},ds-worker:{},"
        ).format(
            session_data["cluster_info"]["for-join"]["agent.ip"],
            session_data["cluster_info"]["for-join"]["function_master.ip"],
            session_data["cluster_info"]["for-join"]["etcd.ip"],
            session_data["cluster_info"]["for-join"]["etcd.port"],
            session_data["cluster_info"]["for-join"]["etcd.peer_port"],
            session_data["cluster_info"]["for-join"]["function_master.port"],
            session_data["cluster_info"]["for-join"]["ds_master.port"],
            session_data["cluster_info"]["for-join"]["function_proxy.port"],
            session_data["cluster_info"]["for-join"]["function_proxy.grpc_port"],
            session_data["cluster_info"]["for-join"]["ds_worker.port"],
        )
        if session_data["cluster_info"]["for-join"]["frontend.port"]:
            master_info = master_info + f"frontend_port:{session_data['cluster_info']['for-join']['frontend.port']},"
        if len(session_data["cluster_info"]["for-join"]["etcd.addresses"]) > 1:
            for addr in session_data["cluster_info"]["for-join"]["etcd.addresses"]:
                master_info += f"etcd_addr_list:{addr},"
        master_info += "\n"
        f.write(master_info)
