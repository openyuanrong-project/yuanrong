KV().Read
=============

.. cpp:function:: template<typename T>\
                  static inline std::shared_ptr<T> YR::KVManager::Read(const std::string &key, int timeout = DEFAULT_GET_TIMEOUT_SEC)

    检索键的值。

    .. code-block:: cpp

       int count = 100;
       Counter c1("Counter1-", count);
       auto result = YR::KV().Write(c1.name, c1);
       auto v1 = *(YR::KV().Read<Counter>(c1.name));  // 获取计数器的值
       Counter c2("Counter2-", count);
       result = YR::KV().Write(c2.name, c2);
       std::vector<std::string> keys{c1.name, c2.name};
       auto returnVal = YR::KV().Read<Counter>(keys);  // get std::vector<shared_ptr<Counter>>

    .. warning::
        当设置了超时时间时，`Read` 方法将等待 `Write` 方法完成，最长等待时间为指定的超时时间。`Read` 和 `Write` 方法调用的顺序没有限制。
        如果 `Write` 操作之后发生了异常（例如，数据服务工作进程重启），必须处理元数据残留，以确保 `Write` 操作完成；
        否则，对相同键调用 `Read` 方法将会立即抛出错误，而不会等待超时。

    模板参数：
        - **T** - 对象类型。

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

.. cpp:function:: template<typename T>\
                  static inline std::vector<std::shared_ptr<T>>\ 
                  YR::KVManager::Read(const std::vector<std::string> &keys, int timeout = DEFAULT_GET_TIMEOUT_SEC, bool allowPartial = false)

    检索键的值。

    .. code-block:: cpp
     
       int count = 100;
       Counter c1("Counter1-", count);
       auto result = YR::KV().Write(c1.name, c1);
       auto v1 = *(YR::KV().Read<Counter>(c1.name));  // 获取计数器的值
       Counter c2("Counter2-", count);
       result = YR::KV().Write(c2.name, c2);
       std::vector<std::string> keys{c1.name, c2.name};
       auto returnVal = YR::KV().Read<Counter>(keys);  // get std::vector<shared_ptr<Counter>>

    .. warning::
        - 当设置 `allowPartial` 时，必须明确提供 `timeout` 参数。
        - 当 `allowPartial` 为 `false` 时：所有键都必须成功检索才能返回结果；任何失败都会抛出异常。
        - 当 `allowPartial` 为 `true` 时：返回任何成功检索的键的结果，失败的键在其索引处将有空结果；如果所有键都失败，则抛出异常。
              
    参数：
        - **keys – [in]** 用于查询数据的键集合。最大限制：``10000``。
        - **timeout - [in]** 超时时间，以秒为单位，默认值为 ``300``。范围为 ``[0, INT_MAX)``。
        - **allowPartial - [in]** 确定是否允许部分成功结果。默认值为 ``false``。
    
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
        检索到的数据。如果键不存在，将抛出异常。