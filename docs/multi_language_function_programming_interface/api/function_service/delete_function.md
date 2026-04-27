# 删除函数

## 功能介绍

该 API 用于在 openYuanrong 集群，调用 meta_service 接口删除函数，如果函数存在多个版本，均会删除。

## 接口约束

无

## URI

`DELETE /serverless/v1/functions/{name}?versionNumber={version}`

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| name | String | 是 | 函数名称。 |
| version | String | 否 | 函数版本。 |

## 请求参数

无

## 响应参数

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| code | int | 否 | 返回码，``0`` 表示创建成功，非 ``0`` 表示删除失败，更多信息参考[错误码](error-code-rest-api)。  |
| message | String | 否 | 返回错误信息。 |

## 状态码

| **状态码** | **描述** |
|-----|----------|
| 200 | 请求成功（ok）。   |
| 400 | 错误的请求（Bad Request）。 |
| 500 | 内部服务器错误（Internal Server Error）。 |

## 请求示例

```shell
curl -X DELETE  http://X.X.X.X:31182/serverless/v1/functions/0@faaspy@hello3
```

## 响应示例

### 正常响应

```json
{}
```

### 错误响应

```json
{
    "code": 4115,
    "message": "function [0@faaspy@hello3] is not found. check input parameters"
}
```
