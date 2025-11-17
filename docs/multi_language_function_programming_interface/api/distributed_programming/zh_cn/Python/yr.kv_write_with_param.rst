yr.kv_write_with_param
=========================

.. py:function:: kv_write_with_param(key: str, value: bytes, set_param: SetParam) -> None

    提供支持自定义参数配置的存储接口。用于以更灵活的方式将二进制数据写入数据系统。

    参数:
        - **key** (str) - 为保存的数据设置一个键，用于标识该数据。查询数据时使用该键进行查询，不能为空。
        - **value** (bytes) - 需要存储的二进制数据。云外限制最大存储 100M。
        - **set_param** (SetParam_) - 此接口在数据系统中的配置参数，
          包括 `existence`、`write_mode`、`ttl_second` 和 `cache_type`。

    异常:
        - **RuntimeError** - 如果未初始化调用 `kv_write_with_param`，会抛出异常。
        - **RuntimeError** - 如果数据写入数据系统失败。

    返回：
        None。

    样例：
        >>> import yr
        >>> yr.init()
        >>> # worker启动参数里面要配置shared_disk_directory， shared_disk_size_mb这两个参数,
        >>> # 否则这个例子会出错
        >>> set_param = yr.SetParam()
        >>> set_param.existence = yr.ExistenceOpt.NX
        >>> set_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
        >>> set_param.ttl_second = 10
        >>> set_param.cache_type = yr.CacheType.DISK
        >>> yr.kv_write_with_param("kv-key", b"value1", set_param)
        >>>
        >>> yr.finalize()

.. _SetParam: ../../Python/generated/yr.kv_write_with_param.html