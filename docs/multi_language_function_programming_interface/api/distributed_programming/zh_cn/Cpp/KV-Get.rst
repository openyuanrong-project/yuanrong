KV().Get
==============

.. cpp:function:: static inline std::string YR::KVManager::Get(const std::string &key, int timeout = DEFAULT_GET_TIMEOUT_SEC)

    检索与指定键关联的值，类似于 Redis 的 `GET` 命令。

    此函数获取存储在指定键下的值。如果键不存在，将抛出异常。

    参数：
        - **key** - 与要检索的值关联的键。键不能为空。
        - **timeout** - 等待检索值的最大时间，以秒为单位。如果在该时间内未能检索到值，将抛出异常。默认值为 ``300`` 秒。超时时间可以设置为 ``-1`` 以无限期等待。
    
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
        std::string：与指定键关联的值。

.. cpp:function:: static inline std::vector<std::string>\
                  YR::KVManager::Get(const std::vector<std::string> &keys,\
                  int timeout = DEFAULT_GET_TIMEOUT_SEC, bool allowPartial = false)
    
    检索与指定键关联的多个值，类似于 Redis 的 `MGET` 命令。

    此函数获取存储在指定键下的值。如果任何键不存在或操作超时，除非将 `allowPartial` 设置为 ``true``，否则将抛出异常。

    .. code-block:: cpp

       int main() {
       YR::Config conf;
       conf.mode = Config::Mode::CLUSTER_MODE;
       YR::Init(conf);
       std::string value1{ "result" };
       std::string value2{ "result1" };
       YR::KV().Set("key", value1.c_str());
       std::string returnVal = YR::KV().Get("key"); // Retrieves "result"

       std::vector<std::string> keys{ "key1", "key2" };
       YR::KV().Set(keys[0], value1);
       YR::KV().Set(keys[1], value2);

       std::vector<std::string> returnVal = YR::KV().Get(keys); // Retrieves { "result", "result1" }
       return 0;
       }

    .. note::
        - 当 `allowPartial` = ``false`` 时：所有键必须成功检索；否则，将抛出异常。
        - 当 `allowPartial` = ``true`` 时：返回一个结果向量，其中失败的键对应空字符串。如果所有键都失败，将抛出异常。

    参数：
        - **keys** - 要检索值的键列表。键的最大数量为 ``10000``。
        - **timeout** - 等待检索值的最大时间，以秒为单位。如果在该时间内未能检索到值，将抛出异常。默认值为 ``300`` 秒。超时时间可以设置为 ``-1`` 以无限期等待。
        - **allowPartial** - 是否允许部分成功结果。如果为 ``true``，函数将返回一个结果向量，其中失败的键对应空字符串。如果为 ``false``，函数在任何键失败时将抛出异常。默认值为 ``false``。
   
    抛出：
        :cpp:class:`Exception` - 在以下情况下抛出：
            
        - 1001: 输入参数无效（例如，键为空、大小不匹配或包含无效字符）。
        - 4005：Get 操作失败（例如，键未找到或超时）。
        - 4201：RocksDB 错误（例如，磁盘问题）。
        - 4202：共享内存超出限制。
        - 4203：磁盘操作失败（例如，权限问题）。
        - 4204：磁盘空间已满。
        - 3002：内部通信错误。