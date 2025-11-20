yr.fnruntime.Producer
=======================

.. py:class:: yr.fnruntime.Producer

    基类：``object`` 
    
    Producer接口类。
	
    样例：
        >>> try:
        ...     producer_config = ProducerConfig(
        ...         delay_flush_time=5,
        ...         auto_clean_up=True,
        ...     )
        ...     producer = yr.create_stream_producer("streamName", producer_config)
        ...     # .......
        ...     data = b"hello"
        ...     element = Element(data=data, id=0)
        ...     producer.send(element)
        ...     producer.flush()
        ...     producer.close()
        ... except RuntimeError as exp:
        ...     # .......
        ...     pass
    
    **方法**：

    +------------------------+------------------------------------------------------------------+
    | :ref:`__init__ <init>` | 初始化 Producer 实例。                                           |
    +------------------------+------------------------------------------------------------------+
    | :ref:`close <close>`   | 关闭生产者将触发数据缓冲区的自动刷新，并表明数据缓冲区不再使用。 |
    +------------------------+------------------------------------------------------------------+
    | :ref:`flush <flush>`   | 手动刷缓冲数据使得消费者可见。                                   |
    +------------------------+------------------------------------------------------------------+
    | :ref:`send <send>`     | 生产者发送数据时，会先将数据存入缓冲区。                         |
    +------------------------+------------------------------------------------------------------+

.. toctree::
    :maxdepth: 1
    :hidden:
    
    yr.fnruntime.Producer.__init__
    yr.fnruntime.Producer.close
    yr.fnruntime.Producer.flush
    yr.fnruntime.Producer.send
