.. _page_size:

yr.ProducerConfig.page_size
------------------------------------

.. py:attribute:: ProducerConfig.page_size
   :type: int
   :value: 1048576

   代表生产者对应的缓冲 Page 大小。
   当缓冲页面写满时，会触发刷新操作。默认值为 ``1`` MB，必须大于 0 且是 4 K 的倍数。默认值为 ``1MB`` (``1024 * 1024``)。
