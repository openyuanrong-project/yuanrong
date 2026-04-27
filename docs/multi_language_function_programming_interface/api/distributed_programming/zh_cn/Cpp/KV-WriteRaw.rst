KV().WriteRaw
==============

.. cpp:function:: static inline void YR::KVManager::WriteRaw\
                  (const std::string &key, const char *value)

    写入键的值（原始字节）。

    参数：
        - **key – [in]** 用于标识存储数据的键。不能为空，且有效字符必须符合正则表达式：``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。
        - **value - [in]** 要在数据系统中存储的原始二进制数据。
    
    抛出：
        :cpp:class:`Exception` - 在以下情况下抛出：
        
        - 1001：参数错误。提供详细的错误信息。
        - 4206：键已存在。如果将参数 `existence` 设置为 ``YR::ExistenceOpt::NX``，并且列表中的一个或多个键已被先前设置或写入。
        - 在执行 Set 操作期间，可能会因其他错误而抛出异常，错误消息中会提供详细描述。
