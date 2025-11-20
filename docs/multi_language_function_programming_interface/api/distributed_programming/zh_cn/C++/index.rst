C++
==============================

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Init
   IsInitialized
   IsLocalMode
   Finalize
   struct-Config
   struct-InstanceRange
   Exception

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   YR_INVOKE
   YR_STATE
   YR_RECOVER
   YR_SHUTDOWN
   Function
   FunctionHandler-Invoke
   FunctionHandler-Options
   FunctionHandler-SetUrn
   Instance
   InstanceCreator-Invoke
   InstanceCreator-Options
   InstanceCreator-SetUrn
   InstanceFunctionHandler-Invoke
   InstanceFunctionHandler-Options
   NamedInstance
   GetInstance
   Cancel
   Exit
   SaveState
   LoadState
   struct-InvokeOptions

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Put
   Get
   Wait
   ObjectRef

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   PyFunction
   PyInstanceClass-FactoryCreate
   JavaFunction
   JavaInstanceClass-FactoryCreate
   CppFunction
   CppInstanceClass-FactoryCreate

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   struct-Group

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   KV-Set
   KV-MSetTx
   KV-Write
   KV-MWriteTx
   KV-WriteRaw
   KV-Get
   KV-GetWithParam
   KV-Read
   KV-ReadRaw
   KV-Del

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   ParallelFor

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

   * - :doc:`Init`
     - 初始化客户端，根据配置连接到 openYuanrong 集群。
   * - :doc:`IsInitialized`
     - 检查是否已调用 init 接口完成初始化。
   * - :doc:`IsLocalMode`
     - 检查函数是否运行在本地模式。
   * - :doc:`Finalize`
     - 关闭客户端，释放函数实例等创建的资源。
   * - :doc:`struct-Config`
     - init 接口使用的配置参数。
   * - :doc:`struct-InstanceRange`
     - Range 调度使用的配置参数。
   * - :doc:`Exception`
     - openYuanrong 抛出的异常。


有状态及无状态函数 API
--------------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`YR_INVOKE`
     - 注册 openYuanrong 函数。
   * - :doc:`YR_STATE`
     - 标记类的成员变量为 openYuanrong 函数的状态。
   * - :doc:`YR_RECOVER`
     - 注册有状态函数实例恢复时的回调函数。
   * - :doc:`YR_SHUTDOWN`
     - 注册有状态函数实例退出时的回调函数。
   * - :doc:`Function`
     - 构造 FunctionHandler 句柄。
   * - :doc:`FunctionHandler-Invoke`
     - 调用无状态函数。
   * - :doc:`FunctionHandler-Options`
     - 指定无状态函数调用时的配置，比如资源、亲和等。
   * - :doc:`FunctionHandler-SetUrn`
     - 指定无状态函数调用时的 URN，配合 CppFunction 或 JavaFunction 接口使用。
   * - :doc:`Instance`
     - 构造 InstanceCreator 对象。
   * - :doc:`InstanceCreator-Invoke`
     - 创建有状态函数实例。
   * - :doc:`InstanceCreator-Options`
     - 指定有状态函数实例创建时的配置项，比如资源、亲和等。
   * - :doc:`InstanceCreator-SetUrn`
     - 指定有状态函数实例创建时的 URN。
   * - :doc:`InstanceFunctionHandler-Invoke`
     - 调用有状态函数的方法。
   * - :doc:`InstanceFunctionHandler-Options`
     - 指定有状态函数方法调用时的配置项，比如重试次数等。
   * - :doc:`NamedInstance`
     - 有状态函数实例的句柄。
   * - :doc:`GetInstance`
     - 获取一个具名有状态函数实例。
   * - :doc:`Cancel`
     - 取消无状态函数或有状态函数的方法调用。
   * - :doc:`Exit`
     - 退出当前函数实例。
   * - :doc:`SaveState`
     - 保存有状态函数实例的状态。
   * - :doc:`LoadState`
     - 导入有状态函数实例的状态。
   * - :doc:`struct-InvokeOptions`
     - 用于设置调用选项。


数据对象 API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Put`
     - 保存数据对象到数据系统。
   * - :doc:`Get`
     - 根据数据对象的键从数据系统中检索值。 
   * - :doc:`Wait`
     - 给定一组数据对象的键，等待指定数量的数据对象的值就绪。
   * - :doc:`ObjectRef`
     - 对象引用，即数据对象的键。


函数互调 API
-----------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`PyFunction`
     - 通过函数名称构造 FunctionHandler 句柄，用于调用 Python 无状态函数。
   * - :doc:`PyInstanceClass-FactoryCreate`
     - 构造 PyInstanceClass 对象，用于创建 Python 有状态函数实例. 
   * - :doc:`JavaFunction`
     - 通过函数名称构造 FunctionHandler 句柄，用于调用 Java 无状态函数。
   * - :doc:`JavaInstanceClass-FactoryCreate`
     - 构造 JavaInstanceClass 对象，用于创建 Java 有状态函数实例。
   * - :doc:`CppFunction`
     - 通过函数名称构造 FunctionHandler 句柄，用于调用 C++ 无状态函数。
   * - :doc:`CppInstanceClass-FactoryCreate`
     - 构造 CppInstanceClass 对象，用于创建 C++ 有状态函数实例。


函数组 API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`struct-Group`
     - 函数组生命周期管理接口。


KV 缓存 API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`KV-Set`
     - 存储二进制数据到数据系统，类似 Redis 的 SET 接口。
   * - :doc:`KV-MSetTx`
     - 批量存储二进制数据到数据系统，类似 Redis 的 Mset 接口。
   * - :doc:`KV-Write`
     - 写入一个键的值。
   * - :doc:`KV-MWriteTx`
     - 设置多个键值对。
   * - :doc:`KV-WriteRaw`
     - 写入键的值（原始字节）。
   * - :doc:`KV-Get`
     - 检索与指定键关联的值，类似于 Redis 的 GET 命令。
   * - :doc:`KV-GetWithParam`
     - 检索与指定键关联的多个值，并支持基于偏移量的额外参数进行读取。
   * - :doc:`KV-Read`
     - 检索键的值。
   * - :doc:`KV-ReadRaw`
     - 检索由 WriteRaw 写入的键的值。
   * - :doc:`KV-Del`
     - 删除一个键及其关联的数据，类似于 Redis 的 DEL 命令。


线程并行 API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`ParallelFor`
     - 用于多线程并行编程的 API。


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