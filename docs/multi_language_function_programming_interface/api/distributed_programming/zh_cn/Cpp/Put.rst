Put
====

将数据保存到后端，不支持 Put 裸指针。

.. cpp:function:: template<typename T> ObjectRef<T> YR::Put(const T &val)

    存储对象到数据系统​​。

    .. code-block:: cpp

       auto objRef = YR::Put(100);
       auto value = *(YR::Get(objRef));
       assert(value == 100);  // 预期结果为 100

    模板参数：
        - **T** - 模板类型。

    参数：
        - **val** - 待处理对象。
    
    抛出：
        :cpp:class:`Exception` - 如果 openYuanrong 未初始化时抛出。
  
    返回：
        ObjectRef<T>，对象的引用。

.. cpp:function:: template<typename T> ObjectRef<T> YR::Put(const T &val, const CreateParam &createParam)

    存储对象到数据系统​​。

    .. code-block:: cpp

       YR::CreateParam param;
       param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
       param.consistencyType = YR::ConsistencyType::PRAM;
       param.cacheType = YR::CacheType::DISK;
       auto objRef = YR::Put(100, param);
       auto value = *(YR::Get(objRef));
       assert(value == 100);  // 预期结果为 100 

    模板参数：
       - **T** - 模板类型。

    参数：
       - **val** - 待处理对象。
       - **createParam** - 创建参数，详见下方参数结构补充说明。
   
    抛出：
        :cpp:class:`Exception` - 如果 openYuanrong 未初始化时抛出。
 
    返回：
       ObjectRef<T>，对象的引用。
    
参数结构补充说明如下：

.. cpp:struct:: CreateParam

    用于配置对象的可靠性等属性。

    **公共成员**
 
    .. cpp:member:: WriteMode writeMode = WriteMode::NONE_L2_CACHE
 
       写入模式
 
       设置数据的可靠性。当服务端配置支持二级缓存时用于保证可靠性时，比如 redis 服务，那么通过这个配置可以保证数据可靠性。默认值为 :cpp:enumerator:`YR::WriteMode::NONE_L2_CACHE`。
 
    .. cpp:member:: ConsistencyType consistencyType = ConsistencyType::PRAM
 
       一致性类型
 
       数据一致性配置。分布式场景可以配置不同级别的一致性语义。可选参数为：
 
       * :cpp:enumerator:`YR::ConsistencyType::PRAM` (异步)
       * :cpp:enumerator:`YR::ConsistencyType::CAUSAL` (因果一致性)
 
       默认值为 :cpp:enumerator:`YR::ConsistencyType::PRAM`。
 
    .. cpp:member:: CacheType cacheType = CacheType::MEMORY
 
       缓存类型
 
       用于标识分配的是内存介质还是磁盘介质。可选参数为：
 
       * :cpp:enumerator:`YR::CacheType::MEMORY` (内存介质)
       * :cpp:enumerator:`YR::CacheType::DISK` (磁盘介质)
 
       默认值为 :cpp:enumerator:`YR::CacheType::MEMORY`。

.. cpp:enum:: YR::ConsistencyType : int

    数据一致性配置

    分布式场景可以配置不同级别的一致性语义

    枚举值：

    .. cpp:enumerator:: PRAM

    异步

    .. cpp:enumerator:: CAUSAL

    因果一致性