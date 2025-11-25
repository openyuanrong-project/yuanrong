# 监控与调试

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   dashboard
   logs
   metrics/index
```

可观察性是指用户通过各种外部输出（例如日志、指标、事件等）观察和推断 openYuanrong 应用程序和 openYuanrong 集群的内部状态的能力。openYuanrong 通过提供可观测性能力帮助用户轻松监控和调试 openYuanrong 应用程序和集群。

监控和调试 openYuanrong 应用程序包括以下步骤：

1. 通过 Dashboard（即将开源）、日志、指标监控集群和应用程序，识别出现的问题或错误。
2. 使用远程调试器（即将开源）等工具和数据进行调试。
3. 判断问题点并修复，测试验证结果。

本节将展开介绍 openYuanrong 提供的相关工具和数据。

- [日志](./logs.md)
- [指标](./metrics/index.md)
- [Dashboard](./dashboard.md)