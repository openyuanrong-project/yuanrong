.. _retain_for_num_consumers:

yr.ProducerConfig.retain_for_num_consumers
-----------------------------------------------

.. py:attribute:: ProducerConfig.retain_for_num_consumers
   :type: int
   :value: 0

   生产者发送的数据将保留到第 N 个消费者接收。
   默认值为 ``0``，表示生产者发送数据后，如果没有消费者，数据不会保留，等消费者创建后可能会接收不到数据。该参数仅对首次创建的消费者有效，不支持多个消费者。