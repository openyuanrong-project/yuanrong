.. _scheduling_affinity_each_bundle_size:

yr.FunctionGroupOptions.scheduling_affinity_each_bundle_size
--------------------------------------------------------------

.. py:attribute:: FunctionGroupOptions.scheduling_affinity_each_bundle_size
   :type: Optional[int]
   :value: None

   每个bundle中的函数实例的数量。
   超时时间，单位为秒。限制：``-1, [0, 0x7FFFFFFF]``， 默认值为 ``-1`` ，表示阻塞等。
