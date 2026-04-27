yr.cancel
=====================

.. py:function:: yr.cancel(obj_refs: Union[ObjectRef, List[ObjectRef]], allow_force: bool = True,\
                           allow_recursive: bool = False) -> None

    取消无状态函数调用。
	
    .. note::
        当前传入 instance 请求不会执行对应的 cancel 操作，且不会报错。

    参数:
        - **obj_refs** (Union[ObjectRef，List[ObjectRef]]) - 单个或多个需要的停止的无状态请求对应的 ObjectRef。
        - **allow_force** (bool，可选) - 强制停止。默认值为 ``_DEFAULT_ALLOW_FORCE``。
        - **allow_recursive** (bool，可选) - 允许递归。默认值为 ``_DEFAULT_ALLOW_RECURSIVE``。

    返回:
        None。

    异常:
        - **TypeError** - 如果参数类型有误。

    样例：
        >>> import time
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.invoke
        >>> def func():
        >>>     time.sleep(100)
        >>>
        >>> ret = func.invoke()
        >>> yr.cancel(ret)
        >>> yr.finalize()