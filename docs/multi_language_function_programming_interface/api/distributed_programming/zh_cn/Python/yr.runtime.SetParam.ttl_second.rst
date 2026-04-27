.. _ttl_second_cp:

yr.runtime.SetParam.ttl_second
------------------------------------

.. py:attribute:: SetParam.ttl_second
   :type: int
   :value: constants.DEFAULT_NO_TTL_LIMIT

   数据生命周期，超过会被删除。
   默认值为 ``0``，表示 key 会一直存在，直到显式调用 del 接口。