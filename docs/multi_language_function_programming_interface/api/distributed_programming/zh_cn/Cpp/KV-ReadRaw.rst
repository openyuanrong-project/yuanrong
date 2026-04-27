KV().ReadRaw
=============

.. cpp:function:: static inline const void *YR::KVManager::ReadRaw\
                  (const std::string &key, int timeout = DEFAULT_GET_TIMEOUT_SEC)

    检索由 `WriteRaw` 写入的键的值。

    .. warning::
        当配置了超时时，`Read` 方法将等待 `Write` 方法完成，最长等待时间为超时时间。
        `Read` 和 `Write` 方法调用的顺序没有限制。如果 `Write` 操作之后发生了异常（例如 ds worker 重启），
        必须处理元数据残留，以确保 `Write` 操作完成；否则，对相同键调用 `Read` 将会立即抛出错误，而不会等待超时。

    参数：
        - **keys – [in]** 用于查询数据的单个键。
        - **timeout - [in]** 超时时间，以秒为单位，默认值为 ``300``。范围为 ``[0, INT_MAX/1000)``。``-1`` 表示永久阻塞等待。
    
    抛出：
        :cpp:class:`Exception` - 在以下情况下抛出：
        
        - 1001: 输入参数无效（例如，键为空、大小不匹配或包含无效字符）。
        - 4005：Get 操作失败（例如，键未找到或超时）。
        - 4201：RocksDB 错误（例如，磁盘问题）。
        - 4202：共享内存超出限制。
        - 4203：磁盘操作失败（例如，权限问题）。
        - 4204：磁盘空间已满。
        - 3002：内部通信错误。

    返回：
        检索到的数据。