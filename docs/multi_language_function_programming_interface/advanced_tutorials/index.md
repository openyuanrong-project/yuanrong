# 进阶教程

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   yr_shutdown
   generator
   buffer_get
   avoid_excessive_concurrency
   use_nested_call
   use_wait
   use_InvokeOptions_limit_concurrent_num
```

本节介绍如何使用 openYuanrong 的高阶特性以及常用的设计模式。

- [自定义优雅退出](./yr_shutdown.md)
- [openYuanrong 生成器](./generator.md)
- [接口免序列化与反序列化](./buffer_get.md)
- [避免过度并发](./avoid_excessive_concurrency.md)
- [嵌套调用](./use_nested_call.md)
- [通过 yr.wait 机制实现反压控制，限制并发/待处理任务的总数](./use_wait.md)
- [使用资源用量限制任务并发数量](./use_InvokeOptions_limit_concurrent_num.md)
