KV().Set
=========

支持存储二进制数据到数据系统，类似 Redis 的 SET 接口。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const char *value, ExistenceOpt existence = ExistenceOpt::NONE)

    设置键的值。

    .. code-block:: cpp

        std::string result{"result"};
        YR::KV().Set("key", result.c_str());  // 设置值

    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **value – [in]** 需要存储的二进制数据，限制最大 ``100M``。
        - **existence – [in]** 是否允许 Key 重复写入。默认值 ``YR::ExistenceOpt::NONE`` 表示允许，可选值 ``YR::ExistenceOpt::NX`` 表示不允许。

    
    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const char *value, size_t len, ExistenceOpt existence = ExistenceOpt::NONE)

    设置键的值。需要指定值的长度，用于值中包含空字符 ``\0`` 的场景。

    .. code-block:: cpp

        std::string result{"result"};
        auto size = result.size() + 1;
        YR::KV().Set("key", result.c_str(), size);  // 设置值的长度
    
    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **value – [in]** 需要存储的二进制数据，限制最大 ``100M``。
        - **len – [in]** 二进制数据的长度。用户需保证该值的正确性。
        - **existence – [in]** 是否允许 Key 重复写入。默认值 ``YR::ExistenceOpt::NONE`` 表示允许，可选值 ``YR::ExistenceOpt::NX`` 表示不允许。

    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const std::string &str, ExistenceOpt existence = ExistenceOpt::NONE)

    设置键的值。

    .. code-block:: cpp

        std::string result{"result"};
        YR::KV().Set("key", result);

    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **str – [in]** 需要存储的字符串。
        - **existence – [in]** 是否允许 Key 重复写入。默认值 ``YR::ExistenceOpt::NONE`` 表示允许，可选值 ``YR::ExistenceOpt::NX`` 表示不允许。

    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const char *value, SetParam setParam)

    设置键的值。

    .. code-block:: cpp

        YR::SetParam setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        std::string result{"result"};
        YR::KV().Set("key", result.c_str(), setParam);
    
    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **value – [in]** 需要存储的二进制数据，限制最大 ``100M``。
        - **setParam – [in]** 设置数据的可靠性级别等属性。

    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const char *value, size_t len, SetParam setParam) 

    设置键的值。需要指定值的长度，用于值中包含空字符 ``\0`` 的场景。

    .. code-block:: cpp

        YR::SetParam setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        std::string result{"result"};
        auto size = result.size() + 1;
        YR::KV().Set("key", result.c_str(), size, setParam);

    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **value – [in]** 需要存储的二进制数据，限制最大 ``100M``。
        - **len – [in]** 二进制数据的长度。用户需保证该值的正确性。
        - **setParam – [in]** 设置数据的可靠性级别等属性。

    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const std::string &str, SetParam setParam) 

    设置键的值。

    .. code-block:: cpp

        YR::SetParam setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        YR::KV().Set("kv-key", "kv-value", setParam);
    
    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **str – [in]** 需要存储的字符串。
        - **setParam – [in]** 设置数据的可靠性级别等属性。

    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const char *val, SetParamV2 setParamV2)

    设置键的值。

    .. code-block:: cpp

        YR::SetParamV2 setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        setParam.cacheType = YR::CacheType::MEMORY;
        std::string result{"result"};
        YR::KV().Set("key", result.c_str(), setParam);

    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **val – [in]** 需要存储的二进制数据，限制最大 ``100M``。
        - **setParamV2 – [in]** 设置数据的可靠性级别等属性。
    
    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const char *val, size_t len, SetParamV2 setParamV2)

    设置键的值。需要指定值的长度，用于值中包含空字符 ``\0`` 的场景。

    .. code-block:: cpp

        YR::SetParamV2 setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        setParam.cacheType = YR::CacheType::MEMORY;
        std::string result{"result"};
        auto size = result.size() + 1;
        YR::KV().Set("key", result.c_str(), size, setParam);

    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **val – [in]** 需要存储的二进制数据，限制最大 ``100M``。
        - **len – [in]** 二进制数据的长度。用户需保证该值的正确性。
        - **setParamV2 – [in]** 设置数据的可靠性级别等属性。

    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

.. cpp:function:: static inline void YR::KVManager::Set(const std::string &key, const std::string &str, SetParamV2 setParamV2)

    设置键的值。

    .. code-block:: cpp

        YR::SetParamV2 setParam;
        setParam.ttlSecond = 0;
        setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
        setParam.existence = YR::ExistenceOpt::NONE;
        setParam.cacheType = YR::CacheType::MEMORY;
        YR::KV().Set("kv-key", "kv-value", setParam);

    参数：
        - **key – [in]** 为存储的数据设置一个键，用于标识。参数不能为空，合法字符的正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **str – [in]** 需要存储的字符串。
        - **setParamV2 – [in]** 设置数据的可靠性级别等属性。

    抛出：
        :cpp:class:`Exception` -

        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。当参数 `existence` 设置为 ``YR::ExistenceOpt::NX`` 且该键已被使用时触发。
        - 包含错误信息的其他异常。

参数结构补充说明如下：

.. cpp:struct:: SetParam

    用于配置数据的可靠性等属性。

    **公共成员**

    .. cpp:member:: WriteMode writeMode = WriteMode::NONE_L2_CACHE

        写入模式

        设置数据的可靠性。服务端配置支持二级缓存比如 redis 服务时，使用该配置可以保证数据可靠性。默认值为 :cpp:enumerator:`YR::WriteMode::NONE_L2_CACHE`。
    
    .. cpp:member:: uint32_t ttlSecond = 0
        
        生存时间（TTL），单位为秒。

        指定数据在删除前保留的时间。默认值为 ``0``，表示该键将一直存在，直到使用 ``Del`` 接口显式删除。

    .. cpp:member:: ExistenceOpt existence = ExistenceOpt::NONE

        是否存在选项。

        用于表示是否允许键重复写入。默认值 ``YR::ExistenceOpt::NONE`` 表示允许，可选值 ``YR::ExistenceOpt::NX`` 表示不允许。

    .. cpp:member:: std::string traceId

        跟踪 ID。

        用于故障排除和性能优化的自定义跟踪ID。最大长度 ``36``，有效字符必须符合正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。

.. cpp:struct:: SetParamV2

    用于配置数据的可靠性等属性。

    **公共成员**

    .. cpp:member:: WriteMode writeMode = WriteMode::NONE_L2_CACHE

        写入模式

        设置数据的可靠性。服务端配置支持二级缓存比如 redis 服务时，使用该配置可以保证数据可靠性。默认值为 :cpp:enumerator:`YR::WriteMode::NONE_L2_CACHE`。
    
    .. cpp:member:: uint32_t ttlSecond = 0
        
        生存时间（TTL），单位为秒。

        指定数据在删除前保留的时间。默认值为 ``0``，表示该键将一直存在，直到使用 ``Del`` 接口显式删除。

    .. cpp:member:: ExistenceOpt existence = ExistenceOpt::NONE

        是否存在选项。

        用于表示是否允许键重复写入。默认值 ``YR::ExistenceOpt::NONE`` 表示允许，``YR::ExistenceOpt::NX`` 表示不允许。

    .. cpp:member:: std::string traceId

        跟踪 ID。

        用于故障排除和性能优化的自定义跟踪ID。最大长度 ``36``，有效字符必须符合正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
    
    .. cpp:member:: CacheType cacheType = CacheType::MEMORY

        缓存类型。

        指定数据缓存的类型。默认值 ``YR::CacheType::Memory`` 表示缓存到内存，``YR::CacheType::Disk`` 表示缓存到磁盘。

    .. cpp:member:: std::unordered_map<std::string, std::string> extendParams

        扩展参数。

        配置其他扩展参数。

.. cpp:enum:: YR::WriteMode : int

    写入模式。

    枚举值：

    .. cpp:enumerator:: NONE_L2_CACHE
    
        不写入二级缓存。

    .. cpp:enumerator:: WRITE_THROUGH_L2_CACHE
    
        同步写入二级缓存。

    .. cpp:enumerator:: WRITE_BACK_L2_CACHE
    
        异步写入二级缓存。

    .. cpp:enumerator:: NONE_L2_CACHE_EVICT
    
        不写入二级缓存。系统资源不足时，数据可能被清除。

.. cpp:enum:: YR::CacheType : int

    缓存介质类型。

    枚举值：

    .. cpp:enumerator:: MEMORY

        内存。
    
    .. cpp:enumerator:: DISK

        磁盘。

.. cpp:enum:: YR::ExistenceOpt : int

    键重复写入选项。

    枚举值：

    .. cpp:enumerator:: NONE

        允许。
    
    .. cpp:enumerator:: NX

        不允许。

