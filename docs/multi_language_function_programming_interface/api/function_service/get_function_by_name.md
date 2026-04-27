# 查询函数的所有版本

## 功能介绍

该 API 用于在 openYuanrong 集群，调用 meta_service 接口查询函数版本列表。

## 接口约束

无

## URI

GET /serverless/v1/functions/{name}/versions?pageIndex={pageIndex}&pageSize={pageSize}

### 查询参数

| **名称**     | **类型**  | **是否必填**  | **约束** | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- | ------------------------- |
| pageSize | int | 否 |   | 查询页数大小 |
| pageIndex | int | 否 |   | 查询起始页数 |

## 请求参数

### 请求 Header 参数

无

### 请求 Body 参数

无

## 响应参数

| **名称**     | **类型**  | **是否必填**  | 约束 | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- | ------------------------- |
| functions | FunctionVersionInfo Array | 否 |  | 函数版本列表  |
| pageInfo | PageInfo | 是 |  | 分页信息 |

### PageInfo

| **名称**     | **类型**  | **是否必填**  | 约束 | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- | ------------------------- |
| total | int | 是 |  | 符合查询条件的函数总数 |
| pageSize | int | 否 |   | 查询页数大小 |
| pageIndex | int | 否 |   | 查询起始页数 |

### FunctionVersionInfo Object

| **名称**     | **类型**  | **是否必填**  | 约束 | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- | ------------------------- |
| id | String | 是 |  | 函数 ID  |
| functionVersionUrn | String | 是 |  | 函数版本 URN, 用于调用函数 |
| revisionId | String| 是 | | 函数 revisionId, 用于发布函数 |
| name | String | 是 | | 函数名称|
| createTime | Date | 是 | | 函数创建时间|
| updateTime | Date | 是 | | 函数更新时间|
| versionNumber | String | 是 | | 版本号|
| runtime | String | 是 | | 函数 runtime 类型 |
| description | String | 否 |  |  函数描述  |
| handler | String | 是 |  | call handler |
| cpu | int | 是 |  | 函数 CPU 大小，单位：m（毫核） |
| memory | int | 是 |  | 函数 MEM 大小，单位：MB |
| timeout | int | 是 |  | 函数调用超时时间 |
| customResources | map | 否 | key-value 格式，key 为 string， value 为 float 类型| 函数自定义资源 |
| environment | map | 否 | key-value 格式，key 和 value 均为 string | 函数环境变量 |
| minInstance | int | 否 | | 最小实例数 (函数服务使用)|
| maxInstance | int | 否 | | 最大实例数(函数服务使用) |
| concurrentNum | int | 否 | | 实例并发度(函数服务使用) |
| storageType | String | 是 | 取值 local,s3 | 代码包存储类型 |
| codePath | String | 否 | | 代码包本地路径，storageType 配置为 local 时生效 |
| bucketId | String | 否 | |  存储桶名|
| objectId | String | 否 | |  存储对象 ID|

## 状态码

| **状态码** | **描述** |
|-----|----------|
| 200 | ok |
| 500 | Internal Server Error 内部服务器错误 |

## 请求示例

```shell
curl -X GET http://x.x.x.x:31182/serverless/v1/functions/0@faaspy@hello2/versions?pageIndex=1&pageSize=1
```

## 响应示例

正常响应

```json
{
    "functions": [
        {
            "id": "sn:cn:yrk:default:function:0@faaspy@hello2:1",
            "createTime": "2024-03-02 08:31:12.724 UTC",
            "updateTime": "2024-03-02 08:47:50.473 UTC",
            "functionUrn": "sn:cn:yrk:default:function:0@faaspy@hello2",
            "name": "0@faaspy@hello2",
            "tenantId": "default",
            "businessId": "yrk",
            "productId": "",
            "reversedConcurrency": 0,
            "description": "this is a function",
            "tag": null,
            "functionVersionUrn": "sn:cn:yrk:default:function:0@faaspy@hello2:1",
            "revisionId": "20240302084750473",
            "codeSize": 0,
            "codeSha256": "",
            "bucketId": "",
            "objectId": "",
            "handler": "handler.my_handler",
            "layers": null,
            "cpu": 600,
            "memory": 512,
            "runtime": "python3.9",
            "timeout": 600,
            "versionNumber": "1",
            "versionDesc": "version2",
            "environment": {},
            "customResources": null,
            "statefulFlag": 0,
            "lastModified": "",
            "Published": "2024-03-02 09:07:53.390 UTC",
            "minInstance": 0,
            "maxInstance": 10,
            "concurrentNum": 10,
            "funcLayer": [],
            "status": "",
            "instanceNum": 0,
            "device": {},
            "created": ""
        }
    ],
    "pageInfo": {
        "pageIndex": "1",
        "pageSize": "1",
        "total": 2
    }
}
```

错误响应

```json
{
    "functions": null,
    "pageInfo": {
        "pageIndex": "1",
        "pageSize": "1",
        "total": 0
    }
}
```
