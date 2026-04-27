#!/usr/bin/env python3
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

import json
import os
from typing import Union
import uuid
import logging
from http import HTTPStatus

import aiohttp
import pydantic
from yr.decorator.instance_proxy import InstanceProxy
import yr
from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse, Response

from constants import (
    CONTROLLER_ACTOR_NAME,
    BALANCER_ACTOR_NAME,
    VLLM_NAMESPACE,
)
from entities import DispatchResult

logger = logging.getLogger(__name__)

app = FastAPI()


class ProxyDeployment:
    #: the balancer handle
    _balancer_handle: Union[InstanceProxy, None]

    def __init__(self):
        self._balancer_handle = None

    @staticmethod
    async def record_exception_info(e):
        """
        record exception info
        Args:
            e: exception info
        """
        import sys
        import traceback
        exc_info = sys.exc_info()
        logger.error(f"Error occurred in proxy server:\n {e}")
        logger.error("".join(traceback.format_exception(*exc_info)))

    async def forward_request(self, url: str, headers: dict, data: dict):
        """
        Send request to the inference instance, return the AsyncGenerator reading the content
        Args:
            url: request url
            headers: request header
            data: request data
        Returns:
            AsyncGenerator: the first iteration is the status code, and subsequent iterations are the response content
        """
        async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=6 * 60 * 60)) as session:
            async with session.post(url=url, json=data, headers=headers) as response:
                # Return status code in advance
                yield response.status
                if response.status == 200:
                    async for chunk_bytes in response.content.iter_chunked(1024):
                        yield chunk_bytes
                else:
                    content = await response.read()
                    yield content

    async def forward_request_without_yield(self, url: str, headers: dict, data: dict):
        """
        Asynchronously sends a POST request with JSON data and returns HTTP status code with raw content.

        This method uses aiohttp.ClientSession with a 6-hour total timeout to POST data to the specified URL.
        Headers and JSON payload are provided by the caller without validation. Response content is returned
        as bytes, requiring manual decoding by the consumer.

        Args:
            url (str): Target endpoint including protocol and path (e.g. https://api.example.com/path)
            headers (dict): Custom HTTP headers provided as dictionary. Authentication headers should be
                            explicitly included when required.
            data (dict): Request body data to be JSON-serialized. Must be a serializable dictionary object.

        Returns:
            Tuple[Any, Any]: 2-element tuple containing:
                response.status: HTTP status code (e.g. 200, 404)
                content: Raw response body content as byte array

        Raises:
            aiohttp.ClientError: For network-related errors like failed DNS resolution or connection issues
            asyncio.TimeoutError: When the request exceeds 6-hour timeout limit

        Example:
            status, content = await forward_request_without_yield(
                'https://api.example.com/submit', 
                {'Authorization': 'Bearer token123'},
                {'key1': 'value1', 'key2': ['list_value']}
            )
            if status == 200:
                json_data = json.loads(content.decode('utf-8'))

        Requires:
            aiohttp library must be installed for asynchronous HTTP client functionality
        """
        async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=6 * 60 * 60)) as session:
            async with session.post(url=url, json=data, headers=headers) as response:
                content = await response.read()
                return response.status, content

    async def schedule(self, prompt: str) -> DispatchResult:
        """
        Async scheduling method for model inference requests.

        Args:
            prompt (str): Input text prompt to be processed by LLM.

        Returns:
            DispatchResult: Object containing:
                prefill_vllm_instance_uri: str
                decode_vllm_instance_uri: str
        """
        if self._balancer_handle is None:
            self._balancer_handle = yr.get_instance(name=BALANCER_ACTOR_NAME, namespace=VLLM_NAMESPACE)
        dispatch_result = await self._balancer_handle.dispatch_request.invoke()
        return dispatch_result


proxy_service = ProxyDeployment()


@app.post("/health")
async def health(request: Request):
    """
    Healthcheck endpoint to verify service availability.

    Returns a 200 OK response with "healthy" status to confirm the service
    is running correctly. This is typically used by Kubernetes or load balancers
    for liveness/readiness probes.

    Args:
        request (Request): FastAPI request object

    Returns:
        Response: HTTP 200 response with content "healthy"

    """
    return Response(status_code=200, content="healthy")


@app.post("/v1/completions")
async def openai_completions(raw_request: Request):
    """
    https://github.com/vllm-project/vllm/blob/main/benchmarks/disagg_benchmarks/disagg_prefill_proxy_server.py
    """
    from vllm.entrypoints.openai.protocol import CompletionRequest, ErrorResponse

    request_body = await raw_request.json()
    headers = {"Authorization": f"Bearer {os.environ.get('OPENAI_API_KEY')}",
               "X-Request-Id": raw_request.headers.get("X-Request-Id") or str(uuid.uuid4())}

    try:
        request = CompletionRequest(**request_body)
    except pydantic.ValidationError as e:
        return Response(
            content=json.dumps(ErrorResponse(
                message=str(e.errors(include_url=False)),
                type="Bad Request",
                code=HTTPStatus.BAD_REQUEST
            ).dict()),
            status_code=HTTPStatus.BAD_REQUEST,
            media_type="application/json"
        )

    dispatch_result = await proxy_service.schedule(request.prompt)
    logger.info(
        f"({headers['X-Request-Id']}) recv request: {request.prompt}, "
        f"prefill to: {dispatch_result.prefill_vllm_instance_uri},"
        f"decode to {dispatch_result.decode_vllm_instance_uri}"
    )

    try:
        prefill_request = request_body.copy()
        prefill_request['kv_transfer_params'] = {
            "do_remote_decode": True,
            "do_remote_prefill": False,
            "remote_engine_id": None,
            "remote_block_ids": None,
            "remote_host": None,
            "remote_port": None
        }
        prefill_request["max_tokens"] = 1
        prefill_request["stream"] = False
        if "stream_options" in prefill_request:
            del prefill_request["stream_options"]
        if dispatch_result.prefill_vllm_instance_uri:
            status_code, prefill_result = await proxy_service.forward_request_without_yield(
                f"{dispatch_result.prefill_vllm_instance_uri}/v1/completions",
                headers=headers,
                data=prefill_request,
            )
            if status_code != 200:
                logger.error(f"prefill request failed, status code:{status_code}, content:{prefill_result}")
            kv_transfer_params = json.loads(prefill_result.decode('utf-8')).get('kv_transfer_params', {})
            if kv_transfer_params:
                request_body["kv_transfer_params"] = kv_transfer_params

        decode_token_generator = proxy_service.forward_request(
            f"{dispatch_result.decode_vllm_instance_uri}/v1/completions",
            headers=headers,
            data=request_body,
        )
        status_code = 200
        # Only iterate once, get the status code and transmit it transparently
        async for status in decode_token_generator:
            status_code = status
            break
        return StreamingResponse(
            decode_token_generator,
            status_code=status_code,
            media_type="application/octet-stream",
        )
    except Exception as e:
        await proxy_service.record_exception_info(e)
        raise


@app.post("/scaleout")
async def scale_out():
    """
    scale out one inference instance, disaggregated PD not supported to scale out.
    """
    try:
        controller = yr.get_instance(name=CONTROLLER_ACTOR_NAME, namespace=VLLM_NAMESPACE)
        controller.scale_out_inference_ins.invoke(1, 0)
        return Response(status_code=200)
    except RuntimeError as e:
        logger.error(f"No available controller, failed to scale out : {str(e)}")
        return Response(status_code=500, content="Failed to scale out, reason: " + str(e))


def deploy_endpoint_to_cluster(host, port):
    """
    Deploys an API endpoint as a service in a distributed cluster.

    Args:
        host (str): IP address for HTTP server to listen on. 
        port (int): Port number for HTTP server. Must be available on target nodes.

    Returns:
        None: Returns nothing.
    """
    import uvicorn
    uvicorn.run(app, host=host, port=port)
