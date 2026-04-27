# 订阅流服务

## 功能介绍

订阅函数服务生产的流。接口使用 HTTP Server Sent Event 给客户端流式返回响应，并在每个流式返回的消息结尾增加分隔符 `\n` 用于控制消息的分隔。

建议您在流生产函数完所有消息生产后，发送一个 openYuanrong 提供的流生产结束符 `\xE0\xFF\xE0\xFF` 来通知 frontend 服务终止接收该流的消息订阅。

## URI

GET /serverless/v1/stream/subscribe

## 请求参数

请求 Header 参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| X-Stream-Name | 是 | string | 流名称。 |
| X-Expect-Num  | 否 | string | 流单次接收的包个数。 |
| X-Timeout-Ms  | 否 | string | 流单次接收的超时时间。 |

<br> 请求 Body 参数

无

## 响应参数

用户自定义的流数据。

## 请求示例

GET {[frontend_endpoint](api-frontend-endpoint)}/serverless/v1/stream/subscribe

## 响应示例

状态码：200

表示订阅流成功。

```text
aaa

bbb
```

## 错误码

请参见[错误码](error-code-rest-api)
