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

"""CLI implementation for executing code in a sandboxed environment."""

import asyncio
import os
import signal
import shutil
import ssl
import sys
import termios
import tty
from dataclasses import dataclass
from typing import Optional
from urllib.parse import quote

import websockets


@dataclass
class TerminalClientParams:
    """Parameters for WebSocket terminal client connection."""

    host: str
    port: str
    instance: Optional[str] = None
    command: Optional[str] = None
    allocate_tty: Optional[bool] = None
    stdin: Optional[bool] = None
    rows: Optional[int] = None
    cols: Optional[int] = None
    user: Optional[str] = None
    use_ssl: bool = False
    cert_file: Optional[str] = None
    key_file: Optional[str] = None
    ca_file: Optional[str] = None
    verify_server: bool = True
    token: Optional[str] = None


def create_ssl_context(
    cert_file=None,
    key_file=None,
    ca_file=None,
    verify_server=True
):
    """Create SSL context for mutual TLS authentication.
    
    Args:
        cert_file: Client certificate file path
        key_file: Client private key file path
        ca_file: CA certificate file path for server verification
        verify_server: Whether to verify server certificate
    
    Returns:
        ssl.SSLContext or None if TLS is not configured
    """
    # If no certificates are provided, return None (no TLS)
    if not cert_file and not key_file and not ca_file:
        return None
    
    try:
        # Create SSL context
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
        ssl_context.minimum_version = ssl.TLSVersion.TLSv1_2
        
        # Load client certificate and key for mutual authentication
        if cert_file and key_file:
            if not os.path.exists(cert_file):
                print(f"Warning: Client certificate file not found: {cert_file}", file=sys.stderr)
            elif not os.path.exists(key_file):
                print(f"Warning: Client key file not found: {key_file}", file=sys.stderr)
            else:
                ssl_context.load_cert_chain(cert_file, key_file)
        
        # Load CA certificate for server verification
        if ca_file:
            if not os.path.exists(ca_file):
                print(f"Warning: CA certificate file not found: {ca_file}", file=sys.stderr)
            else:
                ssl_context.load_verify_locations(ca_file)
        else:
            # Use default system CA certificates
            ssl_context.load_default_certs()
        
        # Configure server certificate verification
        if not verify_server:
            ssl_context.check_hostname = False
            ssl_context.verify_mode = ssl.CERT_NONE
            print("Warning: Server certificate verification is disabled (insecure)", file=sys.stderr)
        
        return ssl_context
    except Exception as e:
        print(f"Error creating SSL context: {e}", file=sys.stderr)
        return None


class RawTerminal:
    def __init__(self, fd):
        self.fd = fd
        self.old = None

    def __enter__(self):
        self.old = termios.tcgetattr(self.fd)
        tty.setraw(self.fd)
        return self

    def __exit__(self, exc_type, exc, tb):
        if self.old:
            termios.tcsetattr(self.fd, termios.TCSADRAIN, self.old)


async def read_stdin(ws, should_exit):
    """读取标准输入并发送到 WebSocket"""
    loop = asyncio.get_event_loop()
    reader = asyncio.StreamReader(loop=loop)
    protocol = asyncio.StreamReaderProtocol(reader, loop=loop)
    try:
        await loop.connect_read_pipe(lambda: protocol, sys.stdin)
    except Exception as e:
        print(f"Warning: failed to attach stdin pipe: {e}", file=sys.stderr)
        return

    try:
        while not should_exit.is_set():
            try:
                data = await asyncio.wait_for(reader.read(4096), timeout=0.1)
                if not data:
                    break
                # 检测 Ctrl+] (ASCII 29) 作为退出信号
                if b"\x1d" in data:
                    should_exit.set()
                    break
                await ws.send(data)
            except asyncio.TimeoutError:
                continue
    except Exception as e:
        # Log the exception but don't interrupt the exit flow
        import logging
        logging.getLogger(__name__).debug("Error in stdin reader: %s", e)
    finally:
        pass


async def read_websocket(ws, should_exit):
    """从 WebSocket 读取输出并写入标准输出"""
    try:
        async for message in ws:
            if should_exit.is_set():
                break
            if isinstance(message, bytes):
                os.write(sys.stdout.fileno(), message)
            else:
                os.write(sys.stdout.fileno(), message.encode("utf-8"))
    except Exception as e:
        # Log the error and signal exit
        import logging
        logging.getLogger(__name__).debug("Error reading from websocket: %s", e)
    finally:
        should_exit.set()


async def watch_terminal_resize(ws, should_exit):
    """监听本地终端尺寸变化并发送 resize 消息到 WebSocket"""
    if not hasattr(signal, "SIGWINCH"):
        return

    loop = asyncio.get_running_loop()
    resize_event = asyncio.Event()

    def on_resize(_signum, _frame):
        loop.call_soon_threadsafe(resize_event.set)

    try:
        old_handler = signal.getsignal(signal.SIGWINCH)
        signal.signal(signal.SIGWINCH, on_resize)
    except Exception as e:
        # If we can't set up signal handler, log and return gracefully
        import logging
        logging.getLogger(__name__).debug("Failed to set SIGWINCH handler: %s", e)
        return

    last_size = None
    try:
        while not should_exit.is_set():
            try:
                await asyncio.wait_for(resize_event.wait(), timeout=0.1)
            except asyncio.TimeoutError:
                continue

            resize_event.clear()
            try:
                terminal_size = shutil.get_terminal_size()
            except Exception as e:
                # If terminal size is unavailable, just skip this iteration
                import logging
                logging.getLogger(__name__).debug("Failed to get terminal size: %s", e)
                continue

            cols = terminal_size.columns
            rows = terminal_size.lines
            if cols <= 0 or rows <= 0:
                continue

            current_size = (cols, rows)
            if current_size == last_size:
                continue

            last_size = current_size
            try:
                await ws.send(f"RESIZE:{cols}:{rows}")
            except Exception as e:
                # If send fails, signal exit and log
                import logging
                logging.getLogger(__name__).debug("Failed to send resize: %s", e)
                should_exit.set()
                break
    finally:
        try:
            signal.signal(signal.SIGWINCH, old_handler)
        except Exception as e:
            # Log but don't raise in cleanup code
            import logging
            logging.getLogger(__name__).debug("Failed to restore SIGWINCH handler: %s", e)


async def send_terminal_resize(ws, rows=None, cols=None):
    """发送终端尺寸到 WebSocket（协议: RESIZE:cols:rows）"""
    try:
        if rows is None or cols is None:
            terminal_size = shutil.get_terminal_size()
            if rows is None:
                rows = terminal_size.lines
            if cols is None:
                cols = terminal_size.columns
    except Exception as e:
        # If terminal size is unavailable, use server defaults
        import logging
        logging.getLogger(__name__).debug("Failed to get terminal size: %s", e)
        return

    if rows is None or cols is None:
        return
    if rows <= 0 or cols <= 0:
        return
    try:
        await ws.send(f"RESIZE:{cols}:{rows}")
    except Exception as e:
        # Log failure to send resize message
        import logging
        logging.getLogger(__name__).debug("Failed to send resize message: %s", e)


async def run_client(params: TerminalClientParams):
    """Connect to runtime terminal WebSocket using ``params``."""
    host = params.host
    port = params.port
    instance = params.instance
    command = params.command
    allocate_tty = params.allocate_tty
    rows = params.rows
    cols = params.cols
    user = params.user
    use_ssl = params.use_ssl
    cert_file = params.cert_file
    key_file = params.key_file
    ca_file = params.ca_file
    verify_server = params.verify_server
    token = params.token

    # 获取当前终端的实际尺寸作为默认值
    if rows is None or cols is None:
        try:
            terminal_size = shutil.get_terminal_size()
            if rows is None:
                rows = terminal_size.lines
            if cols is None:
                cols = terminal_size.columns
        except Exception as e:
            # If unable to get terminal size, use server defaults
            import logging
            logging.getLogger(__name__).debug("Failed to get terminal size: %s", e)

    # 构建 URL query 参数
    query_params = []
    if instance:
        query_params.append(f"instance={instance}")
    if command:
        query_params.append(f"command={quote(command)}")
    if allocate_tty is not None:
        query_params.append(f"tty={str(allocate_tty).lower()}")
    if rows:
        query_params.append(f"rows={rows}")
    if cols:
        query_params.append(f"cols={cols}")
    if user:
        query_params.append(f"tenant_id={quote(user)}")
    if token:
        query_params.append(f"token={quote(token)}")

    query_string = "&".join(query_params)
    
    # Determine protocol based on SSL usage
    protocol = "wss" if use_ssl else "ws"
    uri = f"{protocol}://{host}:{port}/terminal/ws"
    if query_string:
        uri += f"?{query_string}"

    if instance:
        print(f"Connecting to {instance}...\nPress Ctrl+] to disconnect", file=sys.stderr)
    else:
        print("Connecting...\nPress Ctrl+] to disconnect", file=sys.stderr)
    
    # Create SSL context if needed
    ssl_context = None
    if use_ssl:
        ssl_context = create_ssl_context(
            cert_file=cert_file,
            key_file=key_file,
            ca_file=ca_file,
            verify_server=verify_server
        )
        if ssl_context:
            print("Using mutual TLS authentication", file=sys.stderr)
        else:
            print("Warning: SSL requested but failed to create SSL context", file=sys.stderr)

    try:
        async with websockets.connect(uri, ssl=ssl_context) as ws:
            should_exit = asyncio.Event()

            with RawTerminal(sys.stdin.fileno()):
                if allocate_tty and sys.stdin.isatty():
                    await send_terminal_resize(ws, rows=rows, cols=cols)

                # 同时处理输入和输出
                tasks = [
                    asyncio.create_task(read_stdin(ws, should_exit)),
                    asyncio.create_task(read_websocket(ws, should_exit)),
                ]
                if allocate_tty and sys.stdin.isatty():
                    tasks.append(asyncio.create_task(watch_terminal_resize(ws, should_exit)))

                await should_exit.wait()

                # 结束后取消所有后台任务
                for task in tasks:
                    if task.done():
                        continue
                    task.cancel()
                    try:
                        await task
                    except asyncio.CancelledError:
                        # Task was cancelled, this is expected behavior during shutdown
                        import logging
                        logging.getLogger(__name__).debug("Task cancelled during shutdown")
    except KeyboardInterrupt:
        print("\n[Interrupted]", file=sys.stderr)
    except Exception as e:
        print(f"\nConnection error: {e}", file=sys.stderr)
