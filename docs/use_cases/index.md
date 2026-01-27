# 使用案例

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   llm_on_multiple_machines
   accelerate_llm_instance_scaling
   microservice-serverless
```

openYuanrong 可作为 AI 智算的基础设施，用于开发 AI 推理、强化学习等应用，也可作为通用计算的基础设施，开发大数据分析、HPC（高性能计算）等应用。

## openYuanrong 作为 AI 基础设施

openYuanrong 通过开源和兼容方式提供了常用 AI 领域应用框架，简化端到端 AI 应用的开发。内置异构分布式多级缓存能力支持 LLM 高效训练和推理。按需组件化灵活部署方便和已有业务集成。

1. **生态开放**：面向不同领域提供开源应用框架，并通过 adaptor 兼容常用的 vLLM、Verl 等框架，存量应用零代码修改接入。

2. **高效推理训练**：内置异构分布式多级缓存能力，推理可快速传递模型参数、KV Cache 数据，训练可零冗余切换训推参数。

3. **灵活部署**：支持全量或轻量化部署，支持在开源 Kubernetes 或 云上集群部署，方便和已有业务系统集成。

### AI 智算场景案例

- [基于 vLLM 部署多机 PD 分离服务，长序列推理 TTFT 下降 20%](./llm_on_multiple_machines.md)
- [推理实例模型加载速度 10 倍提升，快速弹性响应业务流量变化](./accelerate_llm_instance_scaling.md)

## openYuanrong 作为通用计算基础设施

微服务、大数据分析和 HPC 是通用计算场景中常见的分布式应用。openYuanrong 多语言函数编程接口提供了单机程序分布式并行化能力，简化了开发复杂度。函数粒度的应用实例极速弹性响应业务，不同应用负载支持共资源池部署，保证高性能的同时大幅提升资源利用率。

1. **简化开发**：openYuanrong 多语言函数编程接口支持 Python、C++、Java 常用开发语言。通过抽象资源，自适应动态调度隐藏了弹性、分布式调度等复杂性，简化了分布式应用的开发。

2. **极速弹性**：基于快照的冷启动加速能力让函数应用实例可毫秒级别全自动扩缩容响应业务流量，无需预置大量资源保障业务峰值需求，资源利用率大幅提升。

3. **多样负载共池**：openYuanrong 集群支持微服务、大数据分析、HPC 等分布式应用共资源池部署，高效通信和交换数据，实现高性能。
