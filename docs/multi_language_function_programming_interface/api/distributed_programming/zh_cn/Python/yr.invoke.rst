yr.invoke
=====================

.. py:function:: yr.invoke(*args, **kwargs) -> FunctionProxy

    定义一个无状态函数。

    该装饰器用于标记需要在 openYuanrong 系统上远程调用的函数，并返回该函数的代理对象。
	
    .. note::
        受限于 HTTP 客户端性能，当前单 client 仅支持 10 万并发。

    参数:
        - **func** (FunctionType) - 需要被远程调用的函数。
        - **invoke_options** (InvokeOptions_) - 调用参数，详见 `InvokeOptions`。
        - **return_nums** (int) - 函数返回值数量，限制：大于0，该参数不建议设置过大。

    返回:
        被装饰函数的代理对象。
        数据类型：FunctionProxy_ 。

    异常:
        - **RuntimeError** - 如果 `invoke` 无法装饰 `FunctionType` 以外的对象，会抛出此异常。

    样例：
        简单调用示例：
            >>> import yr
            >>> yr.init()
            >>> @yr.invoke
            ... def add(a, b):
            ...     return a + b
            >>> ret = add.invoke(1, 2)
            >>> print(yr.get(ret))
            >>> yr.finalize()

        函数内调用示例：
            >>> import yr
            >>> yr.init()
            >>> @yr.invoke
            ... def func1(a):
            ...     return a + " func1"
            >>> @yr.invoke
            ... def func2(a):
            ...     return yr.get(func1.invoke(a)) + " func2"
            >>> ret = func2.invoke("hello")
            >>> print(yr.get(ret))
            >>> yr.finalize()

.. _InvokeOptions: ../../zh_cn/Python/yr.InvokeOptions.html#yr.InvokeOptions
.. _FunctionProxy: ../../zh_cn/Python/yr.FunctionProxy.html#yr.FunctionProxy
