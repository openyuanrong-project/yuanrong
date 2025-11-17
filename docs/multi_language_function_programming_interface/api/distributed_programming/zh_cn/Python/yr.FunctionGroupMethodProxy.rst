yr.FunctionGroupMethodProxy
============================

.. py:class:: yr.FunctionGroupMethodProxy(method_name: str, class_descriptor: ObjectDescriptor, proxy_list: List[InstanceProxy])

    基类：``object``

    函数组的方法句柄，使用它执行调用。

    初始化方法，用于创建类的实例。
    
    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`use_shared_memory <use_shared_memory>`
         - 此标志用于启用共享内存，默认值为 ``False``。
       * - :ref:`rpc_broadcast_mq <rpc_broadcast_mq>`
         - 用于 RPC 广播的消息队列实例。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_FunctionGroupMethodProxy>`
         - 
       * - :ref:`invoke <invoke_FunctionGroupMethodProxy>`
         - 执行对用户函数的远程调用。
       * - :ref:`set_rpc_broadcast_mq <set_rpc_broadcast_mq>`
         - 设置用于 RPC 广播的消息队列。        
    
.. toctree::
    :maxdepth: 1
    :hidden:

    yr.FunctionGroupMethodProxy.use_shared_memory
    yr.FunctionGroupMethodProxy.rpc_broadcast_mq
    yr.FunctionGroupMethodProxy.__init__
    yr.FunctionGroupMethodProxy.invoke
    yr.FunctionGroupMethodProxy.set_rpc_broadcast_mq

