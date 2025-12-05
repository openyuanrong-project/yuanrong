yr.get_function_group_context
===============================

.. py:function:: yr.get_function_group_context() -> FunctionGroupContext

    获取函数实例的函数组上下文。

    .. note::
        - 使用异构资源时才能获取到有效的函数组上下文。
        - 进程部署过程中，需要开启异构资源采集功能（默认开启），需要选择不同模式采集不同范围参数，可设置参数 `npu_collection_mode`。

    返回:
        函数组上下文。即 FunctionGroupContext_ 。

    样例：
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def demo_func(name):
        ...     context = yr.fcc.get_function_group_context()
        ...     print(context)
        ...     return name
        >>>
        >>> opts = yr.FunctionGroupOptions(
        ...     cpu=1000,
        ...     memory=1000,
        ...     resources={
        ...         "NPU/Ascend910B4/count": 1,
        ...     },
        ...     scheduling_affinity_each_bundle_size=2,
        ... )
        >>> name = "function_demo"
        >>> objs = yr.fcc.create_function_group(demo_func, args=(name,), group_size=8, options=opts)
        >>> rets = yr.get(objs)
        >>> assert rets == [name for i in range(1, 9)]
        >>> yr.finalize()

.. _FunctionGroupContext: ../../zh_cn/Python/yr.FunctionGroupContext.html#yr.FunctionGroupContext