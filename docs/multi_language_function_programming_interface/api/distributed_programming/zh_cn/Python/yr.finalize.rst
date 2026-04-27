yr.finalize
=====================

.. py:function:: yr.finalize() -> None

    关闭客户端，释放函数实例等创建的资源。

    .. note::
        1. 该接口用于关闭 openYuanrong 客户端，释放已创建的函数实例、未释放的数据对象等资源。
        2. ``finalize`` 必须在 ``init`` 之后调用。

    异常：
        - **RuntimeError** - 如果未初始化 ``yr`` 时调用 ``finalize``，会抛出此异常。

    返回：
        None。

    样例：
        >>> import yr
        >>> conf = yr.Config()
        >>> yr.init(conf)
        >>> yr.finalize()

