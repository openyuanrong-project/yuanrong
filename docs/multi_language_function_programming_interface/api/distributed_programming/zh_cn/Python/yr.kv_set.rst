yr.kv_set
=====================

.. py:function:: kv_set(key: str, value: bytes, set_param: SetParam = SetParam()) -> None

    提供类似 Redis 的集合存储接口，支持二进制数据保存到数据系统。

    参数:
        - **key** (str) - 为保存的数据设置一个键来标识它。使用此键查询数据。它不能是空的。
        - **value** (bytes) - 要存储的二进制数据。云外的最大存储限制为 ``100M``。
        - **set_param** (SetParam，可选) - 数据系统中写入的kv配置参数包括 ``existence``，``write_mode``， ``ttl_second`` 和 ``cache_type``。

    返回:
        None。

    异常:
        - **RuntimeError** - 如果未初始化并调用 ``kv_set``，将抛出异常。
        - **RuntimeError** - 向数据系统写入数据失败。

    样例：
        >>> import yr
        >>> yr.init()
        >>>
        >>> set_param = yr.SetParam()
        >>> set_param.existence = yr.ExistenceOpt.NX
        >>> set_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
        >>> set_param.ttl_second = 10
        >>> yr.kv_set("kv-key", b"value1", set_param)
        >>>
        >>> yr.finalize()