yr.create_stream_consumer
==========================

.. py:function:: yr.create_stream_producer(stream_name: str, config: SubscriptionConfig) -> Consumer

    创建消费者。
	
    参数:
        - **stream_name** (str) - 流的名称。长度必须小于 256 个字符且仅含有下列字符 `(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)` 。
        - **config** (SubscriptionConfig_) - 消费者的配置信息。

    返回:
        消费者。即 Consumer_ 。

    异常:
        - **RuntimeError** - 如果创建 Consumer 失败。

    样例：
        >>> try:
        ...     config = SubscriptionConfig("subName", SubscriptionType.STREAM)
        ...     consumer = create_stream_consumer("streamName", config)
        ... except RuntimeError as exp:
        ...     pass

.. _SubscriptionConfig: ../../Python/generated/yr.SubscriptionConfig.html#yr.SubscriptionConfig
.. _Consumer: ../../Python/generated/yr.fnruntime.Consumer.html#yr.fnruntime.Consumer