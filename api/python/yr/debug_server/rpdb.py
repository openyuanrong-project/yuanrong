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

import inspect
import logging
import os
import signal
import socket
import sys
import time
from pdb import Pdb
import yr
from yr.apis import get_node_ip_address, get_node_id, get_namespace

logging.basicConfig(format="%(asctime)s: %(message)s", level=logging.INFO)

_ACTIVE_RPDB = None


class _CRLFFileWrapper:
    """"""

    def __init__(self, conn):
        self.conn = conn

    def write(self, data):
        if not data:
            return
        if isinstance(data, str):
            data = data.replace("\n", "\r\n").encode()
        try:
            self.conn.sendall(data)
        except Exception:
            pass

    def readline(self):
        buf = b""
        while True:
            ch = self.conn.recv(1)
            if not ch:
                return ""
            buf += ch
            if ch == b"\n":
                break
        try:
            return buf.decode()
        except Exception:
            return ""

    def flush(self):
        """"""
        pass


class SessionPdb(Pdb):
    def __init__(self):
        super().__init__(nosigint=True)
        self.listen_sock = None
        self.conn = None
        self.attached = False
        self._orig_stdout = None
        self._orig_stderr = None
        self.logical_info = {}

    def listen_and_register(self):
        self.listen_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.listen_sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.listen_sock.bind(("0.0.0.0", 0))
        self.listen_sock.listen(1)

        port = self.listen_sock.getsockname()[1]
        logging.info(f"Runtime: rpdb listening on port {port}")
        debug_server_port = os.environ.get("YR_DEBUG_SERVER_PORT")
        instance_id = os.environ.get("YR_RUNTIME_ID", "")
        language = os.environ.get("YR_RUNTIME_LANG", "python")

        if debug_server_port:
            try:
                with socket.create_connection(
                    ("127.0.0.1", int(debug_server_port)), timeout=5
                ) as ds:
                    msg = f"{instance_id}|{os.getpid()}|{port}|{language}\n"
                    ds.sendall(msg.encode())
                    ds.recv(1024)
            except Exception as e:
                logging.warning(f"Runtime: register DebugServer failed: {e}")

    def attach(self):
        logging.info("Runtime: waiting DebugServer connection ...")
        self.conn, addr = self.listen_sock.accept()
        logging.info(f"Runtime: DebugServer connected from {addr}")

        wrapped = _CRLFFileWrapper(self.conn)

        self._orig_stdout = sys.stdout
        self._orig_stderr = sys.stderr
        sys.stdout = wrapped
        sys.stderr = wrapped

        self.stdin = wrapped
        self.stdout = wrapped
        self.use_rawinput = False
        self.prompt = "(yrdb)"

        self.attached = True

    def cleanup(self):
        if not self.attached:
            return
        try:
            sys.stdout = self._orig_stdout
            sys.stderr = self._orig_stderr
            self.conn.close()
            self.listen_sock.close()
        except Exception:
            pass
        self.attached = False
        logging.info("Runtime: debugger detached, program continues")

    def set_trace_here(self, frame, logical_info=None):
        self.logical_info = logical_info or {}
        self.reset()
        self.curframe = frame
        self.interaction(frame, None)

    def do_bt(self, arg):
        Pdb.do_bt(self, arg)
        if self.logical_info:
            self.stdout.write("\n=== YR Logical Stack ===\n")
            for k, v in self.logical_info.items():
                self.stdout.write(f"{k}: {v}\n")
            self.stdout.write("========================\n")

    do_where = do_bt

    def do_continue(self, arg):
        return Pdb.do_continue(self, arg)

    do_c = do_cont = do_continue

    def do_quit(self, arg):
        self.cleanup()
        return Pdb.do_quit(self, arg)

    do_q = do_exit = do_quit


def set_trace():
    global _ACTIVE_RPDB
    frame = inspect.currentframe().f_back
    while frame and "yr/" in frame.f_code.co_filename:
        frame = frame.f_back
    logical_info = {}
    try:
        if yr:
            logical_info["node_ip"] = get_node_ip_address()
            logical_info["node_id"] = get_node_id()
            logical_info["namespace"] = get_namespace()
        logical_info["filename"] = frame.f_code.co_filename
        logical_info["lineno"] = frame.f_lineno
        logical_info["function"] = frame.f_code.co_name
        logical_info["timestamp"] = time.time()
    except Exception:
        pass

    if _ACTIVE_RPDB is None:
        _ACTIVE_RPDB = SessionPdb()
        _ACTIVE_RPDB.listen_and_register()
        os.kill(os.getpid(), signal.SIGSTOP)
        _ACTIVE_RPDB.attach()

    _ACTIVE_RPDB.set_trace_here(frame, logical_info)
