yr.InstanceCreator
=============================================

.. py:class:: yr.InstanceCreator

    基类：``object``

    用于创建有状态函数实例。

    **方法**：

    .. list-table::
       :widths: 40 60
       :header-rows: 0

       * - :ref:`__init__ <init_ic>`
         - 
       * - :ref:`create_cpp_user_class <create_cpp_user_class>`
         - 创建一个 cpp 用户类。
       * - :ref:`create_cross_user_class <create_cross_user_class>`
         - 创建一个 Java 用户类。
       * - :ref:`create_from_user_class <create_from_user_class>`
         - 从用户类创建。
       * - :ref:`get_instance <get_instance>`
         - 在集群中获取一个实例。
       * - :ref:`get_original_cls <get_original_cls>`
         - 获取原始类。
       * - :ref:`invoke <invoke_ic>`
         - 在集群中创建一个实例。
       * - :ref:`options <options_ic>`
         - YR 的 options。
       * - :ref:`set_function_group_size <set_function_group_size_ic>`
         - 设置函数组大小。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.InstanceCreator.__init__
    yr.InstanceCreator.create_cpp_user_class
    yr.InstanceCreator.create_cross_user_class
    yr.InstanceCreator.create_from_user_class
    yr.InstanceCreator.get_instance
    yr.InstanceCreator.get_original_cls
    yr.InstanceCreator.invoke
    yr.InstanceCreator.options
    yr.InstanceCreator.set_function_group_size