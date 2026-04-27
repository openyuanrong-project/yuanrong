KV().GetWithParam
==================

.. cpp:function:: static inline std::vector<std::shared_ptr<YR::Buffer>> YR::KVManager::GetWithParam(const std::vector<std::string> &keys, const YR::GetParams &params, int timeout = DEFAULT_GET_TIMEOUT_SEC)

    检索与指定键关联的多个值，并支持基于偏移量的额外参数进行读取。

    此函数获取存储在指定键下的值，允许根据指定的偏移量和大小进行部分数据检索。如果任何键不存在或操作超时，除非将 `allowPartial` 设置为 `true`，否则将抛出异常。

    .. code-block:: cpp

       int main() {
           YR::Config conf;
           conf.mode = Config::Mode::CLUSTER_MODE;
           YR::Init(conf);
           std::string key = "kv-id-888";
           YR::KV().Del(key);
           std::string value = "kv-id-888wqdq";
           YR::KV().Set(key, value);
           YR::GetParam param = { .offset = 0, .size = 0 };
           YR::GetParams params;
           params.getParams.emplace_back(param);
           std::vector<std::shared_ptr<YR::Buffer>> res = YR::KV().GetWithParam({ key }, params);
           return 0;
       }

    参数：
        - **keys** - 要检索值的键列表。键的最大数量为 ``10000``。
        - **params** - 一个包含检索操作参数的结构体，每个键包括偏移量（offset）、大小（size）和追踪ID（traceId）。
        - **timeout** - 等待检索值的最大时间，以秒为单位。默认值为 ``300`` 秒。设置为 ``-1`` 表示无限期等待。
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
        std::vector<std::shared_ptr<YR::Buffer>>，一个包含检索到的数据的向量。结果的顺序与键的顺序相对应。失败的键将包含空指针。

参数结构补充说明如下：

.. cpp:struct:: GetParam
 
    指定查询键的参数。
    
    **公共成员**

    .. cpp:member:: uint64_t offset

        偏移量指定了要检索的数据的起始位置。

    .. cpp:member:: uint64_t size

        大小指定了要检索的元素数量或数据量。

.. cpp:struct:: GetParams

    指定一组查询键的参数。

    **公共成员**

    .. cpp:member:: std::vector<GetParam> getParams

        一个包含多个 `GetParam` 对象的向量。

    .. cpp:member:: std::string traceId

        一个用于跟踪和识别特定请求的追踪 ID。