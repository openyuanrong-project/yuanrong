KV().MWriteTx
==============

.. cpp:function:: template<typename T>\
                  static inline void YR::KVManager::MWriteTx(const std::vector<std::string> &keys, const std::vector<T> &vals, ExistenceOpt existence)

    设置多个键值对。

    值将通过 msgpack 进行序列化。对键值对的所有操作需完全成功或完全失败。

    .. code-block:: cpp

       int count = 100;
       Counter c("Counter1-", count);
       std::vector<std::string> keys = {c.name};
       std::vector<Counter> vals = {c};
       try {
           YR::KV().MWriteTx<Counter>(keys, vals, YR::ExistenceOpt::NX);
       } catch (YR::Exception &e) {
           // 处理异常...
       }

    .. note::
        最大调用频率限制为每秒250次调用。

    模板参数：
        - **T** - 对象的类型。

    参数：
        - **keys – [in]** 用于标识存储数据的键的数组。数组的最大长度为 ``8``。每个元素需符合正则表达式：``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``。
        - **vals – [in]** 要在数据系统中存储的数据数组。此数组中数据的位置需与 keys 数组中键的位置相对应。此数组的长度需与 keys 数组匹配。
        - **existence – [in]** 需设置为 ``YR::ExistenceOpt::NX``。
    
    抛出：
        :cpp:class:`Exception` - 
        
        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。如果将参数 `existence` 设置为 ``YR::ExistenceOpt::NX``，并且列表中的一个或多个键已被先前设置或写入。
        - 在执行 Set 操作期间，可能会因其他错误而抛出异常，错误消息中会提供详细描述。

.. cpp:function:: template<typename T>\
                  static inline void YR::KVManager::MWriteTx(const std::vector<std::string> &keys, const std::vector<T> &vals, const MSetParam &mSetParam)
    
    设置多个键值对。

    值将通过 msgpack 进行序列化。对键值对的所有操作需完全成功或完全失败。

    .. code-block:: cpp

       int count = 100;
       Counter c("Counter1-", count);
       std::vector<std::string> keys = {c.name};
       std::vector<Counter> vals = {c};
       YR::MSetParam param;
       param.ttlSecond = 0;
       param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
       param.existence = YR::ExistenceOpt::NONE;
       param.cacheType = YR::CacheType::MEMORY;
       try {
           YR::KV().MWriteTx<Counter>(keys, vals, param);
       } catch (YR::Exception &e) {
           // 处理异常...
       }

    .. note::
        最大调用频率限制为每秒250次调用。

    模板参数：
        - **T** - 对象的类型。

    参数：
        - **keys – [in]** 用于标识存储数据的键的数组。数组的最大长度为 ``8``。每个元素必需符合正则表达式：``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``。
        - **vals – [in]** 要在数据系统中存储的数据数组。此数组中数据的位置需与 keys 数组中键的位置相对应。此数组的长度需与 keys 数组匹配。
        - **mSetParam – [in]** 为对象配置属性，例如可靠性。
   
    抛出：
        :cpp:class:`Exception` - 
        
        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。如果将参数 `existence` 设置为 ``YR::ExistenceOpt::NX``，并且列表中的一个或多个键已被先前设置或写入。
        - 在执行 Set 操作期间，可能会因其他错误而抛出异常，错误消息中会提供详细描述。
