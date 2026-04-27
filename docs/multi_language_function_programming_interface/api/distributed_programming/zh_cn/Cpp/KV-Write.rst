KV().Write
=============

.. cpp:function:: template<typename T>\
                  static inline void YR::KVManager::Write\
                  (const std::string &key, const T &value, ExistenceOpt existence = ExistenceOpt::NONE)

    写入一个键的值。

    .. code-block:: cpp

       int count = 100;
       Counter c("Counter1-", count);
       try {
           YR::KV().Write<Counter>(c.name, c);
       } catch (YR::Exception &e) {
           // 处理异常...
       }

    模板参数：
        - **T** - 对象的类型。

    参数：
        - **keys – [in]** 用于标识存储数据的键。不能为空，且有效字符必须符合正则表达式：``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``。
        - **value - [in]** 数据系统中存储的数据。
        - **existence - [in]** 确定键是否允许重复写入。默认值 ``YR::ExistenceOpt::NONE`` 表示允许，可选值 ``YR::ExistenceOpt::NX`` 表示不允许。
    
    抛出：
        :cpp:class:`Exception` - 
        
        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。如果将参数 `existence` 设置为 ``YR::ExistenceOpt::NX``，并且列表中的一个或多个键已被先前设置或写入。
        - 在执行 Set 操作期间，可能会因其他错误而抛出异常，错误消息中会提供详细描述。

.. cpp:function:: template<typename T>\
                  static inline void YR::KVManager::Write(const std::string &key, const T &value, SetParam setParam)
    
    写入一个键的值。

    .. code-block:: cpp

       int count = 100;
       Counter c("Counter1-", count);
       YR::SetParam setParam;
       setParam.ttlSecond = 0;
       setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
       setParam.existence = YR::ExistenceOpt::NONE;
       try {
           YR::KV().Write<Counter>(c.name, c, setParam);
       } catch (YR::Exception &e) {
           // 处理异常...
       }

    模板参数：
        - **T** - 对象的类型。

    参数：
        - **Key - [in]** 用于标识存储数据的键。不能为空，且有效字符必须符合正则表达式：``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``。
        - **value - [in]** 数据系统中存储的数据。
        - **setParam - [in]** 配置对象的属性，例如可靠性。
   
    抛出：
        :cpp:class:`Exception` - 
        
        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。如果将参数 `existence` 设置为 ``YR::ExistenceOpt::NX``，并且列表中的一个或多个键已被先前设置或写入。
        - 在执行 Set 操作期间，可能会因其他错误而抛出异常，错误消息中会提供详细描述。

.. cpp:function:: template<typename T>\
                  static inline void YR::KVManager::Write(const std::string &key, const T &value, SetParamV2 setParam)

    写入一个键的值。

    .. code-block:: cpp

       int count = 100;
       Counter c("Counter1-", count);
       YR::SetParamV2 setParam;
       setParam.ttlSecond = 0;
       setParam.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
       setParam.existence = YR::ExistenceOpt::NONE;
       setParam.cacheType = YR::CacheType::MEMORY;
       try {
           YR::KV().Write<Counter>(c.name, c, setParam);
       } catch (YR::Exception &e) {
           // 处理异常...
       }
    
    模板参数：
       - **T** - 对象的类型。

    参数：
        - **Key - [in]** 用于标识存储数据的键。不能为空，且有效字符必须符合正则表达式：``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``。
        - **value - [in]** 数据系统中存储的数据。
        - **setParamV2 - [in]** 配置对象的属性，例如可靠性。
  
    抛出：
        :cpp:class:`Exception` - 
       
        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。如果将参数 `existence` 设置为 ``YR::ExistenceOpt::NX``，并且列表中的一个或多个键已被先前设置或写入。
        - 在执行 Set 操作期间，可能会因其他错误而抛出异常，错误消息中会提供详细描述。