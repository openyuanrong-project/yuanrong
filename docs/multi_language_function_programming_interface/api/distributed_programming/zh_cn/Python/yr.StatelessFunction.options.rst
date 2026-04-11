.. _options_sf:

yr.StatelessFunction.options
----------------------------

.. py:method:: StatelessFunction.options(invoke_options: InvokeOptions) -> "StatelessFunction"

    动态修改被装饰函数的调用参数。

    .. note::
        该接口在本地模式下不生效。

    参数：
        - **invoke_options** (InvokeOptions_) - 调用选项。

    返回：
        StatelessFunction 对象本身。数据类型为 StatelessFunction。

    示例：
        >>> import yr
        >>>
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def add(a, b):
        ...    return a + b
        >>>
        >>> opt = yr.InvokeOptions(cpu=1000, memory=1024)
        >>> ret = add.options(opt).invoke(1, 2)
        >>> print(yr.get(ret))
        >>>
        >>> yr.finalize()

.. _InvokeOptions: ../../zh_cn/Python/yr.InvokeOptions.html#yr.InvokeOptions
