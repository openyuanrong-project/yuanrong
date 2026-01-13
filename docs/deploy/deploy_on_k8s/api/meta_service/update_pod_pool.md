# 更新 pod 资源池

## 功能介绍

K8s 上部署openYuanrong集群时，更新单个资源池的 size。

## URI

PUT /serverless/v1/podpools/{id}

路径参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| id | 是 | String | pod 资源池 ID。 |

## 请求参数

请求 Header 参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| Content-Type | 是 | String | 消息体类型。<br> **取值：** 建议填写 `application/json`。 |

<br> 请求 Body 参数

| **参数** | **是否必选** | **参数类型** | **描述**  |
| ---------- | ---------- | ---------- | -------------------- |
| size                           | 否 | int    | pool 池子大小。<br> **取值：** 大于等于 0, 默认值为 `0`。 |
| max_size                       | 否 | int    | pool 最大允许副本数。对于未开启自动扩缩 pool 池，忽略该配置。<br> **取值：** 大于等于 0 且大于等于 size。 |
| horizontal_pod_autoscaler_spec | 否 | String | pool 池配置 HPA 声明。当参数 `max_size` 大于 `size` 时，不支持该配置。<br> **取值：** 遵循 k8s 规范和约束。 |

## 响应参数

| **参数** | **是否必选**  | **参数格式** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| code    | 是 | int    | 返回码。0 表示成功，非 0 表示失败。 |
| message | 是 | String | 错误信息。 |

## 请求示例

PUT {[meta service endpoint](api-meta-service-endpoint)}/serverless/v1/podpools/pool1

```json
{
    "minReplicas": -1,
    "maxReplicas": 2,
    "metrics": [
        {
            "resource": {
                "name": "cpu",
                "target": {
                    "averageUtilization": -20,
                    "type": "Utilization"
                }
            },
            "type": "Resource"
        },
        {
            "resource": {
                "name": "memory",
                "target": {
                    "averageUtilization": 50,
                    "type": "Utilization"
                }
            },
            "type": "Resource"
        }
    ]
}
```

## 响应示例

状态码：200

表示更新成功。

```json
{
    "code": 0,
    "message": ""
}
```

## 错误码

请参见[错误码](../../../../multi_language_function_programming_interface/api/error_code.md)。
