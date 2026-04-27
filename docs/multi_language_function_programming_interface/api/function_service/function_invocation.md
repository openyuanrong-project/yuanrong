# 调用服务

## 功能介绍

同步调用函数服务。

## URI

POST /serverless/v1/functions/{functionVersionURN}/invocations

路径参数：

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| functionVersionURN | 是 | String | 函数版本 URN。 |

## 请求参数

请求 Header 参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| Content-Type               | 是 | string | 消息体类型。<br> **取值：** 建议填写 application/json。 |
| X-Instance-Cpu             | 否 | string | 函数实例所需 CPU。 |
| X-Instance-Memory          | 否 | string | 函数实例所需 Memory。 |
| X-Instance-Custom-Resource | 否 | string | 函数实例所需自定义资源。 |
| X-Pool-Label               | 否 | string | 函数实例所需亲和的资源池标签。 |
| X-Instance-Label           | 否 | string | 在具有该标签的函数实例上运行。 |
| X-Instance-Session         | 否 | string | 指定实例会话调用，会话与实例唯一绑定。<br> 样例：{"sessionID":"abc","sessionTTL":10,"concurrency": 5}，其中，sessionID 不超过 63 位，sessionTTL 不小于 0，单位：秒，concurrency 不能超过函数中配置的 concurrentNum，当 concurrency 为 -1 时，表示独占整个函数实例。 |

<br> 请求 Body 参数

用户函数自定义格式。

## 响应参数

用户函数自定义格式。

## 请求示例

POST {[frontend endpoint](api-frontend-endpoint)}/serverless/v1/functions/{functionVersionURN}/invocations

```json
{
    "name":"yuanrong"
}
```

## 响应示例

状态码：200

表示调用函数成功。

```text
"hello yuanrong"
```

## 错误码

请参见[错误码](../error_code.md)
