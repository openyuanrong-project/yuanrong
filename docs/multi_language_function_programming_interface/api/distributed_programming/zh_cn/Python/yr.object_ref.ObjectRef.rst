yr.object_ref.ObjectRef
==========================

.. py:class:: yr.object_ref.ObjectRef(object_id: str, task_id=None, need_incre=True, need_decre=True, exception=None)

    基类：``object``

    对象引用，即数据对象的键。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`id <id_ObjRef>`
         - ObjectRef的id。
       * - :ref:`task_id <task_id_ObjRef>`
         - 任务ID。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_ObjectRef>`
         - 
       * - :ref:`as_future <as_future>`
         - 使用 asyncio.Future 包装 ObjectRef。
       * - :ref:`cancel <cancel_ObjectRef>`
         - 取消对象的未来任务。
       * - :ref:`done <done_ObjectRef>`
         - 如果对象的未来任务已被取消或执行完成，则返回 True。
       * - :ref:`exception <exception_ObjectRef>`
         - 如果异常不为 None，则抛出异常。
       * - :ref:`get <get_ObjectRef>`
         - 此函数用于 FaaS 模式中检索对象。
       * - :ref:`get_future <get_future>`
         - 获取未来任务。
       * - :ref:`is_exception <is_exception>`
         - 是否为未来任务异常。
       * - :ref:`on_complete <on_complete>`
         - 注册回调函数。
       * - :ref:`set_data <set_data>`
         - 设置数据。
       * - :ref:`set_exception <set_exception>`
         - 设置异常。
       * - :ref:`wait <wait_ObjectRef>`
         - 等待任务完成。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.object_ref.ObjectRef.id
    yr.object_ref.ObjectRef.task_id
    yr.object_ref.ObjectRef.__init__  
    yr.object_ref.ObjectRef.as_future  
    yr.object_ref.ObjectRef.cancel  
    yr.object_ref.ObjectRef.done  
    yr.object_ref.ObjectRef.exception  
    yr.object_ref.ObjectRef.get  
    yr.object_ref.ObjectRef.get_future  
    yr.object_ref.ObjectRef.is_exception  
    yr.object_ref.ObjectRef.on_complete  
    yr.object_ref.ObjectRef.set_data  
    yr.object_ref.ObjectRef.set_exception  
    yr.object_ref.ObjectRef.wait

