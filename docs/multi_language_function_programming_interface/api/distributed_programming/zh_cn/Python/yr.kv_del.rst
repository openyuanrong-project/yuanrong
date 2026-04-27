yr.kv_del
=====================

.. py:function:: yr.kv_del(key: Union[str, List[str]]) -> None

    提供类 Redis 的 del 删除数据接口，支持同时删除多个数据。

    参数：
        - **key** (Union[str, List[str]]) – 删除数据指定的单个键或多个键。

    异常：
        - **RuntimeError** – 如果未初始化调用 `kv_del()`，会抛出异常。
        - **RuntimeError** – 如果从数据系统删除数据失败。

    样例：
        >>> yr.kv_write("kv-key", b"value1", yr.ExistenceOpt.NONE, yr.WriteMode.NONE_L2_CACHE, 0) # doctest: +SKIP
        >>> yr.kv_del("kv-key")