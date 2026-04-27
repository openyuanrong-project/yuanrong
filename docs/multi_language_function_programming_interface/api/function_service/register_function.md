# 注册函数

## 功能介绍

该 API 用于 openYuanrong 集群，调用 meta_service 接口注册函数。

## 接口约束

函数名需全局唯一。

## URI

`POST /serverless/v1/functions`

## 请求参数

### 请求 Header 参数

| **参数**     | **是否必选**             | **参数类型**                   | **描述** |
| ----------- | -------------------- | ------------------------- | ----------- |
| Content-Type | 是 | string | 消息体类型。建议填写 `application/json`。 |
| x-storage-type | 是 | string | 代码包上传方式，取值: ``local`` 或 ``s3`` ；不通过接口上传代码，默认 ``local``。 |

### 请求 Body 参数

| **名称**     | **类型**  | **是否必选**  | **描述**                                                                                           |
| ----------- | -------------------- | ------------------------- | --------------------------------------------------------------------------------------------------|
| name | String | 是 | 函数名称。函数服务按照 `0@{serviceName}@{funcName}` 格式填写， yrlib 函数按照 `0-{serviceName}-{funcName}` 格式填写，需唯一。<br> **约束**：`ServiceName` 需为 1 - 16 位字母、数字组合；`funcName` 需为小写字母开头，可使用小写字母、数字、中划线（-）组合，长度不超过 127 位。 |
| runtime | String | 是 | 函数 runtime 类型。                                                                                    |
| description | String | 否 | 函数的描述。                                                                                             |
| handler | String | 否 | 调用处理器。<br> 当 `kind` 为 ``faas`` 时，建议填写。                                                                                    |
| kind | String | 否 | 函数类型。<br> 取值： ``faas`` 或 ``yrlib``。默认 ``yrlib``。                                                                           |
| cpu | int | 是 | 函数 CPU 大小，单位：`1/1000` 核。                                                                                  |
| memory | int | 是 | 函数 MEM 大小，单位：`MB`。                                                                                  |
| timeout | int | 是 | 函数调用超时时间。<br> 最大值：``8640000 s``。当参数超过最大值时，自动设置为最大值。若不填默认 ``900 s``。                                                                                        |
| customResources | map | 否 | 函数自定义资源。<br> **约束**：key-value 格式，其中 key 为 string 类型，value 为 float 类型。                                                                                        |
| environment | map | 否 | 函数环境变量，内置环境变量定义。<br> **约束**：key-value 格式，其中 key 和 value 均为 string 类型。                                                                    |
| extendedHandler | ExtendedHandler | 否 | 配置 init handler 信息。                                                                               |
| extendedTimeout | ExtendedTimeout | 否 | 配置 init 超时信息。                                                                                     |
| minInstance | String | 否 | 最小实例数（函数服务使用）。                                                                               |
| maxInstance | String | 否 | 最大实例数（函数服务使用）。                                                                                |
| concurrentNum | String | 否 | 实例并发度（函数服务使用）。                                                                                |
| storageType | String | 否 | 代码包存储类型。``s3``: 代码包存储在 minio 中；``local``: 代码包存储在磁盘中；``copy``: 代码包存储在磁盘中并且需要拷贝至容器路径。<br> 取值：``local``、``s3`` 或 ``copy``，建议填写。                                                                                        |
| codePath | String | 否 | 代码包本地路径。`storageType` 配置为 ``local`` 或 ``copy`` 时生效。                                                           |
| s3CodePath | S3Object | 否 | 代码包 S3 路径。`storageType` 配置为 ``s3`` 时生效。                                                                 |
| poolId | String | 否 | 自定义亲和 pool 池 ID。函数实例创建资源不足（或者亲和条件不满足），内核创建指定 poolID 的 POD，用于实例调度。<br> **约束**：仅对函数服务生效。配置约束与创建 pool 池接口一致。                                                                |
| resourceAffinitySelectors | ResourceAffinitySelector | 否 | 函数调度亲和、优先级配置。                                                                                     |
| enable_agent_session | boolean | 否 | 是否启用 AI Agent 会话。开启后支持 `wait`/`notify` SDK 及会话亲和性调度。默认 ``false``。 |

:::{Note}

当 `storageType` 配置为 ``copy`` 时，openYuanrong会将 codePath 里的代码包拷贝到容器其他目录缓存，相比 local 方式，可以降低加载代码时磁盘 IO 带来的性能下降，建议设置目录权限为 ``755``，确保有权限被容器内 sn 用户拷贝。

:::

#### 支持 Runtime 类型

| **函数类型**     | **支持 runtime**  |
| ----------- | -------------------- |
| faas | java8、java11、java17、java21、python3.6、python3.8、python3.9、python3.11、posix-custom-runtime。 |
| yrlib | cpp11、java1.8、java11、java17、python3.6、python3.8、python3.9、python3.11、go1.13。 |

#### S3Object 介绍

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| bucketId | String | 否 |  存储桶名。|
| objectId | String | 否 |  存储对象 ID。|
| bucketUrl | String | 否 |  存储 URL。|
| sha512 | String | 否 |  代码包 sha512 值，可通过执行 sha512sum 命令获取。|

#### ExtendedHandler 介绍

| **名称**      | **类型**  | **是否必填**  | **描述**                         |
|-------------| -------------------- | ------------------------- |--------------------------------|
| initializer | String | 否 | 初始化接口，函数服务按需配置。    |
| pre_stop     | String | 否 | 停止接口，函数停止之前执行的退出逻辑，函数服务按需配置。 |

#### ExtendedTimeout 介绍

| **名称**         | **类型**  | **是否必填**  | **描述**                         |
|----------------| -------------------- | ------------------------- |--------------------------------|
| initializer    | int | 否 | 初始化超时时间，函数服务按需配置。            |
| pre_stop | int | 否 | 函数停止超时时间，最大值 ``180s``，函数服务按需配置。 |

:::{Note}

- 当 `storageType` 配置为 ``copy`` 或 ``s3`` 时，`init` 超时时间需要包含 `init handler` 的执行时间，以及代码包的拷贝或者下载时间，对于超大代码包，建议设置一个较大的超时时间。
- 当 `pre_stop` 执行时间超过函数停止超时时间时，会强制终止 `pre_stop` 执行并继续进行实例的退出。

:::

#### ResourceAffinitySelector 介绍

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| group | String | 否 |  资源组名称。|
| priority | int | 否 | 优先级（字段预留）。|

## 响应参数

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| code | int | 是 | 返回码。``0`` 表示创建成功，非 ``0`` 则创建失败，更多信息参考[错误码](error-code-rest-api)。  |
| message | String | 是 | 返回错误信息。 |
| function | CreateFunctionResult | 是 | 创建结果。 |

### CreateFunctionResult 介绍

| **名称**     | **类型**  | **是否必填**  | **描述**               |
| ----------- | -------------------- | ------------------------- | ------------------------- |
| id | String | 是 |  函数 ID。  |
| functionVersionUrn | String | 是 |  函数版本 URN, 用于调用函数。 |
| revisionId | String| 是 | 函数 revisionId, 用于发布函数。 |

## 状态码

| **状态码** | **描述** |
|-----|----------|
| 200 | 请求成功（ok）。   |
| 400 | 错误的请求（Bad Request）。 |
| 500 | 内部服务器错误（Internal Server Error）。 |

## 请求示例

POST {[meta service endpoint](api-meta-service-endpoint)}/serverless/v1/functions

```json
{
    "name": "0@faaspy@hello3",
    "runtime": "python3.9",
    "description": "this is a function",
    "handler": "handler.my_handler",
    "kind": "faas",
    "cpu": 600,
    "memory": 512,
    "timeout": 600,
    "customResources": {},
    "environment": {},
    "extendedHandler": {
        "initializer": "handler.init",
        "pre_stop": "test.prestop"
    },
    "extendedTimeout": {
        "initializer": 600,
        "pre_stop": 10
    },
    "minInstance": "0",
    "maxInstance": "10",
    "concurrentNum": "2",
    "storageType": "local",
    "codePath": "/home/sn/function-packages",
    "s3CodePath": {
    },
    "poolId": "pool1",
    "resourceAffinitySelectors": [
          {
            "group": "rg1",
            "priority": 1
        },
         {
            "group": "rg2",
            "priority": 2
        }
    ]
}
```

## 响应示例

### 正常响应

```json
{
    "code": 0,
    "message": "SUCCESS",
    "function": {
        "id": "sn:cn:yrk:default:function:0@faaspy@hello3:latest",
        "createTime": "",
        "updateTime": "",
        "functionUrn": "",
        "name": "",
        "tenantId": "",
        "businessId": "",
        "productId": "",
        "reversedConcurrency": 0,
        "description": "",
        "tag": null,
        "functionVersionUrn": "sn:cn:yrk:default:function:0@faaspy@hello3:latest",
        "revisionId": "20240301093328978",
        "codeSize": 0,
        "codeSha256": "",
        "bucketId": "",
        "objectId": "",
        "handler": "",
        "layers": null,
        "cpu": 0,
        "memory": 0,
        "runtime": "",
        "timeout": 0,
        "versionNumber": "",
        "versionDesc": "",
        "environment": null,
        "customResources": null,
        "statefulFlag": 0,
        "lastModified": "",
        "Published": "",
        "minInstance": 0,
        "maxInstance": 0,
        "concurrentNum": 0,
        "funcLayer": null,
        "status": "",
        "instanceNum": 0,
        "device": {},
        "created": ""
    }
}
```

### 错误响应

```json
{
    "code": 4101,
    "message": "the function name already exists. rename your function"
}
```

## 内置环境变量

| **名称**            | **是否必填** | **描述**                                 |
|------------------ |----------|----------------------------------------|
| YR_SSL_ENABLE     | 否        | 是否开启函数内置服务的 SSL 鉴权（当前仅支持配置 metrics 接口）。 |
| YR_SSL_ROOT_FILE  | 否        | 根证书文件地址。                                |
| YR_SSL_CERT_FILE  | 否        | 中间证书文件地址。                               |
| YR_SSL_KEY_FILE   | 否        | 加密私钥文件地址。                               |
| YR_SSL_PASSPHRASE | 否        | 私钥解密口令（明文传入后会加密存储在实例元信息中）。              |
