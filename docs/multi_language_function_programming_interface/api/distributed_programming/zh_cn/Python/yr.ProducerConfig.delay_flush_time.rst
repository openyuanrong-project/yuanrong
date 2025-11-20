.. _delay_flush_time:

yr.ProducerConfig.delay_flush_time
------------------------------------

.. py:attribute:: ProducerConfig.delay_flush_time
   :type: int
   :value: 5

   用于配置在发送数据后延迟多长时间触发一次 Flush 操作。
   可选参数为 ``CacheType.Memory`` （内存介质）和 ``CacheType.Disk`` （磁盘介质），默认值为 ``CacheType.Memory``。