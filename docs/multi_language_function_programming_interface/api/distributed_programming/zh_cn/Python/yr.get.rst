yr.get
=====================

.. py:function:: yr.get(obj_refs: Union[ObjectRef, List], timeout: int = constants.DEFAULT_GET_TIMEOUT,\
                        allow_partial: bool = False) -> object

    根据数据对象的键从数据系统中检索值。接口调用后会阻塞直到获取到对象的值或者超时。

    .. note::
        get统一返回 memoryview 指针。
	
    参数:
        - **obj_refs** (ObjectRef_，List[ObjectRef]) - 数据系统中的对象的 `object_ref` 。
        - **timeout** (int，可选) - 超时值，取 ``-1`` 时默认无限等待。取值限制： ``-1``， ``(0, ∞)``，默认值为 ``constants.DEFAULT_GET_TIMEOUT`` 秒 即 ``300``。
        - **allow_partial** (bool，可选) - 如果设置为 ``False``，当数据系统在超时期间返回部分结果时，get 接口将抛出异常。
          如果设置为 ``True``，当数据系统返回部分结果时，get 接口将返回对象列表，并将失败的对象填充为 ``None``。默认值为``_DEFAULT_ALLOW_PARTIAL``。

    返回:
        一个 Python 对象或一组 Python 对象。

    异常:
        - **ValueError** - 传入参数类型错误。
        - **RuntimeError** - 从数据系统中获取对象失败。
        - **YRInvokeError** - 函数执行错误。（该异常详情：YRInvokeError_）。
        - **TimeoutError** - 在指定的超时时间内无法获取所有对象引用的结果。

    样例：
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.invoke()
        >>> def add(a, b):
        ...     return a + b
        >>> obj_ref_1 = add.invoke(1, 2)
        >>> obj_ref_2 = add.invoke(3, 4)
        >>> result = yr.get([obj_ref_1, obj_ref_2], timeout=-1)
        >>> print(result)
        >>> yr.finalize()

.. _ObjectRef: ../../zh_cn/Python/yr.object_ref.ObjectRef.html#yr.object_ref.ObjectRef
.. _YRInvokeError: ../../zh_cn/Python/yr.exception.YRInvokeError.html#yr.exception.YRInvokeError
