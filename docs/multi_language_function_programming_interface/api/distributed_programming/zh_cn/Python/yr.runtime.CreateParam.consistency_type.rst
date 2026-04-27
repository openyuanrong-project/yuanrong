.. _consistency_type:

yr.runtime.CreateParam.consistency_type
----------------------------------------

.. py:attribute:: CreateParam.consistency_type
   :type: ConsistencyType
   :value: 0

   分布式场景下可配置不同级别的一致性语义。
   可选参数为 ``ConsistencyType.PRAM`` （异步模式）和 ``ConsistencyType.CAUSAL`` （因果一致性），默认采用 ``ConsistencyType.PRAM`` 模式。
