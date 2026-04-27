# 查询函数列表

## 功能介绍

该 API 用于在 openYuanrong 集群，调用 meta_service 接口查询函数。

## 接口约束

## URI

`GET /serverless/v1/functions?versionNumber={version}&name={name}&pageIndex={pageIndex}&pageSize={pageSize}`

### 查询参数

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| name | String | 否 | 函数名称，模糊匹配 |
| version | String | 否 | 函数版本名称，精准匹配 |
| pageSize | int | 否 | 查询页数大小 |
| pageIndex | int | 否 | 查询起始页数 |

## 请求参数

### 请求 Header 参数

无

### 请求 Body 参数

无

## 响应参数

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| code | int | 是 |  | 返回码，0 表示查询成功，非 0 查询失败，更多信息参考[错误码](error-code-rest-api)。  |
| message | String | 是 | 返回错误信息。 |
| result | QueryResult Object | 是 | 查询结果。 |

### QueryResult

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| total | int | 是 | 符合查询条件的函数总数。 |
| functions | FunctionInfo 数组 | 是 | pool 信息数组。 |

### FunctionInfo Object

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| id | int | 是 | 函数 ID。  |
| functionVersionUrn | String | 是 | 函数版本 URN, 用于调用函数。 |
| revisionId | String| 是 | 函数 revisionId, 用于发布函数。 |
| name | String | 是 | 函数名称。|
| createTime | Date | 是 | 函数创建时间。|
| updateTime | Date | 是 | 函数更新时间。|
| versionNumber | String | 是 | 版本号。|
| runtime | String | 是 | 函数 runtime 类型。 |
| description | String | 否 |  函数描述。  |
| handler | String | 是 | call handler。 |
| cpu | int | 是 | 函数 CPU 大小，单位：`m（毫核）`。 |
| memory | int | 是 | 函数 MEM 大小，单位：`MB`。 |
| timeout | int | 是 | 函数调用超时时间。 |
| customResources | map | 否 | 函数自定义资源。<br> **约束**：key-value 格式，key 为 string，value 为 float 类型。|
| environment | map | 否 | 函数环境变量。<br> **约束**：key-value 格式，key 和 value 均为 string。|
| minInstance | int | 否 | 最小实例数（函数服务使用）。|
| maxInstance | int | 否 | 最大实例数（函数服务使用）。|
| concurrentNum | int | 否 | 实例并发度（函数服务使用）。|
| storageType | String | 是 | 代码包存储类型。取值：``local``、``s3``。 |
| codePath | String | 否 | 代码包本地路径，storageType 配置为 ``local`` 时生效。 |
| bucketId | String | 否 | 存储桶名。|
| objectId | String | 否 | 存储对象 ID。|

## 状态码

| **状态码** | **描述** |
|-----|----------|
| 200 | 请求成功（ok）。   |
| 500 | 内部服务器错误（Internal Server Error）。 |

## 请求示例

```shell
curl -X GET http://x.x.x.x:31182/serverless/v1/functions?pageIndex=1&pageSize=1&versionNumber=&name=
```

## 响应示例

### 正常响应

```json
{
    "code": 0,
    "message": "",
    "result": {
        "functions": [
            {
                "createTime": "2024-05-09 02:25:49.477 UTC",
                "updateTime": "2024-05-09 02:26:03.267 UTC",
                "functionUrn": "sn:cn:yrk:default:function:0-f-a",
                "name": "0-f-a",
                "tenantId": "default",
                "businessId": "yrk",
                "productId": "",
                "reversedConcurrency": 0,
                "description": "this is func",
                "tag": null,
                "functionVersionUrn": "sn:cn:yrk:default:function:0-f-a:$latest",
                "revisionId": "20240509022603267",
                "codeSize": 3753,
                "codeSha256": "39267d6c674f7ae5f99a284323d1f7f036b68e055e55f23a1fdcaa14c6e267e0",
                "bucketId": "bucket-test-log1",
                "objectId": "a-1715221563270",
                "handler": "",
                "layers": null,
                "cpu": 500,
                "memory": 500,
                "runtime": "cpp11",
                "timeout": 500,
                "versionNumber": "$latest",
                "versionDesc": "$latest",
                "environment": {},
                "customResources": null,
                "statefulFlag": 0,
                "lastModified": "",
                "Published": "2024-05-09 02:25:49.477 UTC",
                "minInstance": 0,
                "maxInstance": 100,
                "concurrentNum": 100,
                "funcLayer": [],
                "status": "",
                "instanceNum": 0,
                "device": {}
            }
        ],
        "total": 3
    }
}
```

### 错误响应

```json
{
    "code": 444,
    "message": "xxx error"
}
```
