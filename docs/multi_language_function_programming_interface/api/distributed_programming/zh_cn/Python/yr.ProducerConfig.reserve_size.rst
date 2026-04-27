.. _reserve_size:

yr.ProducerConfig.reserve_size
------------------------------------

.. py:attribute:: ProducerConfig.reserve_size
   :type: int
   :value: 0

   代表预留内存大小，单位是字节（B）。
   当创建生产者时，会尝试预留 ``reserve_size`` 字节大小的内存。若预留失败，则创建生产者时会抛出异常。``reserve_size`` 必须为 ``page_size`` 的整数倍，且取值范围为 ``[0, max_stream_size]``。若 ``reserve_size`` 为 ``0``，则会被设置成 ``page_size``。默认值为 ``0``。