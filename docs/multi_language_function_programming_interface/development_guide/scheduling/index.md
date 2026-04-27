# 调度

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   logical_resource
   affinity
```

openYuanrong 在选择函数实例的运行节点时，会基于以下因素决策。

## 资源

每个无状态函数或有状态函数都可以指定需要的资源。可供部署的节点有以下状态：

- 可选：节点具有 openYuanrong 函数指定的资源且资源空闲可用。
- 不可选：节点不具备所需资源或资源正在被其他 openYuanrong 函数使用。

资源需求是硬性要求，只有在有可选节点时才会运行有状态函数或无状态函数。当有多个可选节点时，openYuanrong 会通过调度策略选择一个合适的节点使用。

参考[资源](./logical_resource.md)章节了解如何指定节点和 openYuanrong 函数的资源。

## 调度策略

openYuanrong 会基于以下调度策略选择最佳的节点运行 openYuanrong 函数。

### 本地优先

从 openYuanrong 集群中某一节点发起的运行 openYuanrong 函数的请求，会优先在本节点调度。即使本地节点的资源利用率较高，仍优先考虑局部性。

### 同租户优先

在集群中有多个租户的情况下，将优先考虑将 openYuanrong 函数分配到已有相同租户任务运行的节点，提升同一租户的资源利用率。

### 最大剩余资源优先

在本节点资源不足的情况下，优先选择剩余资源多的节点，优化节点的任务负载。

### 亲和

您可通过在节点或者 openYuanrong 函数上设置标签，实现新调度的 openYuanrong 函数部署在特定标签的节点上，或者和特定标签的 openYuanrong 函数部署在相同的节点上。通过亲和，您可以自定义策略实现更灵活的调度满足业务需要。

参考[亲和](./affinity.md)章节了解如何配置 openYuanrong 函数亲和属性。
