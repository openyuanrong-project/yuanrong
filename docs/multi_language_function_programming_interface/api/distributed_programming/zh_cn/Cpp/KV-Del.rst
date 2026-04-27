KV().Del
==================

.. cpp:function:: static inline void YR::KVManager::Del(const std::string &key, const DelParam &delParam = {})

    删除一个键及其关联的数据，类似于 Redis 的 `DEL` 命令。

    此函数移除指定的键及其关联的数据。如果键不存在，操作也被视为成功。

    参数：
        - **key** - 要删除的键。如果键不存在，操作也被视为成功。
        - **delParam** - 删除操作的可选参数，例如自定义追踪 ID。
    
    抛出：
        :cpp:class:`Exception` - 在集群模式下，如果删除操作失败，则会抛出异常。

.. cpp:function:: static inline std::vector<std::string>\
                  YR::KVManager::Del(const std::vector<std::string> &keys, const DelParam &delParam = {})

    删除多个键及其关联的数据，类似于 Redis 的 `DEL` 命令。

    此函数移除指定的键及其关联的数据。如果某个键不存在，对该键的操作也被视为成功。

    .. code-block:: cpp

       int main() {
           YR::Config conf;
           conf.mode = Config::Mode::CLUSTER_MODE;
           YR::Init(conf);
           std::vector<std::string> keys{ "key1", "key2" };
           std::vector<std::string> values{ "val1", "val2" };
           YR::KV().Set(keys[0], values[0]);
           YR::KV().Set(keys[1], values[1]);

           std::vector<std::string> failedKeys = YR::KV().Del(keys);
           if (!failedKeys.empty()) {
               std::cout << "Failed to delete the following keys: ";
               for (const auto& key : failedKeys) {
                   std::cout << key << " ";
               }
               std::cout << std::endl;
           } else {
               std::cout << "All keys deleted successfully." << std::endl;
           }
           return 0;
       }

    参数：
        - **key** - 要删除的键列表。键的最大数量为 ``10000``。如果某个键不存在，对该键的操作也被视为成功。
        - **delParam** - 删除操作的可选参数，例如自定义追踪 ID。
    抛出：
        :cpp:class:`Exception` - 在集群模式下，如果在删除操作期间连接到数据系统的操作超时，则会抛出异常。
    
    返回：
        std::vector<std::string>，一个包含未能成功删除的键的向量。如果所有键都成功删除，则该向量为空。

参数结构补充说明如下：

.. cpp:struct:: DelParam
 
    指定要删除的键的参数，例如设置追踪 ID。
    
    **公共成员**

    .. cpp:member:: std::string traceId

        traceId

        自定义追踪 ID，用于故障排除和性能优化；仅在云环境中支持；在云环境之外的设置不会生效。
        最大长度为 ``36``。有效字符必须符合正则表达式：``^[a-zA-Z0-9\~\.\-\/_!@#%\^\&\*\(\)\+\=\:;]*$``。
