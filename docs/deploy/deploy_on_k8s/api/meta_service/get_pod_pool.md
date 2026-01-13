# 查询 pod 资源池

## 功能介绍

K8s 上部署openYuanrong集群时，查询已创建的 pod 资源池。

## URI

GET /serverless/v1/podpools?id={id}&group={group}&offset={offset}&limit={limit}

查询参数

| **参数** | **是否必选**  | **参数类型**  | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| id     | 否 | String | pod 资源池 ID, 精准匹配。 |
| group  | 否 | String | pod 资源池组, 精准匹配。 |
| offset | 否 | int    | 查询起始页。<br> **取值：** 大于等于 0，小于 0 时返回空。不填或填 0 时, 默认值为 `1`。|
| limit  | 否 | int    | 查询页的大小。<br> **取值：** 大于 0，小于等于 0 时返回空。不填时, 默认值为 `10`。 |

## 请求参数

无

## 响应参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| code    | 是 | int                                               | 返回码，0 表示成功，非 0 表示失败。 |
| message | 是 | String                                            | 错误信息。 |
| result  | 是 | [QueryResult](api-data-struct-queryresult) object | 查询结果。 |

(api-data-struct-queryresult)=

<br> QueryResult 类型参数

| **参数** | **是否必选**  | **参数类型** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| count | 是 | int                                                   | 符合查询条件的资源池数。 |
| pools | 否 | Array of [PoolInfo](api-data-struct-poolinfo) objects | pool 信息列表。 |

(api-data-struct-poolinfo)=

<br> PoolInfo 类型参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ---------- | ---------- | -------------------- |
| id                             | 是 | String                   | pool 池 ID。 |
| group                          | 是 | String                   | 所属组。 |
| reuse                          | 否 | bool                     | pool 池是否复用。当 pod 被函数实例使用并独占时，如果 reuse 为 false，池子会立即补充新的 pod，否则需要等到实例 pod 被销毁时，才会补充新的 pod。 |
| status                         | 是 | int                      | 资源池状态。<br> **取值：** 0（待创建）、1（创建中）、2（待更新）、3（更新中）、4（运行中）、5（创建失败）、6（待删除）。 |
| msg                            | 是 | String                   | 资源池状态信息。 |
| ready_count                    | 是 | int                      | running 状态的 pod 数量。 |
| size                           | 是 | int                      | pool 池子大小。 |
| max_size                       | 是 | int                      | pool 池最大副本数。 |
| image                          | 否 | String                   | runtime-manager 组件镜像 tag。 |
| init_image                     | 否 | String                   | function-agent-init 组件镜像 tag。 |
| labels                         | 否 | Map[String,String]       | 自定义 pool 池标签。 |
| environment                    | 否 | Map[String,String]       | 自定义 runtime-manager 容器环境变量。 |
| volumes                        | 否 | String                   | pool 池需要声明的 volume 卷，支持 HostPath 及 PVC。 |
| volume_mounts                  | 否 | String                   | pool 池需要声明的 volumeMount 卷挂载。 |
| resources                      | 是 | [ResourceRequirement](api-data-struct-resourcerequirement-get) object | pool 池使用资源声明。 |
| affinities                     | 否 | String                   | pool 池亲和声明。 |
| node_selector                  | 否 | Map[String,String]       | 节点标签匹配，pool 池将选择特定节点调度。 |
| runtime_class_name             | 否 | String                   | 容器运行时名称。 |
| tolerations                    | 否 | String                   | pool 池可以容忍的污点。 |
| horizontal_pod_autoscaler_spec | 否 | String                   | pool 池配置 HPA 声明。 |
| topology_spread_constraints    | 否 | String                   | pool 池配置拓扑分布约束。 |
| idle_recycle_time              | 否 | [IdleRecyclePolicy](api-data-struct-idlerecyclepolicy-get) object | 配置的自动扩缩 pod 空闲回收时间。 |

(api-data-struct-idlerecyclepolicy-get)=

<br> IdleRecyclePolicy 类型参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| reserved | 否 | int | 预留 pod 空闲回收时间。单位为秒。-1 表示永不回收，0 表示未配置空闲回收。 |
| scaled   | 否 | int | 弹性 pod 空闲回收时间。单位为秒。-1 表示永不回收，0 表示未配置空闲回收。 |

(api-data-struct-resourcerequirement-get)=

<br> ResourceRequirement 类型参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| requests | 是 | Map[String,String] | Pod request 资源定义。cpu 单位 m，表示毫核；mem 单位 Mi 或 Gi。 |
| limits   | 是 | Map[String,String] | Pod limit 资源定义。cpu 单位 m，表示毫核；mem 单位 Mi 或 Gi。 |

## 请求示例

GET {[meta service endpoint](api-meta-service-endpoint)}/serverless/v1/podpools?id=pool1&group=rg1

## 响应示例

状态码：200

创建查询成功。

```json
{
    "code": 0,
    "message": "",
    "result": {
        "count": 1,
        "pools": [
            {
                "id": "pool1",
                "group": "rg1",
                "size": 1,
                "max_size": 3,
                "ready_count": 1,
                "status": 4,
                "msg": "Running",
                "image": "cd-docker-hub.szxy5.artifactory.cd-cloud-artifact.tools.xxx.com/yuanrong_euleros_x86/runtime-manager:201.2.0.B992.20240301163827-zyw-1",
                "init_image": "",
                "reuse": true,
                "labels": null,
                "environment": null,
                "resources": {
                    "requests": {
                        "cpu": "600m",
                        "memory": "512Mi"
                    },
                    "limits": {
                        "cpu": "600m",
                        "memory": "512Mi"
                    }
                },
                "volumes": "",
                "volume_mounts": "",
                "pod_anti_affinities": "",
                "horizontal_pod_autoscaler_spec": "",
                "topology_spread_constraints": "",
                "idle_recycle_time": {
                    "reserved": 80,
                    "scaled": 20
                }
            }
        ]
    }
}
```

## 错误码

请参见[错误码](../../../../multi_language_function_programming_interface/api/error_code.md)。
