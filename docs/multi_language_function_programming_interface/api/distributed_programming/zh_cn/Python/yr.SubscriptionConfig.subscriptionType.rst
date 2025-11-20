.. _subscriptionType:

yr.SubscriptionConfig.subscriptionType
------------------------------------

.. py:attribute:: SubscriptionConfig.subscriptionType
   :type: SubscriptionType
   :value: 0

   订阅类型。
   包括 ``STREAM``、``ROUND_ROBIN`` 和 ``KEY_PARTITIONS`` 三种类型。
   ``STREAM`` 表示一个订阅组下单消费者消费，``ROUND_ROBIN`` 表示一个订阅组下多消费者循环负载分担式消费，``KEY_PARTITIONS`` 表示一个订阅组下多消费者按 Key 分片负载分担式消费。
   目前仅支持 ``STREAM`` 类型，其他类型暂时不支持。默认订阅类型为 ``STREAM``。