yr.kv_get
=====================

.. py:function:: yr.kv_get(key: Union[str, List[str]], timeout: int = 300) \
                           -> Union[Union[bytes, memoryview], List[Union[bytes, memoryview]]]

    提供类 Redis 的 get 获取数据接口，支持同时获取多个数据。

    参数：
        - **key** (Union[str, List[str]]) – 查询数据指定的单个键或多个键。
        - **timeout** (int，可选) – 单位为秒，默认为 ``300``，为 ``-1`` 时表示永久阻塞等待。

    返回：
        返回查到的一个或一组数据。
        数据类型：Union[bytes, List[bytes]]。

    异常：
        - **RuntimeError** – 如果未初始化调用 `kv_get()`，会抛出异常。
        - **RuntimeError** – 如果从数据系统获取数据失败，会抛出异常。

    样例：
        >>> v1 = yr.kv_get("kv-key")