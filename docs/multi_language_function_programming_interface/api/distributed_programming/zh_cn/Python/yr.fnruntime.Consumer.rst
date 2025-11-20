yr.fnruntime.Consumer
=======================

.. py:class:: yr.fnruntime.Consumer

    基类：``object``

    Consumer接口类。

    样例：
        >>> try:
        ...     config = SubscriptionConfig("subName", SubscriptionType.STREAM)
        ...     consumer = create_stream_consumer("streamName", config)
        ...     # .......
        ...     elements = consumer.Receive(6000, 1)
        ... except RuntimeError as exp:
        ...     # .......
        ...     pass

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_consumer>`
         -
       * - :ref:`ack <ack>`
         - 消费者在处理完指定 elementId 的数据后需执行确认（ack）。
       * - :ref:`close <close_consumer>`
         - 关闭消费者，释放内部资源。
       * - :ref:`receive <receive>`
         - 消费者接收数据。





.. toctree::
    :maxdepth: 1
    :hidden:

    yr.fnruntime.Consumer.__init__
    yr.fnruntime.Consumer.ack
    yr.fnruntime.Consumer.close
    yr.fnruntime.Consumer.receive
