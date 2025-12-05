yr.create_stream_producer
==========================

.. py:function:: yr.create_stream_producer(stream_name: str, config: ProducerConfig) -> Producer

    创建生产者。
	
    参数:
        - **stream_name** (str) - 流的名称。长度必须小于 256 个字符且仅含有下列字符 `(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)` 。
        - **config** (ProducerConfig_) - 生产者的配置信息。

    返回:
        生产者。即 Producer_ 。

    异常:
        - **RuntimeError** - 如果创建 Producer 失败。

    样例：
        >>> try:
        ...     producer_config = ProducerConfig(
        ...         delay_flush_time=5,
        ...         page_size=1024 * 1024,
        ...         max_stream_size=1024 * 1024 * 1024,
        ...         auto_clean_up=True,
        ...     )
        ...     stream_producer = create_stream_producer("streamName", producer_config)
        ... except RuntimeError as exp:
        ...     # 处理异常
        ...     pass

.. _ProducerConfig: ../../zh_cn/Python/yr.ProducerConfig.html#yr.ProducerConfig
.. _Producer: ../../zh_cn/Python/yr.ProducerConfig.html#yr.ProducerConfig