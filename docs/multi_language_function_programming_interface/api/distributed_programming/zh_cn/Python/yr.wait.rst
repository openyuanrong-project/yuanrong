yr.wait
=====================

.. py:function:: yr.wait(obj_refs: Union[ObjectRef, List[ObjectRef]], wait_num: int = 1,\
                         timeout: Optional[int] = None) -> Tuple[List[ObjectRef], List[ObjectRef]]

    根据对象的 `key` 等待数据系统中对象的值准备好。接口调用后会阻塞，直到对象的值计算完成。
	
    .. note::
        每次返回的结果可能不同，因为 `invoke` 完成先后顺序不确定。

    参数:
        - **obj_refs** (list) - 保存到数据系统的数据。
        - **wait_num** (int，可选) - 至少需要等待的数量。默认为 ``1``。取值需不大于 len(object_refs)。
        - **timeout** (int，可选) - 超时时间，单位：秒。需注意，如果取默认值 ``None``，则会一直等待，实际最大等待时间受限于 `get` 中的等待因素。

    返回:
        返回两个列表：第一个列表包含已经完成的请求；第二个列表包含未完成的请求。
        数据类型：tuple[list, list]。

    异常:
        - **TypeError** - 如果参数类型传入错误。
        - **ValueError** - 如果参数传入错误。

    样例：
        >>> import yr
        >>> import time
        >>>
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def demo(a):
        ...     time.sleep(a)
        ...     return "sleep:", a
        ...
        >>> res = [demo.invoke(i) for i in range(4)]
        >>>
        >>> wait_num = 3
        >>> timeout = 10
        >>> result = yr.wait(res, wait_num, timeout)
        >>> print("ready_list  = ", result[0], "unready_list = ", result[1])
        ready_list  =  [ObjectRef(...), ObjectRef(...)] unready_list =  [ObjectRef(...)]
        >>> print(yr.get(result[0]))
        [('sleep:', 0), ('sleep:', 1)]
        >>>
        >>> yr.finalize()