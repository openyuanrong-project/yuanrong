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

import argparse
import json
import logging
import os
import socket
import sys
import threading
import time

logging.basicConfig(format="%(asctime)s: %(message)s", level=logging.INFO)


class DebugServer:
    def __init__(self, host="0.0.0.0", port=5555):
        self.host = host
        self.port = port
        self.runtime_debug_port = None
        self.runtime_pid = None
        self.registered_rpdbs = {}
        self.track_cache = {}

    def safe_close(self, sock):
        try:
            sock.shutdown(sock.SHUT_RDWR)
        except OSError:
            pass
        finally:
            try:
                sock.close()
            except OSError:
                pass

    def forward(self, src, dst):
        try:
            while True:
                data = src.recv(4096)
                if not data:
                    break
                dst.sendall(data)
        except Exception as e:
            logging.info(f"[DebugServer] 转发中断: {e}")
        finally:
            self.safe_close(src)
            self.safe_close(dst)

    def forward_user_to_runtime(self, user_conn, runtime_conn, *_):
        try:
            while True:
                data = user_conn.recv(4096)
                if not data:
                    break
                runtime_conn.sendall(data)
        finally:
            self.safe_close(user_conn)
            self.safe_close(runtime_conn)

    def start_stack_listener(self):
        listener = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            listener.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            listener.bind((self.host, 0))
            listener.listen(5)
            actual_stack_port = listener.getsockname()[1]
            logging.info(
                f"[DebugServer] 堆栈监听端口 {actual_stack_port} 已启动 (PID: {os.getpid()}"
            )

            while True:
                conn, addr = listener.accept()
                threading.Thread(
                    target=self.handle_stack_conn, args=(conn, addr), daemon=True
                ).start()
        except Exception as e:
            logging.error(f"[DebugServer] 堆栈监听失败: {e}")
        finally:
            listener.close()

    def start(self):
        logging.info(
            f"[DebugServer] 监听 {self.host}:{self.port}, 等待 Runtime 或 yr cli 链接..."
        )
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            s.bind((self.host, self.port))
            host, port = s.getsockname()
            os.environ["YR_DEBUG_SERVER_PORT"] = str(port)
            logging.info(f"[DebugServer] Listening on port {port}", file=sys.stderr)
            s.listen(5)

            threading.Thread(target=self.start_stack_listener, daemon=True).start()

            while True:
                conn, addr = s.accept()
                self._handle_connection(conn, addr)

    def _bridge_stack_conn(self, conn, addr):
        buffer = ""
        try:
            while True:
                data = conn.recv(4096)
                if not data:
                    break
                buffer += data.decode("utf-8", errors="ignore")
                while "\n" in buffer:
                    line, buffer = buffer.split("\n", 1)
                    line = line.strip()
                    if not line:
                        continue
                    try:
                        stack_info = json.loads(line)
                        if "stack" in stack_info and self.runtime_pid:
                            self.stack_cache[self.runtime_pid] = stack_info["stack"]
                            logging.info(
                                f"[DebugServer] 缓存堆栈 (PID {self.runtime_pid})"
                            )
                    except json.JSONDecodeError:
                        pass
        except Exception as e:
            logging.info(f"[DebugServer] 堆栈通道错误: {e}")
        finally:
            self.safe_close(conn)

    def _handle_connection(self, conn, addr):
        try:
            data = conn.recv(1024)
            if not data:
                self.safe_close(conn)
                return
            data_str = data.decode().strip()

            if self._handle_registration(conn, data_str):
                return

            logging.info(f"[DebugServer] 收到 yr cli 连接: {addr}")
            runtime_conn = self._connect_to_rpdb()

            if runtime_conn is None:
                msg = "[DebugServer] 无法连接 RPDB, yr CLI 无法调试\n"
                try:
                    conn.sendall(msg.encode())
                except Exception:
                    pass
                self.safe_close(conn)
                return
            self._bridge_stack_conn(conn, runtime_conn)
        except Exception as e:
            logging.error(f"[DebugServer] 处理连接时发生未预期错误: {e}")
            try:
                conn.close()
            except:
                pass

    def _handle_registration(self, conn, data_str):
        if "|" not in data_str:
            return False
        parts = data_str.split("|")
        if len(parts) != 4:
            return False
        instance_id, pid_str, port_str, language = parts
        self.runtime_pid = int(pid_str)
        self.runtime_debug_port = int(port_str)
        self.runtime_instance_id = instance_id
        self.runtime_language = language
        logging.info(
            f"[DebugServer] 注册 instance {instance_id} pid = {pid_str} "
            f"port = {port_str} language = {language}"
        )
        conn.sendall(b"OK\n")
        self.safe_close(conn)
        return True

    def _connect_to_rpdb(self):
        runtime_conn = None
        for _ in range(10):
            try:
                if not self.runtime_debug_port:
                    raise Exception("没有注册的 RPDB")
                runtime_conn = socket.create_connection(
                    ("127.0.0.1", self.runtime_debug_port), timeout=1
                )
                logging.info(
                    f"[DebugServer] 成功连接到 RPDB 端口 {self.runtime_debug_port}"
                )
                break
            except Exception:
                time.sleep(0.1)
        return runtime_conn


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", "-p", type=int, default=5555)
    args = parser.parse_args()
    server = DebugServer(port=args.port)
    server.start()


if __name__ == "__main__":
    main()
