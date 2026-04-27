.. _timeout_go:

yr.config.GroupOptions.timeout
------------------------------------

.. py:attribute:: GroupOptions.timeout
   :type: int
   :value: -1

   内核资源不足时，重新调度的超时时间，单位为秒。如果设置为 `-1`，内核将无限次尝试调度。如果设置为小于 `0` 的其他值，将抛出异常。默认值：``-1``。