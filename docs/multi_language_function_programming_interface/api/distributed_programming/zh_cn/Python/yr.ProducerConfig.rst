yr.ProducerConfig
=======================

.. py:class:: yr.ProducerConfig(delay_flush_time: int = 5, page_size: int = 1048576, max_stream_size: int = 1073741824, auto_clean_up: bool = False, encrypt_stream: bool = False, retain_for_num_consumers: int = 0, reserve_size: int = 0, extend_config: ~typing.Dict[str, str] = <factory>)

    基类：``object``

    生产者创建的配置类。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`auto_clean_up <auto_clean_up>`
         - 指定流是否开启自动清理功能。
       * - :ref:`delay_flush_time <delay_flush_time>`
         - 发送数据后最多延迟相应时长后触发一次刷新。
       * - :ref:`encrypt_stream <encrypt_stream>`
         - 指定流是否开启内容加密功能。
       * - :ref:`max_stream_size <max_stream_size>`
         - 指定一个流可使用的最大共享内存大小，单位为字节（B）。
       * - :ref:`page_size <page_size>`
         - 表示生产者对应的缓冲页大小，单位为字节（B）。
       * - :ref:`reserve_size <reserve_size>`
         - 表示预留内存大小，单位为字节（B）。
       * - :ref:`retain_for_num_consumers <retain_for_num_consumers>`
         - 生产者发送的数据将保留到第 ``N`` 个消费者接收。
       * - :ref:`extend_config <extend_config>`
         - 扩展配置，以字典形式存储，允许用户自定义配置项。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_ProducerConfig>`
         -


.. toctree::
    :maxdepth: 1
    :hidden:

    yr.ProducerConfig.auto_clean_up
    yr.ProducerConfig.delay_flush_time
    yr.ProducerConfig.encrypt_stream
    yr.ProducerConfig.max_stream_size
    yr.ProducerConfig.page_size
    yr.ProducerConfig.reserve_size
    yr.ProducerConfig.retain_for_num_consumers
    yr.ProducerConfig.extend_config
    yr.ProducerConfig.__init__

