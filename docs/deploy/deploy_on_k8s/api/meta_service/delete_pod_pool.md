# 删除 Pod 资源池

## 功能介绍

K8s 上部署openYuanrong集群时，删除已创建的资源池。

## URI

DELETE /serverless/v1/podpools?id={id}&group={group}

查询参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| id    | 否，如果未设置，则必须设置 `group` 参数。 | String | pod 资源池 ID。 |
| group | 否，如果未设置，则必须设置 `id` 参数。    | String | pod 资源池组。<br> **约束：** 只设置该参数时，将删除组内的所有资源池。 |

## 请求参数

无

## 响应参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| code    | 是 | int    | 返回码。0 表示成功，非 0 表示失败。 |
| message | 是 | String | 错误信息。 |

## 请求示例

DELETE {[meta service endpoint](api-meta-service-endpoint)}/serverless/v1/podpools?id=pool1

## 响应示例

状态码：200

表示删除成功。

```json
{
    "code": 0,
    "message": ""
}
```

## 错误码

请参见[错误码](../../../../multi_language_function_programming_interface/api/error_code.md)。
