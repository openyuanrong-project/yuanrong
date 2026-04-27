.. _preferred_priority:

yr.InvokeOptions.preferred_priority
----------------------------------------------

.. py:attribute:: InvokeOptions.preferred_priority
   :value: True

   设置是否启用弱亲和优先级调度。如果启用，当传递多个弱亲和条件时，将按顺序匹配和打分。只要满足一个条件，调度即成功。