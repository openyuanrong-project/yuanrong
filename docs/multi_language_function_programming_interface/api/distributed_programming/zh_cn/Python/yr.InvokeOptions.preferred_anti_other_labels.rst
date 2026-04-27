.. _preferred_anti_other_labels:

yr.InvokeOptions.preferred_anti_other_labels
----------------------------------------------

.. py:attribute:: InvokeOptions.preferred_anti_other_labels
   :value: True

   是否启用不可选资源的反亲和性。如果启用，当没有任何弱亲和条件满足时，调度失败。当 `preferred_anti_other_labels` 设置为 ``True`` 时，如果找不到满足弱亲和/反亲和条件的 POD，调度失败，且不会选择其他资源的 POD 进行调度。