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

"""Faas executor, an adapter between posix and faas"""
import json
import logging
import os
import time
import traceback
import queue

from typing import Any, List
from yr.code_manager import CodeManager
from yr.common.utils import to_json_string
from yr.err_type import ErrorInfo, ModuleCode, ErrorCode
from yr.functionsdk.utils import encode_base64, timeout

from yr import log
from yr.common import constants
from yr.common.constants import META_PREFIX, METALEN
from yr.functionsdk.context import init_context, init_context_invoke, load_context_meta
from yr.functionsdk.logger_manager import UserLogManager
from yr.functionsdk.error_code import FaasErrorCode

_STAGE_INIT = "init"
_STAGE_INVOKE = "invoke"
_INDEX_META_DATA = 0
_INDEX_CALL_USER_EVENT = 1
_RUNTIME_MAX_RESP_BODY_SIZE = 6 * 1024 * 1024
_SHUTDOWN_CHECK_INTERVAL = 0.1
_EVENT_HEADER = "Accept"
_EVENT_HEADER_VALUE = "text/event-stream"
_EVENT_EOF = "yuanrong_event_EOF"
requestQueue = queue.Queue(maxsize=1000)

_logger = logging.getLogger(__name__)


def faas_init_handler(posix_args: List[Any]) -> str:
    """faas init handler"""
    _logger.debug("Started to call FaaS init handler.")
    try:
        context_meta = parse_faas_param(posix_args[_INDEX_META_DATA])
        load_context_meta(context_meta)
    except TypeError as e:
        err_msg = f"faas init request args undefined: {repr(e)}, traceback: {traceback.format_exc()}"
        _logger.error(err_msg)
        raise RuntimeError(transform_init_response_to_str(err_msg, FaasErrorCode.INIT_FUNCTION_FAIL)) from e
    except json.decoder.JSONDecodeError as e:
        err_msg = f"faas init request args json decode error: {repr(e)}, traceback: {traceback.format_exc()}"
        _logger.error(err_msg)
        raise RuntimeError(transform_init_response_to_str(err_msg, FaasErrorCode.INIT_FUNCTION_FAIL)) from e
    except Exception as e:
        err_msg = f"faaS failed to load context and user logger, err: {repr(e)}, traceback: {traceback.format_exc()}"
        _logger.error(err_msg)
        raise RuntimeError(transform_init_response_to_str(err_msg, FaasErrorCode.INIT_FUNCTION_FAIL)) from e
    code_path = CodeManager().get_code_path(constants.KEY_USER_INIT_ENTRY)
    if code_path == "":
        return transform_init_response_to_str("success", FaasErrorCode.NONE_ERROR)
    user_init_code = CodeManager().load(constants.KEY_USER_INIT_ENTRY)
    if user_init_code is None:
        raise RuntimeError(
            transform_init_response_to_str(
                f"failed to find init handler: {code_path}", FaasErrorCode.INIT_FUNCTION_FAIL))
    # Load and run user init code
    error_code = FaasErrorCode.NONE_ERROR

    @timeout(int(os.getenv('RUNTIME_INITIALIZER_TIMEOUT')))
    def _init_with_timeout(_code, _context):
        _code(_context)

    try:
        context = init_context(_STAGE_INIT)
        user_init_code(context)
    except Exception as err:
        err_msg = f"Fail to run user init handler. err: {repr(err)}. traceback: {traceback.format_exc()}"
        error_code = FaasErrorCode.INIT_FUNCTION_FAIL
        _logger.exception(err_msg)
    finally:
        UserLogManager().shutdown()
    if error_code != FaasErrorCode.NONE_ERROR:
        raise RuntimeError(transform_init_response_to_str(err_msg, error_code))
    _logger.info("Succeeded to call FaaS user init handler: [%s]", context_meta['funcMetaData']['handler'])
    return transform_init_response_to_str("success", FaasErrorCode.NONE_ERROR)


def faas_call_handler(posix_args: List[Any]) -> str:
    """faas call handler"""
    _logger.info("Faas call handler called.")
    user_code = CodeManager().load(constants.KEY_USER_CALL_ENTRY)
    error_code = FaasErrorCode.NONE_ERROR
    if user_code is None:
        err_msg = "faas executor find empty user call code"
        _logger.error(err_msg)
        error_code = FaasErrorCode.INIT_FUNCTION_FAIL
        return transform_call_response_to_str(err_msg, error_code)
    event = parse_faas_param(posix_args[_INDEX_CALL_USER_EVENT])
    trace_id = get_trace_id_from_params(posix_args[_INDEX_META_DATA])
    header = {}
    if isinstance(event, dict):
        header = event.get("header", {})
        if not isinstance(header, dict):
            err_msg = f'header type is not dict'
            error_code = FaasErrorCode.ENTRY_EXCEPTION
            return transform_call_response_to_str(err_msg, error_code)
        event = event.get('body', {})
        if isinstance(event, str):
            try:
                event = json.loads(event)

            except ValueError as err:
                err_msg = f'failed to loads event body err: {err}'
                error_code = FaasErrorCode.ENTRY_EXCEPTION
                _logger.error(err_msg)
                return transform_call_response_to_str(err_msg, error_code)
        if event is None:
            event = {}

    is_event_enable = isinstance(header, dict) and header.get(_EVENT_HEADER) == _EVENT_HEADER_VALUE
    context = init_context_invoke(_STAGE_INVOKE, header)
    if len(context.get_trace_id()) == 0:
        context.set_trace_id(trace_id)

    @timeout(int(os.getenv('RUNTIME_TIMEOUT')))
    def _invoke_with_timeout(_code, _event, _context):
        return _code(_event, _context)

    try:
        requestQueue.put(1)
        result = user_code(event, context)
    except SystemExit as exit_error:
        _logger.exception("Fail to run user call handler. err: %s. traceback: %s",
                                   exit_error, traceback.format_exc())
        result = f"Fail to run user call handler. err: user code sys.exit()."
        error_code = FaasErrorCode.ENTRY_EXCEPTION
    except Exception as err:
        _logger.exception("Fail to run user call handler. err: %s. traceback: %s",
                                   err, traceback.format_exc())
        result = f"Fail to run user call handler. err: {err}."
        error_code = FaasErrorCode.ENTRY_EXCEPTION
    finally:
        if not requestQueue.empty():
            requestQueue.get()

    # SSE: always send EOF after user handler finished (success or failure).
    if is_event_enable:
        try:
            from yr.fnruntime import stream_write

            stream_write(_EVENT_EOF, context.get_invoke_id(), context.get_instance_id())
        except Exception as err:
            log.get_logger().exception("failed to send event EOF, err: %s", err)
            error_code = FaasErrorCode.FUNCTION_INVOCATION_EXCEPTION
            result = f"failed to send event EOF, err: {err}"

    try:
        result_str = transform_call_response_to_str(result, error_code)
    except Exception as err:
        # Can be RecursionError, RuntimeError, UnicodeError, MemoryError, etc...
        err_msg = f"Fail to stringify user call result. " \
                  f"err: {err}. traceback: {traceback.format_exc()}"
        _logger.exception(err_msg)
        raise RuntimeError(err_msg) from err
    finally:
        UserLogManager().shutdown()
    _logger.info("Succeeded to call FaaS user call handler: [%s]", os.environ.get(
        "RUNTIME_HANDLER", "handler name Not found in environment"))
    return result_str


def faas_shutdown_handler(grace_period_second) -> ErrorInfo:
    """faas shutdown handler"""
    _logger.info("start shutdown user function")
    user_code = CodeManager().load(constants.KEY_USER_SHUT_DOWN_ENTRY)
    error_info = ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME_KILL, "user function shutdown ok")
    if user_code is None:
        err_msg = "can not find shutdown entry."
        _logger.warning(err_msg)
    else:
        @timeout(int(os.getenv("PRE_STOP_TIMEOUT")))
        def _invoke_with_timeout(_code):
            return _code()

        exit_loop = True
        while exit_loop:
            if requestQueue.empty():
                exit_loop = False
                try:
                    _logger.info("start exec user shutdown code")
                    _invoke_with_timeout(user_code)
                except TimeoutError as err:
                    err_msg = f"Fail to run user shutdown  handler. err: {err}. traceback: {traceback.format_exc()}"
                    _logger.exception(err_msg)
                    error_info = ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME, err_msg)
                except BaseException as err:
                    err_msg = f"Fail to run  user shutdown handler. err: {err}. traceback: {traceback.format_exc()}"
                    _logger.exception(err_msg)
                    error_info = ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME, err_msg)
            else:
                time.sleep(_SHUTDOWN_CHECK_INTERVAL)
    return error_info


# Helpers
def transform_call_response_to_str(response, status_code: FaasErrorCode):
    """Method transform_call_response_to_str"""
    key_for_body = "body"
    result = {}
    if response is None:
        result[key_for_body] = ""
    else:
        try:
            json.dumps(response)
        except TypeError as err:
            _logger.exception("result is not JSON serializable, err: %s", err)
            result[key_for_body] = f"failed to convert the result to a JSON string, err:{err}"
            status_code = FaasErrorCode.FUNCTION_RESULT_INVALID
        else:
            result[key_for_body] = response
    result["innerCode"] = str(status_code.value)
    result["billingDuration"] = "this is billing duration TODO"
    result["logResult"] = encode_base64("this is user log TODO".encode('utf-8'))
    result["invokerSummary"] = "this is summary TODO"

    resp_json = to_json_string(result)
    if len(resp_json.encode()) > _RUNTIME_MAX_RESP_BODY_SIZE:
        result[key_for_body] = f"response body size {len(resp_json.encode())} exceeds the limit of 6291456"
        result["innerCode"] = str(FaasErrorCode.RESPONSE_EXCEED_LIMIT.value)
        resp_json = to_json_string(result)
    return make_faas_result(resp_json)


def transform_init_response_to_str(response, status_code: FaasErrorCode):
    """Method transform_call_response_to_str"""
    result = dict(
        message="" if response is None else response,
        errorCode=str(status_code.value)
    )
    if status_code != FaasErrorCode.NONE_ERROR:
        return to_json_string(result)
    return make_faas_result(to_json_string(result))


def parse_faas_param(arg):
    """parse param of faas"""
    arg_str = arg.to_pybytes()
    if len(arg_str) > METALEN:
        return json.loads(arg_str[METALEN:])
    return json.loads(arg_str)


def get_trace_id_from_params(arg):
    """get trace id from params"""
    arg_str = arg.to_pybytes()
    if isinstance(arg_str, str):
        return arg_str
    return ""


def make_faas_result(result):
    """make result of faas"""
    res = META_PREFIX + result
    return res
