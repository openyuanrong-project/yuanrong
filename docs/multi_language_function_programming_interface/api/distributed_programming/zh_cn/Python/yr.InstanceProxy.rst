yr.InstanceProxy
===========================================

.. py:class:: yr.InstanceProxy

    基类：``object``

    有状态函数实例的句柄，用于终止实例等操作。

    **方法**：

    .. list-table::
       :widths: 40 60
       :header-rows: 0

       * - :ref:`__init__ <init_ip>`
         - 
       * - :ref:`deserialization_ <deserialization_>`
         - 反序列化以重建实例代理。
       * - :ref:`get_function_group_handler <get_function_group_handler>`
         - 获取 FunctionGroupHandler。
       * - :ref:`is_activate <is_activate>`
         - 返回实例状态。
       * - :ref:`real_id <real_id>`
         - 获取运行时分配给实例的真实 ID。
       * - :ref:`serialization_ <serialization_>`
         - 实例代理的序列化。
       * - :ref:`terminate <terminate>`
         - 终止实例。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.InstanceProxy.__init__
    yr.InstanceProxy.deserialization_
    yr.InstanceProxy.get_function_group_handler
    yr.InstanceProxy.is_activate
    yr.InstanceProxy.real_id
    yr.InstanceProxy.serialization_
    yr.InstanceProxy.terminate