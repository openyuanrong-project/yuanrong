# 创建预留实例配置

## 功能介绍

该 API 用于在 openYuanrong 集群，调用 meta_service 接口针对函数服务创建特定标签预留实例配置。

## 接口约束

- 函数版本需存在。

- 同一个函数版本最多创建 100 个 Label。

## URI

`POST /serverless/v1/functions/reserve-instance`

## 请求参数

### 请求 Header 参数

| **参数**      | **是否必选** | **参数类型**                   | **描述**              |
|-------------|----------| ------------------------- |---------------------|
| X-Tenant-Id | 否        | string | 租户 ID，默认 ``default``。 |
| X-Trace-Id  | 否        | string | traceID。             |

### 请求 Body 参数

| **名称**              | **类型**                    | **是否必填** | **描述**                                      |
|---------------------|---------------------------|----------|---------------------------------------------|
| funcName            | String                    | 是        | 函数服务按照 0@{serviceName}@{funcName} 格式填写。<br> **约束**：其中 ServiceName 为 1 - 16 位字母，数字组合；funcName 为小写字母开头，可使用小写字母、数字、中划线组合，长度不超过 127 位。|
| version             | String                    | 是        | 函数版本号。                                       |
| instanceLabel       | String                    | 是        | 标签名称，不允许为空。<br> **约束**：超过 63 个字符，只能包含大小写字母，中划线，点。                                    |
| instanceConfigInfos | Array[InstanceConfigInfo] | 是        | 预留实例配置，不允许配置重复 cluster。                      |

#### InstanceConfigInfo

| **名称**     | **类型**  | **是否必填** | **描述** |
| ----------- | -------------------- |----------|--------|
| clusterId | String | 否        | 集群 ID。默认 ``cluster001``，不超过 64 个字符，集群 ID 需存在，通过部署配置指定（参考部署文档参数配置）。  |
| maxInstance | String | 否        | 最大实例数。取值：``0-1000``。  |
| minInstance | String | 否        | 最小实例数。取值：``0-1000``。  |

## 响应参数

| **名称**     | **类型**                    | **是否必填**  | **描述**                 |
| ----------- |---------------------------| ------------------------- | ------------------------|
| code | int                       | 是 | 返回码，``0`` 表示创建成功，非 ``0`` 则创建失败，更多信息参考[错误码](error-code-rest-api)。 |
| message | String                    | 是 | 返回错误信息。                 |
| instanceConfigInfos | Array[InstanceConfigInfo] | 是        | 预留实例配置。                 |
| reserveInsBaseInfo | ReserveInsBaseInfo        | 是        | 基础配置信息。                 |

### ReserveInsBaseInfo

| **名称**     | **类型**  | **是否必填**  | **描述** |
| ----------- | -------------------- | ------------------------- | --------|
| funcName            | String                    | 是        | 函数名称。   |
| version             | String                    | 是        | 函数版本号。  |
| instanceLabel       | String                    | 否        | 标签名称。   |

## 状态码

| **状态码** | **描述** |
|-----|----------|
| 200 | 请求成功（ok）。   |
| 400 | 错误的请求（Bad Request）。 |
| 500 | 内部服务器错误（Internal Server Error）。 |

## 请求示例

POST {[meta service endpoint](api-meta-service-endpoint)}/serverless/v1/functions/reserve-instance

```json
{
  "funcName": "0@faaspy@hello",
  "version": "latest",
  "instanceLabel": "label001",
  "instanceConfigInfos": [
    {
      "clusterId": "cluster001",
      "maxInstance": 101,
      "minInstance": 0
    }
  ]
}
```

## 响应示例

### 正常响应

```json
{
  "code": 0,
  "message": "",
  "instanceConfigInfos": [
    {
      "clusterId": "cluster001",
      "maxInstance": 101,
      "minInstance": 0
    }
  ],
  "reserveInsBaseInfo": {
    "funcName": "0@faaspy@hello",
    "version": "latest",
    "instanceLabel": "label001"
  }
}
```

### 错误响应

```json
{
  "code": 4115,
  "message": "function [0@faaspy@hello1] is not found. check input parameters"
}
```
