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

import asyncio
import json
import logging
import os
import shutil
import subprocess
import sys
import traceback
import uuid
from dataclasses import dataclass
from typing import Any, Dict, Optional

import click
import requests
from requests.exceptions import RequestException

import yr
from yr.cli.exec import TerminalClientParams, run_client


__server_address = None
__ds_address = None
__client_cert = None
__client_key = None
__ca_cert = None
__insecure = False
__server_name = None
__user = None
__client_auth_type = "mutual"  # "mutual" or "one-way"
__jwt_token = None


@dataclass
class CliGlobalOptions:
    """Grouped CLI global flags (reduces parameter count on entry command)."""

    server_address: Optional[str] = None
    ds_address: Optional[str] = None
    client_cert: Optional[str] = None
    client_key: Optional[str] = None
    ca_cert: Optional[str] = None
    insecure: bool = False
    server_name: Optional[str] = None
    client_auth_type: str = "mutual"
    jwt_token: Optional[str] = None
    log_level: str = "INFO"
    user: str = "default"


@dataclass
class DeployOptions:
    """Grouped deploy command parameters."""

    backend: str = "ds"
    code_path: str = "."
    package_format: str = "zip"
    function_json: Optional[str] = None
    skip_package: bool = False
    update: bool = False
    requirements: Optional[str] = None


class FunctionName:
    """Function name components"""

    def __init__(self, raw_name: str, version: str = "latest"):
        self.raw_name = raw_name
        self.type = "0"
        self.service = None
        self.name = None
        self.version = version
        self.parse()

    def __str__(self):
        return f"{self.service}@{self.name}:{self.version}"

    def parse(self):
        """Parse function name into components"""
        name_part = self.raw_name
        if ":" in self.raw_name:
            name_part, version_part = self.raw_name.split(":", 1)
            self.version = version_part
        if "@" in name_part:
            parts = name_part.split("@")
            if len(parts) == 3:
                self.type = parts[0]
                self.service = parts[1]
                self.name = parts[2]
            elif len(parts) == 2:
                self.service = parts[0]
                self.name = parts[1]
            else:
                raise ValueError(f"Invalid function name format: {self.raw_name}")

    def full_name(self):
        """Get full function name"""
        return f"{self.type}@{self.service}@{self.name}:{self.version}"

    def full_name_no_version(self):
        """Get full function name without version"""
        return f"{self.type}@{self.service}@{self.name}"


class HTTPClient:
    """HTTP client with TLS authentication support (mutual or one-way) and JWT token"""

    def __init__(
        self,
        timeout: int = 30,
        client_cert: Optional[str] = None,
        client_key: Optional[str] = None,
        ca_cert: Optional[str] = None,
        insecure: bool = False,
        server_name: Optional[str] = None,
        client_auth_type: str = "mutual",  # "mutual" or "one-way"
        jwt_token: Optional[str] = None,
    ):
        self.timeout = timeout
        self.session = requests.Session()
        self.client_cert = client_cert
        self.client_key = client_key
        self.ca_cert = ca_cert
        self.insecure = insecure
        self.server_name = server_name
        self.client_auth_type = client_auth_type
        self.jwt_token = jwt_token

    def request(
        self,
        url: str,
        data: Dict[str, Any],
        headers: Optional[Dict[str, str]] = None,
        method: str = "POST",
    ) -> Dict[str, Any]:
        """
        Send POST JSON request

        Args:
            url: Target URL
            data: JSON data
            headers: Request headers

        Returns:
            Response data
        """

        default_headers = {
            "Content-Type": "application/json",
            "User-Agent": "DeploymentScript/1.0",
        }
        # Add JWT token if provided
        if self.jwt_token:
            default_headers["X-Auth"] = self.jwt_token
        if headers:
            default_headers.update(headers)

        logging.debug("%s to %s", method.lower(), url)
        logging.debug("headers: %s", json.dumps(default_headers, indent=2))
        logging.debug("body: %s", json.dumps(data, indent=2, ensure_ascii=False))

        # Configure certificates based on client_auth_type
        cert = None
        if self.client_auth_type == "mutual":
            # Mutual TLS: Use client certificate
            if self.client_cert and self.client_key:
                cert = (self.client_cert, self.client_key)
            elif self.client_cert:
                cert = self.client_cert
        # For "one-way" TLS, cert remains None (only verify server)
        verify = self.ca_cert if self.ca_cert else True
        if url.startswith("http://") and verify:
            url = url.replace("http://", "https://", 1)
        if self.insecure:
            import urllib3
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
            verify = False

        # If server_name is specified, use hostname overwrite
        if self.server_name:
            from requests.adapters import HTTPAdapter

            class SNIAdapter(HTTPAdapter):
                def __init__(self, sni_hostname, *args, **kwargs):
                    self.sni_hostname = sni_hostname
                    super().__init__(*args, **kwargs)

                def init_poolmanager(self, connections, maxsize, block=False, **pool_kwargs):
                    # server_hostname sets SNI, assert_hostname sets certificate verification hostname
                    pool_kwargs['server_hostname'] = self.sni_hostname
                    pool_kwargs['assert_hostname'] = self.sni_hostname
                    super().init_poolmanager(connections, maxsize, block, **pool_kwargs)

                def cert_verify(self, conn, url, verify, cert):
                    # Ensure SNI is set during connection
                    conn.server_hostname = self.sni_hostname
                    super().cert_verify(conn, url, verify, cert)

            # Create temporary session for this request
            temp_session = requests.Session()
            adapter = SNIAdapter(self.server_name)
            temp_session.mount("https://", adapter)

            response = temp_session.request(
                method.upper(),
                url,
                json=data,
                headers=default_headers,
                timeout=self.timeout,
                cert=cert,
                verify=verify,
            )
        else:
            response = self.session.request(
                method.upper(),
                url,
                json=data,
                headers=default_headers,
                timeout=self.timeout,
                cert=cert,
                verify=verify,
            )

        try:
            try:
                result = response.json() if response.content else {}
            except ValueError:
                result = response.content

            logging.debug("response: %s\n%s", response.headers, result)

            return {
                "success": response.status_code == 200,
                "error": result,
                "status_code": response.status_code,
                "data": result,
                "headers": dict(response.headers),
            }

        except RequestException as e:
            logging.debug("HTTP failed: %s", str(e))
            traceback.print_exc()
            return {
                "success": False,
                "error": str(e),
                "status_code": (
                    getattr(e.response, "status_code", None)
                    if hasattr(e, "response")
                    else None
                ),
            }


class YRContext:
    def __init__(self, server_address, ds_address, user=None):
        self.__server_address = server_address
        self.__ds_address = ds_address
        self.__user = user

    def __enter__(self):
        cfg = yr.Config()
        cfg.log_dir = "/tmp/yr_sessions/driver"
        if self.__user:
            cfg.tenant_id = self.__user
        if self.__server_address and self.__ds_address:
            cfg.server_address = self.__server_address
            cfg.ds_address = self.__ds_address
            return yr.init(cfg)
        if self.__server_address:
            cfg.server_address = self.__server_address
            cfg.in_cluster = False
            return yr.init(cfg)
        return yr.init(cfg)

    def __exit__(self, exc_type, exc_value, exc_traceback):
        yr.finalize()
        return False


def get_name_from_info(function_info):
    name = function_info.get("name")
    version = function_info.get("versionNumber")
    return FunctionName(name, version)


def deploy_function(function_json, user):
    http_client = HTTPClient(
        timeout=30,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    url = f"http://{__server_address}/admin/v1/functions"
    headler = {}
    if user:
        headler = {"X-Tenant-Id": user}
    resp = http_client.request(url, function_json, headers=headler, method="POST")
    if resp["success"]:
        return True, resp["data"]["function"]
    else:
        return False, resp


def update_function(function_json, user):
    name = function_json.get("name")
    if not name:
        raise RuntimeError("function name is required to update function.")
    http_client = HTTPClient(
        timeout=30,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    url = f"http://{__server_address}/admin/v1/functions/{name}"
    headler = {}
    if user:
        headler = {"X-Tenant-Id": user}
    resp = http_client.request(url, function_json, headers=headler, method="PUT")
    if resp["success"]:
        return True, resp["data"]["result"]
    else:
        return False, resp


def delete_function(function_name, user):
    http_client = HTTPClient(
        timeout=30,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    url = (
        f"http://{__server_address}/admin/v1/functions/"
        f"{function_name.full_name_no_version()}?versionNumber={function_name.version}"
    )
    headler = {}
    if user:
        headler = {"X-Tenant-Id": user}
    resp = http_client.request(url, {}, headers=headler, method="DELETE")
    if resp["success"]:
        return True, None
    else:
        return False, resp


def query_function(function_name, user=None):
    http_client = HTTPClient(
        timeout=30,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    if function_name is None:
        url = f"http://{__server_address}/admin/v1/functions"
    else:
        url = (
        f"http://{__server_address}/admin/v1/functions/"
        f"{function_name.full_name_no_version()}?versionNumber={function_name.version}"
    )
    headler = {}
    if user:
        headler = {"X-Tenant-Id": user}
    resp = http_client.request(url, {}, headers=headler, method="GET")
    if resp["success"]:
        if function_name is None:
            return True, resp["data"]["result"]["functions"]
        else:
            return True, resp["data"]["function"]
    else:
        return False, resp


def query_instances(user=None):
    """Query instance list for a specific tenant"""
    http_client = HTTPClient(
        timeout=30,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    tenant_id = user if user else "default"
    url = f"http://{__server_address}/api/instances?tenant_id={tenant_id}"
    resp = http_client.request(url, {}, method="GET")
    if resp["success"]:
        return True, resp["data"]
    else:
        return False, resp


def query_instance(instance_id, user=None):
    """Query single instance detail by instance ID"""
    http_client = HTTPClient(
        timeout=30,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    tenant_id = user if user else "default"
    url = f"http://{__server_address}/api/instances/{instance_id}?tenant_id={tenant_id}"
    resp = http_client.request(url, {}, method="GET")
    if resp["success"]:
        return True, resp["data"]
    else:
        return False, resp


def publish_function(function_name, publish_json, user=None):
    http_client = HTTPClient(
        timeout=30,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    url = f"http://{__server_address}/admin/v1/functions/{function_name.full_name_no_version()}/versions"
    headler = {}
    if user:
        headler = {"X-Tenant-Id": user}
    resp = http_client.request(url, publish_json, headers=headler, method="POST")
    if resp["success"]:
        return True, resp["data"]["function"]
    else:
        return False, resp


def install_requirements(requirements_file, target_dir):
    """
    Install Python dependencies from requirements file to target directory.

    Args:
        requirements_file: Path to requirements.txt file
        target_dir: Target directory to install dependencies

    Returns:
        True if successful, False otherwise
    """
    if not os.path.exists(requirements_file):
        logging.error(f"Requirements file not found: {requirements_file}")
        return False

    logging.info(f"Installing dependencies from {requirements_file} to {target_dir}...")

    try:
        result = subprocess.run(
            [
                sys.executable,
                "-m",
                "pip",
                "install",
                "-r",
                requirements_file,
                "-t",
                target_dir,
                "--no-warn-script-location",
            ],
            capture_output=True,
            text=True,
        )

        if result.returncode != 0:
            logging.error(f"Failed to install dependencies: {result.stderr}")
            return False

        logging.info(f"Successfully installed dependencies to {target_dir}")
        return True

    except Exception as e:
        logging.error(f"Error installing dependencies: {str(e)}")
        return False


def package(backend, code_path, package_format):
    real_code_path = os.path.realpath(code_path)
    file_name = f"code-{uuid.uuid4().hex}"
    archive_file = os.path.join("/tmp", file_name)
    if package_format == "zip":
        shutil.make_archive(archive_file, package_format, real_code_path)
    elif package_format == "img":
        result = subprocess.run(
            [
                "mkfs.erofs",
                "-E",
                "noinline_data",
                f"{archive_file}.{package_format}",
                real_code_path,
            ]
        )
        if result.returncode != 0:
            raise click.ClickException(f"mkfs.erofs failed with exit code {result.returncode}")
    else:
        raise click.ClickException(f"unknown format: {package_format}")
    if backend == "ds":
        with YRContext(__server_address, __ds_address):
            with open(f"{archive_file}.{package_format}", "rb") as f:
                yr.kv_write(file_name, f.read())

        package_key = f"ds://{file_name}.{package_format}"
    else:
        raise click.ClickException(f"not support backend: {backend}")
    return real_code_path, package_key


def invoke_function(function_name, payload, headers=None, user=None, timeout=30):
    if headers is None:
        headers = {}
    http_client = HTTPClient(
        timeout=timeout,
        client_cert=__client_cert,
        client_key=__client_key,
        ca_cert=__ca_cert,
        insecure=__insecure,
        server_name=__server_name,
        client_auth_type=__client_auth_type,
        jwt_token=__jwt_token,
    )
    url = f"http://{__server_address}/invocations/{user}/{str(function_name).replace('@', '/')}"
    resp = http_client.request(url, payload, headers=headers, method="POST")
    if resp["success"]:
        return True, resp["data"]
    else:
        return False, resp


@click.group(context_settings=dict(help_option_names=["-h", "--help"]))
@click.option(
    "--server-address",
    required=False,
    type=str,
    envvar="YR_SERVER_ADDRESS",
    help="YuanRong Server address",
)
@click.option(
    "--ds-address",
    required=False,
    type=str,
    envvar="YR_DS_ADDRESS",
    help="YuanRong DataSystem address",
)
@click.option(
    "--client-cert",
    required=False,
    type=str,
    envvar="YR_CERT_FILE",
    help="Client certificate file path",
)
@click.option(
    "--client-key",
    required=False,
    type=str,
    envvar="YR_PRIVATE_KEY_FILE",
    help="Client private key file path",
)
@click.option(
    "--ca-cert",
    required=False,
    type=str,
    envvar="YR_VERIFY_FILE",
    help="CA certificate file path",
)
@click.option(
    "--server-name",
    required=False,
    type=str,
    envvar="YR_SERVER_NAME",
    help="Server name for certificate verification (SNI)",
)
@click.option(
    "--insecure",
    is_flag=True,
    default=False,
    envvar="YR_INSECURE",
    help="Skip server TLS certificate verification (connects via HTTPS without cert check)",
)
@click.option(
    "--client-auth-type",
    required=False,
    type=click.Choice(["mutual", "one-way"], case_sensitive=False),
    default="mutual",
    envvar="YR_CLIENT_AUTH_TYPE",
    help="TLS client authentication type: 'mutual' for mTLS (default), 'one-way' for server-only verification",
)
@click.option(
    "--jwt-token",
    required=False,
    type=str,
    envvar="YR_JWT_TOKEN",
    help="JWT token for API authentication (sent in X-Auth header)",
)
@click.option("--log-level", required=False, type=str, default="INFO")
@click.option("--user", required=False, type=str, default="default")
@click.version_option(package_name="openyuanrong-sdk")
def cli(**kwargs):
    """
    run command
    """
    opts = CliGlobalOptions(**kwargs)
    if opts.server_address:
        global __server_address
        __server_address = opts.server_address
    if opts.ds_address:
        global __ds_address
        __ds_address = opts.ds_address
    if opts.client_cert:
        global __client_cert
        __client_cert = opts.client_cert
    if opts.client_key:
        global __client_key
        __client_key = opts.client_key
    if opts.ca_cert:
        global __ca_cert
        __ca_cert = opts.ca_cert
    if opts.insecure:
        global __insecure
        __insecure = opts.insecure
    if opts.server_name:
        global __server_name
        __server_name = opts.server_name
    if opts.client_auth_type:
        global __client_auth_type
        __client_auth_type = opts.client_auth_type.lower()
    if opts.jwt_token:
        global __jwt_token
        __jwt_token = opts.jwt_token
    if opts.user:
        global __user
        __user = opts.user
    logging.basicConfig(level=getattr(logging, opts.log_level.upper(), None))


@cli.command("help")
@click.argument("command_name", required=False)
@click.pass_context
def show_help(ctx, command_name):
    """Show help for commands

    Examples:
        yrcli help               # Show general help
        yrcli help deploy        # Show help for deploy command
        yrcli help token-auth    # Show help for token-auth command
    """
    if command_name is None:
        # Show general help
        click.echo(ctx.parent.get_help())  # Using built-in help() is fine here
    else:
        # Show help for specific command
        cmd = cli.commands.get(command_name)
        if cmd is None:
            click.echo(f"Error: No such command '{command_name}'.")
            click.echo("\nAvailable commands:")
            for name in sorted(cli.commands.keys()):
                click.echo(f"  {name}")
            raise click.ClickException("")
        else:
            click.echo(cmd.get_help(ctx))


@cli.command()
@click.option("--backend", required=False, type=str, default="ds")
@click.option("--code-path", required=False, type=str, default=".")
@click.option("--format", "package_format", required=False, type=str, default="zip")
@click.option("--function-json", required=False, type=str, default=None)
@click.option("--skip-package", required=False, type=bool, default=False)
@click.option("--update", required=False, is_flag=True, default=False)
@click.option(
    "-r",
    "--requirements",
    required=False,
    type=str,
    default=None,
    help="Path to requirements.txt file for installing dependencies",
)
def deploy(**kwargs):
    opts = DeployOptions(**kwargs)
    backend = opts.backend
    code_path = opts.code_path
    package_format = opts.package_format
    function_json = opts.function_json
    skip_package = opts.skip_package
    update = opts.update
    requirements = opts.requirements
    if function_json:
        with open(function_json, "r") as f:
            function_json = json.load(f)
    # Install dependencies if requirements file is provided
    if requirements and not skip_package:
        real_code_path = os.path.realpath(code_path)
        if not install_requirements(requirements, real_code_path):
            logging.error("Failed to install dependencies. Deployment aborted.")
            raise click.ClickException("")
    if not skip_package:
        real_code_path, package_key = package(backend, code_path, package_format)
        if function_json:
            function_json["storageType"] = "working_dir"
            function_json["codePath"] = package_key
        else:
            logging.info(
                f"""already upload {real_code_path} to {backend}.
export YR_WORKING_DIR={package_key} to set this package.
yrcli clear {package_key} to delete this package.
yrcli download {package_key} to download this package."""
            )
            return
    if function_json:
        name = function_json.get("name")
        if name is None:
            logging.error("function name is required to deploy function.")
            raise click.ClickException("")
        version = "latest" if function_json.get("kind", "faas") == "faas" else "$latest"
        name = FunctionName(name, version)
        query_ret, function_info = query_function(name, __user)
        if query_ret and not update:
            logging.warning(f"function {name} already exists, use --update to update it.")
            raise click.ClickException("")
        if query_ret:
            function_json["revisionId"] = function_info.get("revisionId")
            ret = update_function(function_json, __user)
            if ret[0]:
                name = get_name_from_info(ret[1])
                logging.info(f"succeed to update function: {name}")
            else:
                logging.error(f"failed to update function: {ret[1]['error']}")
        else:
            ret = deploy_function(function_json, __user)
            if ret[0]:
                name = get_name_from_info(ret[1])
                logging.info(f"succeed to deploy function: {name}")
            else:
                logging.error(f"failed to deploy function: {ret[1]}")
    else:
        logging.error("function json is required to deploy function.")


@cli.command()
@click.option("-f", "--function-name", required=False, type=str, default=None)
@click.option("-v", "--version", required=False, type=str, default=None)
def publish(function_name, version):
    if ":" not in function_name and version is None:
        raise click.ClickException("version is required if function name has no version.")
    if ":" in function_name and version is not None:
        raise click.ClickException("version should not be specified if function name has version.")
    if ":" in function_name:
        version = function_name.split(":")[1]
    function_name = FunctionName(function_name, version)
    publish_json = {}
    query_ret, function_info = query_function(function_name, __user)
    if not query_ret:
        raise click.ClickException(f"function not found: {function_name}")
    publish_json["revisionId"] = function_info.get("revisionId")
    publish_json["kind"] = function_info.get("kind", "faas")
    if version:
        publish_json["versionNumber"] = version
    ret = publish_function(function_name, publish_json, __user)
    if ret[0]:
        logging.info(f"succeed to publish function: {ret[1]}")
    else:
        logging.error(f"failed to publish function: {ret[1]}")


@cli.command()
@click.option(
    "-f",
    "--function-name",
    required=False,
    type=str,
    default=None,
    help="Function name to query",
)
@click.option(
    "-i",
    "--instance-id",
    required=False,
    type=str,
    default=None,
    help="Instance ID to query",
)
def query(function_name, instance_id):
    """Query function or instance details

    Examples:
        yrcli query -f myservice@myfunction:latest       # Query function
        yrcli query -i db6126e0-0000-4000-8000-00faf8d1692b  # Query instance
    """
    if function_name and instance_id:
        raise click.ClickException("Error: Cannot specify both function-name and instance-id")

    if not function_name and not instance_id:
        raise click.ClickException("Error: Must specify either --function-name or --instance-id")

    if function_name:
        # Query function
        function_name = FunctionName(function_name)
        ret, resp = query_function(function_name, __user)
        if ret:
            print(json.dumps(resp, indent=2, ensure_ascii=False))
        else:
            logging.warning(f"function not found: {function_name}")

    if instance_id:
        # Query instance
        ret, resp = query_instance(instance_id, __user)
        if ret:
            print(json.dumps(resp, indent=2, ensure_ascii=False))
        else:
            logging.warning(f"instance not found: {instance_id}")


@cli.command("list")
@click.argument(
    "resource_type",
    type=click.Choice(
        [
            "function",
            "functions",
            "func",
            "fun",
            "instance",
            "instances",
            "inst",
            "ins",
        ],
        case_sensitive=False,
    ),
    default="function",
    required=False,
)
def list_resources(resource_type):
    """List functions or instances

    Examples:
        yrcli list                  # List functions (default)
        yrcli list function         # List functions
        yrcli list instance         # List instances
    """
    if resource_type in ("instance", "instances", "inst", "ins"):
        # List instances
        ret, resp = query_instances(__user)
        if ret and len(resp) > 0:
            for instance in resp:
                instance_id = instance.get("id", "N/A")
                print(instance_id)
        else:
            logging.info(f"user {__user} has no instance.")
    elif resource_type in ("function", "functions", "func", "fun"):
        # List functions (default)
        ret, resp = query_function(None, __user)
        if ret and len(resp) > 0:
            for function in resp:
                print(f"{function['name'][2:]}:{function['versionNumber']}")
        else:
            logging.info(f"user {__user} has no function.")


@cli.group("sandbox")
def sandbox():
    """Manage detached sandbox instances.

    Examples:
        yrcli sandbox list
        yrcli sandbox create --namespace aaa --name bbb
        yrcli sandbox query aaa-bbb
        yrcli sandbox delete aaa-bbb
    """


@sandbox.command("create")
@click.option(
    "--namespace",
    required=True,
    type=str,
    help="Namespace for sandbox instance",
)
@click.option(
    "--name",
    required=True,
    type=str,
    help="Name for sandbox instance",
)
def sandbox_create(namespace, name):
    """Create a detached sandbox instance directly in YR runtime context."""
    os.environ.pop("YR_WORKING_DIR", None)
    try:
        with YRContext(__server_address, __ds_address, __user):
            opt = yr.InvokeOptions()
            opt.custom_extensions["lifecycle"] = "detached"
            opt.idle_timeout = 60 * 60 * 24 * 1
            opt.cpu = 1000
            opt.memory = 2048
            opt.name = name
            opt.namespace = namespace

            sandbox_inst = yr.sandbox.SandBoxInstance.options(opt).invoke()
            instance_name = yr.get(sandbox_inst.get_name.invoke())
            if not instance_name:
                instance_name = f"{namespace}-{name}"
            logging.info(f"sandbox created, instance_name={instance_name}")
    except Exception as e:
        raise click.ClickException(f"sandbox create failed, name={name}, namespace={namespace}, error={e}")


@sandbox.command("list")
@click.option(
    "--namespace",
    required=False,
    type=str,
    default=None,
    help="Filter by namespace prefix",
)
def sandbox_list(namespace):
    """List sandbox instances."""
    ret, resp = query_instances(__user)
    if not ret:
        raise click.ClickException(f"failed to list instances: {resp.get('error', resp)}")

    sandbox_ids = []
    for instance in resp:
        instance_id = instance.get("id", "")
        if not instance_id:
            continue
        if instance_id.startswith("app-"):
            continue
        if "-" not in instance_id:
            continue
        if namespace and not instance_id.startswith(f"{namespace}-"):
            continue
        sandbox_ids.append(instance_id)

    if not sandbox_ids:
        logging.info("no sandbox instance found")
        return

    for sandbox_id in sandbox_ids:
        print(sandbox_id)


@sandbox.command("query")
@click.argument("sandbox_id", type=str)
def sandbox_query(sandbox_id):
    """Query sandbox instance detail by instance id."""
    if not __server_address:
        raise click.ClickException("Error: server address is required. Use --server-address or set YR_SERVER_ADDRESS.")

    ret, resp = query_instance(sandbox_id, __user)
    if ret:
        print(json.dumps(resp, indent=2, ensure_ascii=False))
    else:
        raise click.ClickException(f"sandbox not found: {sandbox_id}")


@sandbox.command("delete")
@click.argument("sandbox_id", type=str)
def sandbox_delete(sandbox_id):
    """Delete (terminate) a sandbox instance by instance id."""
    try:
        with YRContext(__server_address, __ds_address, __user):
            yr.kill_instance(sandbox_id)
        logging.info(f"succeed to delete sandbox: {sandbox_id}")
    except Exception as e:
        raise click.ClickException(f"failed to delete sandbox {sandbox_id}: {e}")


@cli.command()
@click.option("-f", "--function-name", required=True, type=str, default=None)
@click.option("--no-clear-package", is_flag=True, default=False)
@click.option("-v", "--version", required=False, type=str, default="latest")
def delete(function_name, no_clear_package, version):
    function_name = FunctionName(function_name, version)
    if not no_clear_package:
        ret, function_info = query_function(function_name, __user)
        if not ret:
            logging.warning(f"function not found.")
            return
        code_path = function_info.get("codePath")
        if code_path and code_path.startswith("ds://"):
            key = code_path.strip("ds://").split(".")[0]
            with YRContext(__server_address, __ds_address):
                yr.kv_del(key)
            logging.info(f"succeed to del package {code_path}")
    ret, _ = delete_function(function_name, __user)
    if not ret:
        logging.warning(f"function not found.")
    else:
        logging.info(f"succeed to delete function: {function_name}")


@cli.command
@click.argument("package_uri", metavar="PACKAGE", type=str)
def clear(package_uri):
    if package_uri.startswith("ds://"):
        key = package_uri.strip("ds://").split(".")[0]
        with YRContext(__server_address, __ds_address):
            yr.kv_del(key)
        logging.info(f"succeed to del {package_uri}")


@cli.command
@click.argument("package_uri", metavar="PACKAGE", type=str)
def download(package_uri):
    if package_uri.startswith("ds://"):
        key = package_uri.strip("ds://").split(".")[0]
        file_name = package_uri.strip("ds://")
        with YRContext(__server_address, __ds_address):
            with open(file_name, "wb") as f:
                value = yr.kv_get(key)
                f.write(value)
        logging.info(f"save {package_uri} to {file_name}")


@cli.command
@click.option("-f", "--function-name", required=True, type=str, default=None)
@click.option("--payload", required=False, type=str, default=None)
@click.option("--timeout", required=False, type=int, default=30)
@click.option("--header", required=False, type=str, multiple=True)
def invoke(function_name, payload, timeout, header):
    headers = {}
    for hdr in header:
        if ":" in hdr:
            key, value = hdr.split(":", 1)
            headers[key.strip()] = value.strip()
    function_name = FunctionName(function_name)
    if payload:
        payload_dict = json.loads(payload)
    else:
        payload_dict = {}
    ret, resp = invoke_function(
        function_name, payload_dict, headers=headers, user=__user, timeout=timeout
    )
    if ret:
        print(json.dumps(resp, indent=2, ensure_ascii=False))
    else:
        logging.error(f"failed to invoke function: {resp['error']}")


@cli.command(
    "deploy-language-rt",
    context_settings=dict(
        ignore_unknown_options=True,
        allow_extra_args=True,
        allow_interspersed_args=False,
    ),
)
@click.option(
    "--runtime",
    required=False,
    type=click.Choice(["python3.11", "python3.9", "python3.10", "python3.12", "python3.13"]),
    help="Runtime language version",
)
@click.option(
    "--sdk",
    is_flag=True,
    default=False,
    help="sdk language",
)
@click.option("--no-rootfs", is_flag=True, default=False, help="Deploy without rootfs")
@click.option(
    "--function-json",
    required=False,
    type=str,
    default=None,
    help="Path to function JSON file",
)
@click.pass_context
def deploy_language_rt(ctx, runtime, sdk, function_json, no_rootfs):
    """Deploy language runtime executor function

    You can override any JSON field using dot notation, for example:
    yrcli deploy-language-rt --runtime python3.11 --cpu=1000 --memory=1024 --rootfs.storageInfo.accessKey=mykey
    yrcli deploy-language-rt --runtime python3.11 --rootfs.storageInfo.object=rootfs_python3.11.img

    Naming rules:
    - sdk + python3.x -> 0-defaultservice-py3xx (e.g. python3.11 -> py311, python3.9 -> py39)
    - non-sdk -> 0-system-faasExecutorPython3.x

    Hook handler defaults:
    - sdk -> yrlib_handler.*
    - non-sdk -> faas_executor.*
    Or provide a function JSON file directly with --function-json
    """

    # Get extra arguments from context
    overrides = ctx.args

    # Determine which user to use
    current_user = "default"

    # If function_json is provided, load it as base configuration
    if function_json:
        with open(function_json, "r") as f:
            function_json_data = json.load(f)
    else:
        # Validate runtime is provided when not using function_json
        if not runtime:
            raise click.ClickException("Error: --runtime is required when not using --function-json")

        # Generate function name based on runtime and sdk flag
        # sdk + python3.x -> 0-defaultservice-py3xx (e.g. python3.11 -> py311, python3.9 -> py39)
        if sdk and runtime.startswith("python3."):
            runtime_suffix = f"py{runtime.replace('python', '').replace('.', '')}"
            function_name = f"0-defaultservice-{runtime_suffix}"
        else:
            # python3.11 -> 0-system-faasExecutorPython3.11
            runtime_name = runtime[0].upper() + runtime[1:]  # Capitalize first letter
            function_name = f"0-system-faasExecutor{runtime_name}"

        # Build default function configuration based on the template
        if sdk:
            hook_handler = {
                "call": "yrlib_handler.call",
                "checkpoint": "yrlib_handler.checkpoint",
                "init": "yrlib_handler.init",
                "recover": "yrlib_handler.recover",
                "shutdown": "yrlib_handler.shutdown",
                "signal": "yrlib_handler.signal",
            }
        else:
            hook_handler = {
                "call": "faas_executor.faasCallHandler",
                "checkpoint": "faas_executor.faasCheckPointHandler",
                "init": "faas_executor.faasInitHandler",
                "recover": "faas_executor.faasRecoverHandler",
                "shutdown": "faas_executor.faasShutDownHandler",
                "signal": "faas_executor.faasSignalHandler",
            }
        function_json_data = {
            "name": function_name,
            "runtime": runtime,
            "kind": "yrlib",
            "cpu": 600,
            "memory": 512,
            "timeout": 600,
            "storageType": "local",
            "codePath": "/var/task",
            "hookHandler": hook_handler,
            "warmup": "seed",
            "rootfs": {
                "runtime": "runsc",
                "type": "s3",
                "imageurl": f"registry.cn-hangzhou.com",
                "readonly": True,
                "mountpoint": "/var/task/code",
                "storageInfo": {
                    "endpoint": "cn-hangzhou.alipay.aliyun-inc.com",
                    "bucket": "crfs-dev",
                    "object": f"rootfs.img",
                },
            },
        }

    # Apply overrides from command line arguments using dot notation (see docstring for examples).
    for override in overrides:
        if override.startswith("--"):
            override = override[2:]  # Remove leading --

        if "=" in override:
            key_path, value = override.split("=", 1)
            keys = key_path.split(".")

            # Navigate to the nested dict and set the value
            current = function_json_data
            for key in keys[:-1]:
                if key not in current:
                    current[key] = {}
                current = current[key]

            # Try to parse value as int, bool, or keep as string
            final_key = keys[-1]
            if value.lower() in ("true", "false"):
                current[final_key] = value.lower() == "true"
            elif value.isdigit():
                current[final_key] = int(value)
            else:
                current[final_key] = value

    if no_rootfs:
        if "rootfs" in function_json_data:
            del function_json_data["rootfs"]
        if "warmup" in function_json_data:
            del function_json_data["warmup"]

    # Get function name from the json data
    function_name = function_json_data.get("name")
    if not function_name:
        raise click.ClickException("Error: function name is required in configuration")

    # Check if function already exists
    version = "$latest"
    full_name = f"{function_name}:{version}"
    query_ret, function_info = query_function(full_name, current_user)

    if query_ret:
        # Update existing function
        function_json_data["revisionId"] = function_info.get("revisionId")
        ret = update_function(function_json_data, current_user)
        if ret[0]:
            name = get_name_from_info(ret[1])
            logging.info(f"Successfully updated FaaS language runtime function: {name}")
            if runtime:
                logging.info(f"Runtime: {runtime}")
        else:
            logging.error(f"Failed to update FaaS language runtime function: {ret[1]['error']}")
            raise click.ClickException("")
    else:
        # Deploy new function
        ret = deploy_function(function_json_data, current_user)
        if ret[0]:
            name = get_name_from_info(ret[1])
            logging.info(f"Successfully deployed FaaS language runtime function: {name}")
            if runtime:
                logging.info(f"Runtime: {runtime}")
        else:
            logging.error(f"Failed to deploy FaaS language runtime function: {ret[1]}")
            raise click.ClickException("")


@cli.command("run-spark")
@click.option(
    "--script",
    required=True,
    type=str,
    help="Path to the Python script to run with Spark",
)
@click.option(
    "--args",
    required=False,
    type=str,
    default="",
    help="Arguments to pass to the Spark job",
)
def run_spark(script, args):
    """Run a Python script with Spark job

    Example:
        yrcli run-spark --script /path/to/script.py
        yrcli run-spark --script /path/to/script.py --args "arg1 arg2 arg3"
    """
    # Validate script path
    if not os.path.exists(script):
        raise click.ClickException(f"Error: Script file not found: {script}")

    # Get absolute path
    script_path = os.path.abspath(script)
    # Build Java command
    cmd = [
        "java",
        "-cp",
        "/home/yuanrong/sparkjob/spark-job.jar",
        f"-Dentry.point.path={script_path}",
    ]

    if args:
        cmd.append(f"-Dargs={args}")

    cmd.append("com.SparkJobExample")

    logging.info("Running Spark job with script: %s", script_path)
    logging.debug("Command: %s", " ".join(cmd))

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=False)
        if result.stderr:
            logging.debug("Spark job stderr: %s", result.stderr)
        if result.returncode != 0:
            raise click.ClickException(
                f"\nSpark job failed with exit code: {result.returncode}"
            )
    except FileNotFoundError as e:
        raise click.ClickException(
            "Error: Java command not found. Make sure Java is installed and in PATH."
        ) from e
    except click.ClickException:
        raise
    except Exception as e:
        raise click.ClickException(f"Error executing Spark job: {str(e)}") from e


@cli.command("exec")
@click.option(
    "-i",
    "--stdin",
    required=False,
    type=bool,
    is_flag=True,
    default=False,
    help="Whether to allocate stdin for the instance",
)
@click.option(
    "-t",
    "--tty",
    required=False,
    type=bool,
    is_flag=True,
    default=False,
    help="Whether to allocate a TTY for the instance",
)
@click.option(
    "--verify-server",
    required=False,
    type=bool,
    is_flag=True,
    default=True,
    help="Verify server certificate (default: True)",
)
@click.argument("instance", type=str)
@click.argument("command", type=str)
def exec_command(stdin, tty, verify_server, instance, command):
    use_ssl = __client_cert is not None and __client_key is not None
    try:
        host, port = __server_address.split(":")
        asyncio.run(
            run_client(
                TerminalClientParams(
                    host=host,
                    port=port,
                    instance=instance,
                    command=command,
                    allocate_tty=tty,
                    stdin=stdin,
                    user=__user,
                    use_ssl=use_ssl,
                    cert_file=__client_cert,
                    key_file=__client_key,
                    ca_file=__ca_cert,
                    verify_server=verify_server,
                    token=__jwt_token,
                )
            )
        )
    except KeyboardInterrupt:
        logging.warning("\nDisconnected")


@cli.command(
    "runtime_main",
    context_settings=dict(ignore_unknown_options=True, allow_extra_args=True),
)
def runtime_main():
    """Start the runtime main process

    All unknown options will be ignored, allowing the command to run without errors.
    """
    from yr.main.yr_runtime_main import main as yr_runtime_main

    yr_runtime_main()


@cli.command("token-auth")
@click.option("--token", required=True, type=str, help="JWT token to authenticate")
@click.option(
    "--iam-address",
    required=True,
    type=str,
    envvar="YR_IAM_ADDRESS",
    help="YuanRong IAM Server address",
)
def token_auth(token, iam_address):
    """Authenticate/verify a JWT token

    Example:
        yrcli token-auth --iam-address 127.0.0.1:31112 --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    """
    http_client = HTTPClient(timeout=30)
    url = f"http://{iam_address}/iam-server/v1/token/auth"
    headers = {"X-Auth": token}
    resp = http_client.request(url, {}, headers=headers, method="GET")

    if resp["success"]:
        logging.info("Token is valid. Authentication successful.")
    else:
        raise click.ClickException(f"Token authentication failed: {resp.get('error', 'Unknown error')}")


@cli.command("token-require")
@click.option(
    "--tenant-id", required=False, type=str, help="Tenant ID for token generation"
)
@click.option("--ttl", required=False, type=int, help="Token time-to-live in seconds")
@click.option("--role", required=False, type=str, help="Role for the token")
@click.option(
    "--iam-address",
    required=True,
    type=str,
    envvar="YR_IAM_ADDRESS",
    help="YuanRong IAM Server address",
)
def token_require(tenant_id, ttl, role, iam_address):
    """Request/generate a new JWT token

    Example:
        yrcli token-require --iam-address 127.0.0.1:31112 --tenant-id tenant_789 --role viewer
        yrcli token-require --iam-address 127.0.0.1:31112 --tenant-id user --ttl 3600 --role admin
    """
    http_client = HTTPClient(timeout=30)
    url = f"http://{iam_address}/iam-server/v1/token/require"
    headers = {}
    if tenant_id:
        headers["X-Tenant-ID"] = tenant_id
    if ttl:
        headers["X-TTL"] = str(ttl)
    if role:
        headers["X-Role"] = role

    resp = http_client.request(url, {}, headers=headers, method="GET")

    if resp["success"]:
        # Print token separately for easy copy
        if "X-Auth" in resp["headers"]:
            print(f"Token: {resp['headers']['X-Auth']}")
    else:
        raise click.ClickException(f"Token generation failed: {resp.get('error', 'Unknown error')}")


@cli.command("token-abandon")
@click.option("--token", required=True, type=str, help="JWT token to abandon/revoke")
@click.option("--tenant-id", required=False, type=str, help="Tenant ID")
@click.option(
    "--iam-address",
    required=True,
    type=str,
    envvar="YR_IAM_ADDRESS",
    help="YuanRong IAM Server address",
)
def token_abandon(token, tenant_id, iam_address):
    """Abandon/revoke a JWT token

    Example:
        yrcli token-abandon --iam-address 127.0.0.1:31112 \\
            --token "<jwt>"
        yrcli token-abandon --iam-address 127.0.0.1:31112 \\
            --token "<jwt>" --tenant-id user
    """
    http_client = HTTPClient(timeout=30)
    url = f"http://{iam_address}/iam-server/v1/token/abandon"
    headers = {"X-Auth": token}
    if tenant_id:
        headers["X-Tenant-ID"] = tenant_id

    resp = http_client.request(url, {}, headers=headers, method="POST")

    if resp["success"]:
        logging.info("Token successfully abandoned/revoked")
    else:
        raise click.ClickException(f"Token abandonment failed: {resp.get('error', 'Unknown error')}")


def main():
    return cli()


if __name__ == "__main__":
    sys.exit(main())

