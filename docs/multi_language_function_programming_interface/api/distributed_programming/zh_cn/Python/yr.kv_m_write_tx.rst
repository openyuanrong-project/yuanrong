yr.kv_m_write_tx
==================

.. py:function:: yr.kv_m_write_tx(keys: ~typing.List[str], values: typing.List[bytes], m_set_param: yr.runtime.MSetParam = MSetParam(existence=<ExistenceOpt.NX: 1>, write_mode=<WriteMode.NONE_L2_CACHE: 0>, ttl_second=0, cache_type=<CacheType.MEMORY: 0>)) -> None

    提供一个类似 Redis 的集合存储接口，支持将一组二进制数据保存到数据系统。

    参数:
        - **key** (List[str]) - 为保存的数据设置一组键来标识数据。使用此键查询数据时，不能为空。
        - **values** (List[bytes]) - 需要存储的一组二进制数据。云外最大存储限制为 ``100`` M。
        - **m_set_param** (MSetParam，可选) 多千伏配置参数写入数据系统。包括 ``existence``、``write_mode``、``ttl_second`` 和 ``cache_type``。

    返回:
        无。

    异常:
        - **RuntimeError** - 如果 `kv_m_write_tx` 没有初始化和调用，将抛出异常。向数据系统写入数据失败。

    样例：
        >>> import yr
        >>> yr.init()
        >>> mset_param = yr.MSetParam()
        >>> mset_param.existence = yr.ExistenceOpt.NX
        >>> mset_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
        >>> mset_param.ttl_second = 100
        >>> mset_param.cache_type = yr.CacheType.DISK
        >>> yr.kv_m_write_tx(["key1", "key2"], [b"value1", b"value2"], mset_param)
        >>>
        >>> yr.finalize()