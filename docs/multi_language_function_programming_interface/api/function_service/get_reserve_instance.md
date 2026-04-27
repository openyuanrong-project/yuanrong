# 查询预留实例配置

## 功能介绍

该 API 用于在 openYuanrong 集群，调用 meta_service 接口针对函数服务查询特定标签预留实例配置。

## 接口约束

无。

## URI

GET /serverless/v1/functions/reserve-instance

## 请求参数

### 请求 Header 参数

| **参数**      | **是否必选** | **参数类型**                   | **描述**              |
|-------------|----------| ------------------------- |---------------------|
| X-Tenant-Id | 否        | string | 租户 ID，默认 default |
| X-Trace-Id  | 否        | string | traceID             |

### 请求 Body 参数

| **名称**              | **类型**                    | **是否必填** | 约束                                                                              | **描述**                                      |
|---------------------|---------------------------|----------|---------------------------------------------------------------------------------|---------------------------------------------|
| funcName            | String                    | 是        | 其中 ServiceName 为 1 - 16 位字母，数字组合；funcName 为小写字母开头，可使用小写字母、数字、中划线组合，长度不超过 127 位。 | 函数服务按照 0@{serviceName}@{funcName} 格式填写 |
| version             | String                    | 是        |                                                                                 | 函数版本号                                       |

## 响应参数

| **名称**  | **类型**                     | **是否必填**  | 约束 | **描述**   |
|---------|----------------------------| ------------------------- | ---------------------- |----------|
| total   | int                        | 是 |  | 总数       |
| results| Array[GetReserveInsResult] | 是 |  | 预留实例配置数组 |

### GetReserveInsResult

| **名称**     | **类型**  | **是否必填** | 约束                       | **描述** |
| ----------- | -------------------- |----------|--------------------------|--------|
| instanceLabel       | String                    | 否        |                                         | 标签名称                                   |
| instanceConfigInfos | Array[InstanceConfigInfo] | 是        |           | 预留实例配置                      |

### InstanceConfigInfo

| **名称**     | **类型**  | **是否必填** | 约束                       | **描述** |
| ----------- | -------------------- |----------|--------------------------|--------|
| clusterId | String | 否        |  | 集群 ID  |
| maxInstance | String | 否        |                  | 最大实例数  |
| minInstance | String | 否        |                 | 最小实例数  |

## 状态码

| **状态码** | **描述** |
|-----|----------|
| 200 | ok |
| 400 | Bad Request 错误的请求 |
| 500 | Internal Server Error 内部服务器错误 |

## 请求示例

```json
{
  "funcName": "0@faaspy@hello",
  "version": "latest"
}
```

## 响应示例

正常响应

```json
{
  "total": 2,
  "results": [
    {
      "instanceLabel": "label001",
      "instanceConfigInfos": [
        {
          "clusterId": "cluster001",
          "maxInstance": 101,
          "minInstance": 0
        }
      ]
    },
    {
      "instanceLabel": "",
      "instanceConfigInfos": [
        {
          "clusterId": "cluster001",
          "maxInstance": 100,
          "minInstance": 0
        }
      ]
    }
  ]
}
```

错误响应，更多信息参考[错误码](error-code-rest-api)

```json
{
  "code":4115,
  "message": ""
}
```
