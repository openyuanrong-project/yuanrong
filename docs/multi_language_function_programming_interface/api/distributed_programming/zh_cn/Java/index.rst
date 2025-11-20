Java
==============================

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1
   
   init
   Finalize
   Config

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   invoke
   function
   FunctionHandler
   function-options
   Instance
   InstanceCreator
   InstanceHandler
   InstanceHandler-function
   InstanceFunctionHandler
   InstanceHandler-exportHandler
   InstanceHandler-importHandler
   InstanceHandler-terminate
   saveState
   loadState
   yrShutdown
   yrRecover
   InvokeOptions

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   put
   get
   wait
   ObjectRef

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   setUrn
   cpp-function
   CppFunctionHandler
   cpp_instance_class
   cpp_instance_method
   CppInstanceCreator
   CppInstanceHandler
   CppInstanceHandler-function
   CppInstanceFunctionHandler
   CppInstanceHandler-exportHandler
   CppInstanceHandler-importHandler
   CppInstanceHandler-terminate
   java-function
   JavaFunctionHandler
   java_instance_class
   java_instance_method
   JavaInstanceCreator
   JavaInstanceHandler
   JavaInstanceHandler-function
   JavaInstanceFunctionHandler
   JavaInstanceHandler-exportHandler
   JavaInstanceHandler-importHandler
   JavaInstanceHandler-terminate
   VoidFunctionHandler
   VoidInstanceFunctionHandler

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   kv.set
   kv.mSetTx
   kv.write
   kv.mWriteTx
   kv.get
   kv.getWithParam
   kv.read
   kv.del

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Group

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Affinity

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Stream

基础 API
----------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`init`
     - 初始化客户端，根据配置连接到 openYuanrong 集群。
   * - :doc:`Finalize`
     - 关闭客户端，释放函数实例等创建的资源。
   * - :doc:`Config`
     - init 接口使用的配置参数。


有状态及无状态函数 API
--------------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`invoke`
     - 执行有状态函数实例的创建、无状态函数或有状态函数方法的调用。
   * - :doc:`function`
     - 构造 FunctionHandler 句柄。
   * - :doc:`FunctionHandler`
     - 无状态函数句柄，使用它执行调用。
   * - :doc:`function-options`
     - 用于在创建有状态函数实例、调用无状态函数或有状态函数的方法时，指定资源，亲和等配置。
   * - :doc:`Instance`
     - 构造 InstanceCreator 对象。
   * - :doc:`InstanceCreator`
     - 用于创建有状态函数实例。
   * - :doc:`InstanceHandler`
     - 有状态函数实例的句柄，用于终止实例等操作。
   * - :doc:`InstanceHandler-function`
     - 构造 InstanceFunctionHandler 句柄。
   * - :doc:`InstanceFunctionHandler`
     - 有状态函数实例的方法句柄，使用它执行调用。
   * - :doc:`InstanceHandler-exportHandler`
     - 导出有状态函数实例的句柄信息。
   * - :doc:`InstanceHandler-importHandler`
     - 导入有状态函数实例的句柄信息。
   * - :doc:`InstanceHandler-terminate`
     - 终止有状态函数实例。
   * - :doc:`saveState`
     - 保存有状态函数实例的状态。
   * - :doc:`loadState`
     - 加载有状态函数实例的状态。
   * - :doc:`yrShutdown`
     - 注册有状态函数实例退出时的回调函数。
   * - :doc:`yrRecover`
     - 注册有状态函数实例恢复时的回调函数。
   * - :doc:`InvokeOptions`
     - invoke 接口使用的配置参数。


数据对象 API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`put`
     - 保存数据对象到数据系统。
   * - :doc:`get`
     - 根据数据对象的键从数据系统中检索值。
   * - :doc:`wait`
     - 给定一组数据对象的键，等待指定数量的数据对象的值就绪。
   * - :doc:`ObjectRef`
     - 对象引用，即数据对象的键。


函数互调 API
-----------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`setUrn`
     - 设置有状态或无状态函数的 URN。
   * - :doc:`cpp-function`
     - 辅助构造 CppFunctionHandler 句柄。
   * - :doc:`CppFunctionHandler`
     - C++ 无状态函数句柄，使用它执行调用。
   * - :doc:`cpp_instance_class`
     - 辅助生成 CppInstanceCreator 对象。 
   * - :doc:`cpp_instance_method`
     - 辅助构造 CppInstanceFunctionHandler 句柄。
   * - :doc:`CppInstanceCreator`
     - 用于创建 C++ 有状态函数实例。
   * - :doc:`CppInstanceHandler`
     - C++ 有状态函数实例的句柄，用于终止实例等操作。
   * - :doc:`CppInstanceHandler-function`
     - 构造 CppInstanceFunctionHandler 句柄。
   * - :doc:`CppInstanceFunctionHandler`
     - C++ 有状态函数实例的方法句柄，使用它执行调用。
   * - :doc:`CppInstanceHandler-exportHandler`
     - 导出 C++ 有状态函数实例的句柄信息。
   * - :doc:`CppInstanceHandler-importHandler`
     - 导入 C++ 有状态函数实例的句柄信息。
   * - :doc:`CppInstanceHandler-terminate`
     - 终止 C++ 有状态函数实例。
   * - :doc:`java-function`
     - 辅助构造 JavaFunctionHandler 句柄。
   * - :doc:`JavaFunctionHandler`
     - Java 无状态函数句柄，使用它执行调用。
   * - :doc:`java_instance_class`
     - 辅助生成 JavaInstanceCreator 对象。 
   * - :doc:`java_instance_method`
     - 辅助构造 JavaInstanceFunctionHandler 句柄。
   * - :doc:`JavaInstanceCreator`
     - 用于创建 Java 有状态函数实例。
   * - :doc:`JavaInstanceHandler`
     - Java 有状态函数实例的句柄，用于终止实例等操作。
   * - :doc:`JavaInstanceHandler-function`
     - 构造 JavaInstanceFunctionHandler 句柄。
   * - :doc:`JavaInstanceFunctionHandler`
     - Java 有状态函数实例的方法句柄，使用它执行调用。
   * - :doc:`JavaInstanceHandler-exportHandler`
     - 导出 Java 有状态函数实例的句柄信息。
   * - :doc:`JavaInstanceHandler-importHandler`
     - 导入 Java 有状态函数实例的句柄信息。
   * - :doc:`JavaInstanceHandler-terminate`
     - 终止 Java 有状态函数实例。
   * - :doc:`VoidFunctionHandler`
     - 无返回值的无状态函数句柄，使用它执行调用。
   * - :doc:`VoidInstanceFunctionHandler`
     - 无返回值的有状态函数实例的方法句柄，使用它执行调用。


KV 缓存 API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`kv.set`
     - 存储二进制数据到数据系统，类似 Redis 的 SET 接口。
   * - :doc:`kv.mSetTx`
     - 批量存储二进制数据到数据系统，类似 Redis 的 Mset 接口。
   * - :doc:`kv.write`
     - 写入一个键的值。
   * - :doc:`kv.mWriteTx`
     - 设置多个键值对。
   * - :doc:`kv.get`
     - 检索与指定键关联的值，类似于 Redis 的 GET 命令。
   * - :doc:`kv.getWithParam`
     - 检索与指定键关联的多个值，并支持基于偏移量的额外参数进行读取。
   * - :doc:`kv.read`
     - 检索与指定键关联的指定类型的值。
   * - :doc:`kv.del`
     - 删除一个键及其关联的数据，类似于 Redis 的 DEL 命令。


函数组 API
-----------------
 
.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Group`
     - 函数组生命周期管理接口。


亲和调度
-----------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Affinity`
     - 亲和调度的配置参数。


数据流 API
-----------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Stream`
     - 在 openYuanrong 函数间传递共享的有序无界内存数据集。
