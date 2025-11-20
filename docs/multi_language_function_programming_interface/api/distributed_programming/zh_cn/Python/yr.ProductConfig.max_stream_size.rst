.. _max_stream_size:

yr.ProducerConfig.max_stream_size
------------------------------------

.. py:attribute:: ProducerConfig.max_stream_size
   :type: int
   :value: 1024 * 1024 * 1024

   指定流在 worker 上最大能使用的共享内存大小，单位是字节（B）。
   默认值为 ``1`` GB，范围在 64 KB 到 worker 共享内存的大小之间。