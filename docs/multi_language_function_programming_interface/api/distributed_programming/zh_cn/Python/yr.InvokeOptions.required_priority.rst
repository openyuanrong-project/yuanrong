.. _required_priority:

yr.InvokeOptions.required_priority
----------------------------------------------

.. py:attribute:: FunctionGroupOptions.required_priority
   :value: False

   设置是否启用强亲和优先级调度。如果启用，当传递多个强亲和条件时，将按顺序匹配和打分。如果没有任何强亲和条件满足，调度失败。