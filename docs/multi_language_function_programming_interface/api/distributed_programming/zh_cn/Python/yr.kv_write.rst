yr.kv_write
=====================

.. py:function:: yr.kv_write(key: str, value: bytes, existence: ExistenceOpt = ExistenceOpt.NONE,\
                             write_mode: WriteMode = WriteMode.NONE_L2_CACHE, ttl_second: int = constants.DEFAULT_NO_TTL_LIMIT,\
                             cache_type: CacheType = CacheType.MEMORY) -> None

    提供类 Redis 的 set 存储接口，支持保存二进制数据到数据系统。

    参数:
        - **key** (str) - 为保存的数据设置一个键，用于标识该数据。查询数据时使用该键进行查询，不能为空。
        - **value** (bytes) - 需要存储的二进制数据。云外限制最大存储 ``100M``。
        - **existence** (ExistenceOpt，可选) - 是否支持 `Key` 重复写入。
          该参数可选，默认为 ``ExistenceOpt.NONE``，表示支持；``ExistenceOpt.NX`` 表示不支持。
        - **write_mode** (WriteMode，可选) - 设置数据的可靠性。
          当服务端配置支持二级缓存时用于保证可靠性时（比如 Redis 服务），
          可以通过该配置保证数据可靠性。该参数可选，默认为 ``WriteMode.NONE_L2_CACHE``。
        - **ttl_second** (int，可选) - 数据生命周期，超过会被删除。
          该参数可选，默认为 ``0``，表示 `key` 会一直存在，直到显式调用 `kv_del` 接口。
        - **cache_type** (CacheType，可选) - 设置数据存储的介质。
          该参数可选，默认为 ``CacheType.MEMORY`` ，表示存储至内存；``CacheType.DISK`` 表示存储至硬盘。

    返回：
        None。

    异常:
        - **RuntimeError** - 如果未初始化调用 `kv_write`，会抛出异常。
        - **RuntimeError** - 如果数据写入数据系统失败。

    样例：
        >>> import yr
        >>> yr.init()
        >>>
        >>> yr.kv_write("kv-key", b"value1", yr.ExistenceOpt.NONE, yr.WriteMode.NONE_L2_CACHE, 0, yr.CacheType.MEMORY)
        >>>
        >>> yr.finalize()
