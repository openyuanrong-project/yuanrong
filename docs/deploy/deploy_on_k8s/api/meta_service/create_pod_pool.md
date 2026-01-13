# 创建 pod 资源池

## 功能介绍

创建 pod 资源池用于在 K8s 上部署openYuanrong集群时运行函数。该接口为异步接口，创建完成后，您可通过[查询接口](get_pod_pool.md)获取 Pod 资源池的状态。

## 接口约束

pool 池的亲和基于 K8s 调度能力实现，详见[将 Pod 指派给节点](https://kubernetes.io/zh-cn/docs/concepts/scheduling-eviction/assign-pod-node/){target="_blank"}，使用约束如下。

- 多个 pool 池之间默认反亲和部署，如果您需要修改，可通过配置 [PoolInfo 结构](api-data-struct-poolinfo)中的 `affinities` 参数实现。
- 配置 pool 池的亲和属性时，[PoolInfo 结构](api-data-struct-poolinfo)中 `node_selector` 与 `affinities` 参数会同时生效。配置冲突时可能导致 pool 池处于 `pending` 状态而无法拉起，建议您配置一项即可。

pool 池的自动扩缩通过 [PoolInfo 结构](api-data-struct-poolinfo)中的 `size` 和 `max_size` 参数控制，原理详见 [Pod 水平自动扩缩](https://kubernetes.io/zh-cn/docs/tasks/run-application/horizontal-pod-autoscale/){target="_blank"}，使用约束如下。

- 开启自动扩缩时: 每个实例创建请求最多触发一次 Pod 扩容，但扩容出来的 Pod 并不与该实例绑定。可能存在扩容的 Pod 被其他实例占用的情况，这时无法获得可用 Pod 的实例将在 2 分钟超时后返回创建失败。
- 固定资源池大小时：可以在 [PoolInfo 结构](api-data-struct-poolinfo)中的 `labels` 参数增加键值对 `{"yr-idle-to-recycle":5}` 配置 Pod 回收时间, 值的单位是秒。如果永不回收，配置为 `{"yr-idle-to-recycle":"unlimited"}` 即可。

## URI

POST /serverless/v1/podpools

## 请求参数

请求 Header 参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| Content-Type | 是 | String | 消息体类型。<br> **取值：** 建议填写 `application/json`。 |

<br> 请求 Body 参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| pools | 是 | Array of [PoolInfo](api-data-struct-poolinfo) objects | pool 信息数组。<br> **约束：** 数组长度大于 `0` |

(api-data-struct-poolinfo)=

<br> PoolInfo 类型参数

| **参数** | **是否必选** | **参数类型** | **描述**|
| ---------- | ----- | ---------- | ------------------------- |
| id                             | 是 | String              | pool 池 id，需保证全局唯一。<br> **取值：** 支持小写字母、数字以及中划线，允许长度 1-40，不以中划线开头或结尾 |
| group                          | 否 | String              | pool 池所属 group。<br> **取值：** 支持大小写字母、数字以及中划线，允许长度 1-40，不以中划线开头或结尾，不设置时使用默认值 default |
| reuse                          | 否 | Boolean             | pool 池是否复用。当 Pod 被函数实例独占时，如果 `reuse` 为 false，池子会立即补充新的 Pod，否则需要等到实例 Pod 被销毁时才会补充新的 pod。<br> **取值：** true 或 false, 默认 `false`。 |
| size                           | 否 | int                 | pool 池大小。当参数 `max_size` 值大于 `size` 时，`size` 为最小副本数。<br> **取值：** 大于等于 0, 默认值为 `0`。 |
| max_size                       | 否 | int                 | pool 池最大副本数。当 `max_size` 值大于参数 `size` 时，pool 池开启自动扩缩。<br> **取值：** 大于等于 0 且大于等于参数 `size` 的值。当 `size` 不等于 0, `max_size` 设置为 0 或者不设置时，`max_size` 实际值将与 `size` 保持一致。 |
| image                          | 否 | String              | runtime-manager 组件镜像 tag，不设置时使用版本默认镜像。<br> **取值：** 最大长度 200。 |
| init_image                     | 否 | String              | function-agent-init 组件镜像 tag，不设置时使用版本默认镜像。<br> **取值：** 最大长度 200。 |
| labels                         | 否 | Map[String,String]  | 自定义 pool 池标签。<br> **取值：** key-value 形式，遵循 K8s 规范（key 必须小于等于 63 个字符，以字母数字开头和结尾，包含连字符（-）、下划线（_）、点（.）和字母或数字）。 |
| environment                    | 否 | Map[String,String]  | 自定义 runtime-manager 容器环境变量。<br> **取值：** key-value 形式，遵循 K8s 规范（key 由字母、数字、下划线、点或连字符组成，但第一个字符不能是数字）。 |
| volumes                        | 否 | String              | pool 池需要声明的 volume 卷，支持 HostPath、PVC。<br> **取值：** 遵循 K8s 规范（name 必须小于等于 63 个字符，以字母数字开头和结尾，包含连字符（-）、下划线（_）、点（.）和字母或数字）。<br> **约束：** 如果无法正常解析，创建 pool 池时将忽略该配置参数。 |
| volume_mounts                  | 否 | String              | pool 池需要声明的 volumeMount 卷挂载。<br> **取值：** 遵循 K8s 规范（name 必须与 volumes 名称匹配；mountPath 不得包含 ‘:’）。<br> **约束：** 如果无法正常解析，创建 pool 池时将忽略该配置参数。 |
| resources                      | 是 | [ResourceRequirement](api-data-struct-resourcerequirement) object | pool 池使用资源声明。 |
| affinities                     | 否 | String              | pool 池亲和声明。<br> **取值：** 遵循 K8s 规范（配置 Pod 亲和策略时，topologyKey 不能为空；weight 范围为 1-100）。<br> **约束：** 如果无法正常解析，创建 pool 池时将忽略该配置参数。 |
| node_selector                  | 否 | Map[String,String]  | 节点标签匹配，pool 池将选择特定节点调度。<br> **取值：** 遵循 K8s 规范。 |
| runtime_class_name             | 否 | String              | 容器运行时名称。<br> **取值：** 不超过 64 字符并遵循 K8s 规范。 |
| tolerations                    | 否 | String              | pool 池可以容忍的污点。<br> **取值：** 遵循 K8s 规范。 |
| horizontal_pod_autoscaler_spec | 否 | String              | pool 池配置 HPA 声明。<br> **取值：** 遵循 k8s 规范（maxReplicas 不能小于 minReplicas）。 |
| topology_spread_constraints    | 否 | String              | pool 池配置拓扑分布约束。<br> **取值：** 遵循 k8s 规范。<br> **约束：** 当参数 `max_size` 大于 参数 `size` 的值时，不支持配置该参数 |
| pod_pending_duration_threshold | 否 | int                 | pool 池中 Pod 处于 Pending 状态时长告警阈值。pool 池中 Pod 持续处于 Pending 状态时长超过该值，将触发告警上报。<br> **取值：** 大于等于 0，等于 0 时实际使用默认值 `120` 秒。 |
| idle_recycle_time              | 否 | [IdleRecyclePolicy](api-data-struct-idlerecyclepolicy) object | 配置自动扩缩 pod 空闲回收时间。<br> **约束：** 仅当参数 max_size 大于 参数 size 的值时，配置生效。当开启租户隔离特性时，空闲回收时间配置优先级高于租户 Pod 复用窗口配置。 |

(api-data-struct-idlerecyclepolicy)=

<br> IdleRecyclePolicy 类型参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| reserved | 否 | int | 预留 pod 空闲回收时间。<br> **取值：** 大于等于 -1，单位为秒。-1 表示永不回收，0 表示未配置空闲回收。 |
| scaled   | 否 | int | 弹性 pod 空闲回收时间。<br> **取值：** 大于等于 -1，单位为秒。-1 表示永不回收，0 表示未配置空闲回收。 |

(api-data-struct-resourcerequirement)=

<br> ResourceRequirement 类型参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| requests | 是 | Map[String,String] | Pod request 资源定义。<br> **取值：** key-value 格式，遵循 k8s 规范。cpu 单位 m，表示毫核；mem 单位 Mi 或 Gi。 |
| limits   | 是 | Map[String,String] | Pod limit 资源定义。<br> **取值：** key-value 格式，遵循 k8s 规范。cpu 单位 m，表示毫核；mem 单位 Mi 或 Gi。 |

## 响应参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| code    | 是 | int                 | 错误码。0 表示成功，非 0 表示失败。 |
| message | 是 | String              | 错误信息。 |
| result  | 是 | [CreateResult](api-data-struct-createresult) object | 是 | 创建结果。 |

(api-data-struct-createresult)=

<br> CreateResult 类型参数

| **参数** | **是否必选** | **类型** | **描述** |
| ---------- | ----- | ---------- | ------------------------- |
| failed_pools | Array of String objects | 失败 id 列表 |

## 请求示例

POST {[meta service endpoint](api-meta-service-endpoint)}/serverless/v1/podpools

```json
{
    "pools": [
        {
            "id": "pool1",
            "size": 2,
            "max_size": 3,
            "group": "rg1",
            "reuse": true,
            "image": "runtime-manager:v1",
            "init_image": "function-agent-init:v1",
            "labels": {
              "label1": "val1"
            },
            "environment": {
              "env1": "key1"
            },
           "volumes": "[{\"name\":\"pv-function\",\"persistentVolumeClaim\":{\"claimName\":\"test-client-pvc-claim\"}}]",
      "volume_mounts": "[{\"name\":\"pv-function\",\"mountPath\":\"/home/sn/function-packages\"}]",
            "resources": {
                "limits": {
                    "cpu": "600m",
                    "memory": "512Mi"
                },
                "requests": {
                    "cpu": "600m",
                    "memory": "512Mi"
                }
            },
            "node_selector": {
              "node-role.kubernetes.io/controlplane": "true"
            },
            "runtime_class_name": "runc",
            "idle_recycle_time": {
                "reserved": -1,
                "scaled": 10
            },
            "tolerations": "[{\"key\":\"is-ds-worker-unready\",\"operator\": \"Equal\", \"value\": \"true\", \"effect\": \"NoSchedule\"}]",
            "affinities": "{\"nodeAffinity\": {\"requiredDuringSchedulingIgnoredDuringExecution\":{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"node-type\",\"operator\":\"In\",\"values\":[\"system\"]}]}]}}}",
            "horizontal_pod_autoscaler_spec": "{\"minReplicas\": 1, \"maxReplicas\": 2, \"metrics\":[{\"resource\": {\"name\":\"cpu\", \"target\":{\"averageUtilization\":20, \"type\":\"Utilization\"}}, \"type\":\"Resource\"}, {\"resource\": {\"name\":\"memory\", \"target\":{\"averageUtilization\":50, \"type\":\"Utilization\"}}, \"type\":\"Resource\"}]}",
            "topology_spread_constraints": "[{\"maxSkew\":1,\"minDomains\":1,\"topologyKey\":\"kubernetes.io/hostname\",\"whenUnsatisfiable\":\"DoNotSchedule\", \"matchLabelKeys\": [\"pod-template-hash\"],\"labelSelector\":{\"matchLabels\": {\"rg1\":\"rg1\"}} }]"
        }
    ]
}
```

## 响应示例

状态码：200

表示创建资源池成功。

```json
{
    "code": 0,
    "message": "",
    "result": {
        "failed_pools": null
    }
}
```

## 错误码

请参见[错误码](../../../../multi_language_function_programming_interface/api/error_code.md)。
