yr.create_function_group
=============================

.. py:function:: yr.create_function_group(wrapper: Union[function_proxy.FunctionProxy, instance_proxy.InstanceCreator],\
                                          args: tuple,\
                                          group_size: int,\
                                          options: FunctionGroupOptions \
                                         ) -> Union[List["yr.ObjectRef"], instance_proxy.FunctionGroupHandler]

    创建函数组实例。该函数用于创建函数组实例，支持函数代理或类实例创建器的调用。

    .. note::
        使用约束：

        1. 单个 group 最多成组创建 256 个实例。

        2. 单个 group 实例个数要求能被单个 bundle 的实例个数整除。

        3. 进程部署过程中，需要开启异构资源采集功能(默认开启)，需要选择不同模式采集不同范围参数，可设置参数 `npu_collection_mode`。

    参数：
        - **wrapper** (Union[function_proxy.FunctionProxy_, instance_proxy.InstanceCreator_]) - 被修饰的函数代理或被装饰类的创建器。
        - **args** (tuple) - 函数入参或者 class 的构造函数的位置参数。
        - **group_size** (int) - 函数组中的实例数。
        - **options** (FunctionGroupOptions_) - 创建函数组的选项。

    返回：
        返回数据对象的引用列表或函数组句柄。数据类型：Union[List[ObjectRef], FunctionGroupHandler]。

    异常：
        - **ValueError** - 如果 `FunctionGroupOptions` 或 `group_size` 参数填写错误，抛出此异常。
        - **RuntimeError** - 如果函数没有被 `@yr.invoke` 或 `@yr.instance` 包装，抛出此异常。

    样例：
        task 调用示例：
            >>> import yr
            >>>
            >>> yr.init()
            >>>
            >>> @yr.invoke
            ... def demo_func(name):
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

        class 调用示例：
            >>> import yr
            >>> @yr.instance
            ... class Demo(object):
            ...     name = ""
            >>>
            >>>     def __init__(self, name):
            ...         self.name = name
            >>>
            >>>     def return_name(self):
            ...         return self.name
            >>>
            >>> yr.init()
            >>> opts = yr.FunctionGroupOptions(
            ...         cpu=1000,
            ...         memory=1000,
            ...         resources={
            ...             "NPU/Ascend910B4/count": 1,
            ...         },
            ...         scheduling_affinity_each_bundle_size=2,
            ... )
            >>> name = "class_demo"
            >>> function_group_handler = yr.fcc.create_function_group(Demo, args=(name, ), group_size=8, options=opts)
            >>> objs = function_group_handler.return_name.invoke()
            >>> results = yr.get(objs)
            >>> assert results == [name for i in range(1, 9)]
            >>> function_group_handler.terminate()
            >>>
            >>> yr.finalize()

.. _FunctionGroupOptions: ../../zh_cn/Python/yr.FunctionGroupOptions.html#yr.FunctionGroupOptions
.. _function_proxy.FunctionProxy: ../../zh_cn/Python/yr.FunctionProxy.html#yr.FunctionProxy
.. _instance_proxy.InstanceCreator: ../../zh_cn/Python/yr.InstanceCreator.html#yr.InstanceCreator
