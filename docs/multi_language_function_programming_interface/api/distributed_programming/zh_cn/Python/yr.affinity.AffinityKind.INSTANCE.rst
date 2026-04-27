.. _INSTANCE:

yr.affinity.AffinityKind.INSTANCE
------------------------------------------------

.. py:attribute:: AffinityKind.INSTANCE
   :value: 20

    实例亲和类型

    实例的标签只包含键。对于实例亲和，标签选择条件只能设置为 LABEL_EXISTS 或 LABEL_NOT_EXISTS。
   
    弱亲和/弱反亲和不能保证实例一定被调度到指定的 POD 上。如果未找到满足条件的 POD，可能会选择其他有资源的 POD 进行调度。