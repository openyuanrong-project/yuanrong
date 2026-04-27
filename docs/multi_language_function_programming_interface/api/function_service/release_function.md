# 发布函数

## 功能介绍

该 API 用于在 openYuanrong 集群，调用 meta_service 接口发布函数。

## 接口约束

使用默认生成版本号发布时，多次发布间隔须大于 ``1 s``。当前默认版本号生成依据为当前时间，短时间多次发布会生成重复版本号，会返回重复版本号错误。

## URI

`POST /serverless/v1/functions/{function_id}/versions`

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| function_id | String | 是 | 函数名称。 |

## 请求参数

### 请求 Header 参数

| **参数**     | **是否必选**             | **参数类型**                   | **描述** |
| ----------- | -------------------- | ------------------------- | ----------- |
| Content-Type | 是 | string | 消息体类型；建议填写 ``application/json``。 |

### 请求 Body 参数

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| revisionId | String| 是 | 函数 revisionId, 需为 latest 版本的。 |
| versionNumber | String | 否 | 版本号。<br> **约束**：取值字母数字开头结尾，可包含字母数字，中划线（-），下划线（_）及点（.），长度不超过 42。为空时默认生成格式形如 "v20060102-150405"。|
| versionDesc | String | 否 | 版本描述。  |
| kind | String | 否 | 函数类型，取值 ``faas`` 或者 ``yrlib``。<br> 建议按照函数注册时的 kind 填写。 |

## 响应参数

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| code | int | 是 | 返回码，``0`` 表示更新成功，非 ``0`` 则更新失败，更多信息参考[错误码](error-code-rest-api)。  |
| message | String | 是 | 返回错误信息。 |
| function | PublishFunctionResult | 是 | 更新结果。 |

### PublishFunctionResult 介绍

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| id | int | 是 | 函数 ID。  |
| functionVersionUrn | String | 是 | 函数版本 URN, 用于调用函数。 |
| revisionId | String| 是 | 函数 revisionId, 用于发布函数。 |

## 状态码

| **状态码** | **描述** |
|-----|----------|
| 200 | 请求成功（ok）。 |
| 500 | 内部服务器错误（Internal Server Error）。 |

## 请求示例

POST {[meta service endpoint](api-meta-service-endpoint)}/serverless/v1/functions/{function_id}/versions

```json
{
    "revisionId": "20240302084750473",
    "versionDesc": "try-mod-version2",
    "versionNumber": "B.2233",
    "kind": "yrlib"
}
```

## 响应示例

### 正常响应

```json
{
    "code": 0,
    "message": "SUCCESS",
    "function": {
        "id": "sn:cn:yrk:default:function:0-f-a:B.2233",
        "createTime": "2024-04-16 12:45:31.639 UTC",
        "updateTime": "",
        "functionUrn": "sn:cn:yrk:default:function:0-f-a",
        "name": "0-f-a",
        "tenantId": "default",
        "businessId": "yrk",
        "productId": "",
        "reversedConcurrency": 0,
        "description": "this is func",
        "tag": null,
        "functionVersionUrn": "sn:cn:yrk:default:function:0-f-a:B.2233",
        "revisionId": "20240416124531639",
        "codeSize": 3131,
        "codeSha256": "557ed9f661be0020c8a3243fa362a1aaccd0f5a4bb6cc8040a26bfd3844ccf14",
        "bucketId": "bucket-test-log1",
        "objectId": "a-1713271531641",
        "handler": "",
        "layers": null,
        "cpu": 500,
        "memory": 500,
        "runtime": "cpp11",
        "timeout": 500,
        "versionNumber": "B.2233",
        "versionDesc": "try-mod-version2",
        "environment": {},
        "customResources": {},
        "statefulFlag": 0,
        "lastModified": "2024-04-16 12:45:31.639 UTC",
        "Published": "2024-04-16 12:45:38.004 UTC",
        "minInstance": 0,
        "maxInstance": 100,
        "concurrentNum": 100,
        "funcLayer": null,
        "status": "",
        "instanceNum": 0,
        "device": {},
        "created": "2024-04-16 12:45:31.639 UTC"
    }
}
```

### 错误响应

```json
{
    "code": 4134,
    "message": "revisionId is non latest version"
}
```
