Stream
==================

.. cpp:function:: std::shared_ptr<Producer> YR::CreateProducer(const std::string &streamName, ProducerConf producerConf = {})

    创建一个生产者。

    .. code-block:: cpp

       std::string streamName = "streamName";
       // 创建流生产者
       YR::ProducerConf producerConf{};
       std::shared_ptr<YR::Producer> producer = YR::CreateProducer(streamName, producerConf);

    参数：
        - **streamName** - 流的名称。
        - **producerConf** - 生产者的配置信息。
    
    抛出：
        - :cpp:class:`Exception` - 4006: 不支持本地模式。
  
    返回：
        一个指向创建的生产者的指针。

参数结构补充说明如下：

.. cpp:struct:: ProducerConf
    
    **公共成员**

    .. cpp:member:: int64_t delayFlushTime = 5

        生产者在发送后会延迟指定的时间再触发刷新。

        <0：不自动刷新；0：立即刷新；其他值：延迟时间，以毫秒为单位。默认值：5。
    
    .. cpp:member:: int64_t pageSize = 1024 * 1024ul

        指定生产者的缓冲区页面大小，以字节（B）为单位。

        默认值：1 MB。必须大于 0 且是 4 KB 的倍数。

    .. cpp:member:: uint64_t maxStreamSize = 100 * 1024 * 1024ul
        
        指定流在工作进程中可以使用的最大共享内存大小，以字节（B）为单位。

        默认值：100 MB。范围：[64 KB, 工作进程共享内存大小]。

    .. cpp:member:: bool autoCleanup = false
        
        指定是否为流启用自动清理。

        默认值：false（禁用）。当最后一个生产者/消费者退出时，流将被自动清理。

    .. cpp:member:: bool encryptStream = false

        指定是否为流启用内容加密。

        默认值：false（禁用）。

    .. cpp:member:: uint64_t retainForNumConsumers = 0

        指定为多少个消费者保留生产者的数据。
        
        默认值：0。如果设置为 0，则在没有消费者时不会保留数据。
        此参数仅对第一个创建的消费者有效，当前有效范围为 [0, 1]。不支持多个消费者。在生产者之后创建的消费者可能无法接收到数据。

    .. cpp:member:: uint64_t reserveSize = 0

        指定预留内存的大小，以字节（B）为单位。

        创建生产者时，它会尝试预留 reserveSize 字节的内存。
        如果预留失败，将抛出异常。reserveSize 必须是 pageSize 的整数倍，并且在范围 [0, maxStreamSize] 内。
        如果 reserveSize 为 0，则默认设置为 pageSize。默认值：0。

    .. cpp:member:: std::unordered_map<std::string, std::string> extendConfig

        生产者的扩展配置。
        
        常见配置项包括：“STREAM_MODE”：流模式，可以是 “MPMC”、“MPSC” 或 “SPSC”。
        默认值：“MPMC”。如果指定了不支持的模式，将抛出异常。MPMC 表示多生产者多消费者，MPSC 表示多生产者单消费者，SPSC 表示单生产者单消费者。如果选择 MPSC 或 SPSC，数据系统将在内部启用多流共享页面功能。

    .. cpp:member:: std::string traceId

        自定义追踪 ID，用于故障排除和性能优化。

        仅在云环境中支持；在云环境之外的设置不会生效。
        最大长度：36。有效字符必须符合正则表达式： ``^[a-zA-Z0-9.-\/_!#%\^&*()+=\:;]*$``。

.. cpp:class:: StreamProducer : public YR::Producer

    **公共函数**

    .. cpp:function:: virtual void Send(const Element &element)

       向生产者发送数据。

       数据首先被放置在缓冲区中，然后根据配置的自动刷新策略（在一定时间间隔后或缓冲区满时）或通过手动调用 `Flush` 方法刷新，使数据对消费者可用。

       .. code-block:: cpp

          // 生产者发送数据
          std::string str = "hello";
          YR::Element element((uint8_t *)(str.c_str()), str.size());
          producer->Send(element);
        
       参数：
           - **element** - 要发送的元素数据。
       
       抛出：
           - :cpp:class:`Exception` - 4299: 发送元素失败。
  
    .. cpp:function:: virtual void Send(const Element &element, int64_t timeoutMs)

        向生产者发送数据。

        数据首先被放置在缓冲区中，然后根据配置的自动刷新策略（在一定时间间隔后或缓冲区满时）或通过手动调用 `Flush` 方法刷新，使数据对消费者可用。

        参数：
            - **element** - 要发送的元素数据。
            - **timeoutMs** - 可选的超时时间，以毫秒为单位。
       
        抛出：
            - :cpp:class:`Exception` - 4299: 发送元素失败。
  
    .. cpp:function:: virtual void Close()

        关闭生产者，触发缓冲区的自动刷新，并表明缓冲区将不再被使用。

        关闭后，生产者不能再被使用。

        .. code-block:: cpp

           producer->Close();
        
        抛出：
            - :cpp:class:`Exception` - 4299: 关闭生产者失败。
  
    .. cpp:function:: std::shared_ptr<Consumer> YR::Subscribe(const std::string &streamName, const SubscriptionConfig &config, bool autoAck = false)

        创建一个消费者

        .. code-block:: cpp
    
           YR::SubscriptionConfig config("subName", YR::SubscriptionType::STREAM);
           std::shared_ptr<YR::Consumer> consumer = YR::Subscribe(streamName, config);

        参数：
            - **streamName** - 流的名称。必须少于 256 个字符，且只能包含以下字符：``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``。
            - **config** - 消费者的配置信息。
            - **autoAck** - 如果 autoAck 为 true，则消费者会自动向数据系统发送确认消息（Ack），以确认已接收消息。默认值：false。
    
        抛出：
            - :cpp:class:`Exception` - 4006: 不支持本地模式。
        
        返回值：
            指向已创建消费者的指针。

参数结构补充说明如下：

.. cpp:struct:: Element
                
    **公共成员**
            
    .. cpp:member:: uint8_t *ptr
            
       指针指向数据。

    .. cpp:member:: uint64_t size

       数据的大小。

    .. cpp:member:: uint64_t id
    
       元素的ID。

.. cpp:struct:: SubscriptionConfig
                
    **公共函数**
            
    .. cpp:function:: inline SubscriptionConfig(std::string subName, const SubscriptionType subType)
        
       SubscriptionConfig 的构造函数。

       参数：
           - **subName** - 订阅名称。
           - **subType** - 订阅类型。
  
    **公共成员**

    .. cpp:member:: std::string subscriptionName
            
       订阅名称。
    
    .. cpp:member:: SubscriptionType subscriptionType

       订阅类型，包含三种类型：STREAM、ROUND_ROBIN 和 KEY_PARTITIONS。
       STREAM 表示订阅组中的单个消费者会消费流。
       ROUND_ROBIN 表示订阅组中的多个消费者会以轮询负载均衡的方式消费流。
       KEY_PARTITIONS 表示订阅组中的多个消费者会以基于键的分区负载均衡方式消费流。
       目前仅支持 STREAM 类型，其他类型暂不支持。默认订阅类型为 STREAM。
    
    .. cpp:member:: std::unordered_map<std::string, std::string> extendConfig
            
       SubscriptionConfig 的扩展配置。
    
    .. cpp:member:: std::string traceId

       用于故障排查和性能优化的自定义追踪 ID。

       仅在云端支持；在云端之外设置将不会生效。最大长度：36。有效字符必须符合正则表达式： ``^[a-zA-Z0-9~.-/_!@#%^&*+=:;]*$``。

.. cpp:class:: StreamConsumer : public YR::Consumer

    **公共函数**
                
    .. cpp:function:: virtual void Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements)

       通过订阅功能接收数据。

       消费者会等待接收预期数量的元素（expectNum）。当达到超时时间（timeoutMs）或接收到预期数量的元素时，调用将返回。

       .. code-block:: cpp
    
          // 消费者接收数据
          std::vector<YR::Element> elements;
          consumer->Receive(1, 6000, elements);  // 超时6秒
          consumer->Ack(elements[0].id);
          std::string actualData0(reinterpret_cast<char *>(elements[0].ptr), elements[0].size);
          std::cout << "receive: " << actualData0 << std::endl;

       参数：
           - **expectNum** - 预期接收的元素数量。
           - **timeoutMs** - 超时时间，单位为毫秒。
           - **outElements** - 实际接收到的元素。
  
       抛出：
           - :cpp:class:`Exception` - 抛出异常情形如下：

             - 3003：总大小超出 uint64_t 的最大值，或总大小超出限制。
             - 4299：未能按预期数量（expectNum）接收元素。
  
    .. cpp:function:: virtual void Receive(uint32_t timeoutMs, std::vector<Element> &outElements)

       通过订阅功能接收数据。

       消费者会等待接收预期数量的元素（expectNum）。当达到超时时间（timeoutMs）或接收到预期数量的元素时，调用将返回。

       参数：
           - **timeoutMs** - 超时时间，单位为毫秒。
           - **outElements** - 实际接收到的元素。

       抛出：
           - :cpp:class:`Exception` - 抛出异常情形如下：

             - 3003：总大小超出 uint64_t 的最大值，或总大小超出限制。
             - 4299：接收元素失败。
  
    .. cpp:function:: virtual void Ack(uint64_t elementId)

       这允许工作节点确定所有消费者是否已完成对该元素的消费，如果所有消费者都已确认，则可以触发内部内存回收机制。如果未确认，当消费者退出时，该元素将自动被确认。

       .. code-block:: cpp
    
          // 消费者接收数据
          std::vector<YR::Element> elements;
          consumer->Receive(1, 6000, elements);  // 超时6秒
          consumer->Ack(elements[0].id);
          std::string actualData0(reinterpret_cast<char *>(elements[0].ptr), elements[0].size);
          std::cout << "receive: " << actualData0 << std::endl;

       参数：
           - **elementId** - 要确认的元素的 ID。
    
       抛出：
           - :cpp:class:`Exception` - 4299：确认失败。

    .. cpp:function:: virtual void Close()

       关闭消费者。

       一旦关闭，消费者将无法再次使用。
   
       .. code-block:: cpp
       
          consumer->Close();
       
       抛出：
           - :cpp:class:`Exception` - 4299：关闭消费者失败。

.. cpp:function:: void YR::DeleteStream(const std::string &streamName)

    删除一个流。

    当流的全局生产者和消费者数量达到零时，该流将不再被使用，所有相关元数据（包括工作节点和主节点上的）都将被清理。此函数可以在任意主机节点上调用。

    .. code-block:: cpp
       
       // 删除流
       YR::DeleteStream(streamName);

    参数：
        - **streamName** - 流的名称。必须少于 256 个字符，且只能包含以下字符：``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``。
    
    抛出：
        - :cpp:class:`Exception` - 4006：不支持本地模式。