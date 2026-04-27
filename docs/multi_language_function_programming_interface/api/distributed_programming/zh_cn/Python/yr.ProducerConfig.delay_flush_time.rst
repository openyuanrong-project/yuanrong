.. _delay_flush_time:

yr.ProducerConfig.delay_flush_time
------------------------------------

.. py:attribute:: ProducerConfig.delay_flush_time
   :type: int
   :value: 5

   用于配置在发送数据后延迟多长时间触发一次 Flush 操作，单位是毫秒（ms）。
   <0：不自动刷新；0：立即刷新；其他值：延迟时间。默认值为 ``5``。