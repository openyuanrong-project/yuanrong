Python
==============================

.. TODO yr_shutdown  generator 


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.init
   yr.is_initialized
   yr.finalize
   yr.Config
   yr.config.ClientInfo


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.invoke
   yr.FunctionProxy
   yr.instance
   yr.InstanceProxy
   yr.method
   yr.MethodProxy
   yr.get_instance
   yr.InstanceCreator
   yr.cancel
   yr.exit
   yr.save_state
   yr.load_state
   yr.InvokeOptions
   yr.list_named_instances


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.put
   yr.get
   yr.wait
   yr.object_ref.ObjectRef

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.cpp_function
   yr.cpp_instance_class
   yr.java_function
   yr.java_instance_class


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.create_function_group
   yr.get_function_group_context
   yr.FunctionGroupOptions
   yr.FunctionGroupContext
   yr.FunctionGroupHandler
   yr.FunctionGroupMethodProxy
   yr.device.DataInfo


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.create_resource_group
   yr.remove_resource_group
   yr.ResourceGroup
   yr.config.ResourceGroupOptions
   yr.config.SchedulingAffinityType
   yr.config.UserTLSConfig
   yr.config.DeploymentConfig


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.kv_write
   yr.kv_write_with_param
   yr.kv_m_write_tx
   yr.kv_read
   yr.kv_del
   yr.kv_set
   yr.kv_get
   yr.kv_get_with_param
   yr.runtime.ExistenceOpt
   yr.runtime.WriteMode
   yr.runtime.CacheType
   yr.runtime.ConsistencyType
   yr.runtime.GetParam
   yr.runtime.GetParams
   yr.runtime.SetParam
   yr.runtime.MSetParam
   yr.runtime.CreateParam


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.Gauge
   yr.Alarm
   yr.AlarmInfo
   yr.AlarmSeverity
   yr.DoubleCounter
   yr.UInt64Counter
   yr.resources


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1
   
   yr.affinity.AffinityType
   yr.affinity.AffinityKind
   yr.affinity.OperatorType
   yr.affinity.LabelOperator
   yr.affinity.Affinity


.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   yr.exception.YRError
   yr.exception.CancelError
   yr.exception.YRInvokeError
   yr.exception.YRequestError

基础 API
---------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`yr.init`
     - 初始化客户端，根据配置连接到 openYuanrong 集群。 
   * - :doc:`yr.is_initialized`
     - 检查是否已调用 init 接口完成初始化。
   * - :doc:`yr.finalize`
     - 关闭客户端，释放函数实例等创建的资源。
   * - :doc:`yr.Config`
     - init 接口使用的配置参数。
   * - :doc:`yr.config.ClientInfo`
     - 用于存储 openYuanrong 客户端信息。


有状态及无状态函数 API
-----------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`yr.invoke`
     - 定义一个无状态函数。
   * - :doc:`yr.FunctionProxy`
     - 无状态函数句柄，使用它执行调用。
   * - :doc:`yr.instance`
     - 定义一个有状态函数。
   * - :doc:`yr.InstanceProxy`
     - 有状态函数实例的句柄，用于终止实例等操作。
   * - :doc:`yr.method`
     - 定义有状态函数的方法。
   * - :doc:`yr.MethodProxy`
     - 有状态函数实例的方法句柄，使用它执行调用。
   * - :doc:`yr.get_instance`
     - 根据具名实例的 `name` 和 `namespace` 获取实例句柄。
   * - :doc:`yr.InstanceCreator`
     - 用于创建有状态函数实例。
   * - :doc:`yr.cancel`
     - 取消无状态函数调用。
   * - :doc:`yr.exit`
     - 退出当前函数实例。
   * - :doc:`yr.save_state`
     - 保存有状态函数实例的状态。
   * - :doc:`yr.load_state`
     - 导入有状态函数实例的状态。
   * - :doc:`yr.InvokeOptions`
     - 用于设置调用选项。

数据对象 API
--------------

.. list-table::
   :header-rows: 0
   :widths: 30 70
   
   * - :doc:`yr.put`
     - 保存数据对象到数据系统。 
   * - :doc:`yr.get`
     - 根据数据对象的键从数据系统中检索值。
   * - :doc:`yr.wait`
     - 给定一组数据对象的键，等待指定数量的数据对象的值就绪。
   * - :doc:`yr.object_ref.ObjectRef`
     - 对象引用，即数据对象的键。

函数互调 API
-------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`yr.cpp_function`
     - 用于构造 cpp 函数的代理，远程调用 cpp 函数。 
   * - :doc:`yr.cpp_instance_class`
     - 为 cpp 类构造代理以启用远程调用。
   * - :doc:`yr.java_function`
     - 用于构造 java 函数的代理，远程调用 java 函数。
   * - :doc:`yr.java_instance_class`
     - 用于构造 Java 类，并远程调用 Java 类的代理。

函数组 API
--------------
   
.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`yr.create_function_group`
     - 创建函数组实例。
   * - :doc:`yr.get_function_group_context`
     - 获取函数实例的函数组上下文。
   * - :doc:`yr.FunctionGroupOptions`
     - 函数组的配置项。
   * - :doc:`yr.FunctionGroupContext`
     - 用于管理函数组信息的上下文。
   * - :doc:`yr.FunctionGroupHandler`
     - 函数组句柄，用于终止实例组等操作。
   * - :doc:`yr.FunctionGroupMethodProxy`
     - 函数组的方法句柄，使用它执行调用。
   * - :doc:`yr.device.DataInfo`
     - DataInfo 类用于存储数据的基本信息。

资源组 API
-------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70
   
   * - :doc:`yr.create_resource_group`
     - 异步创建一个 ResourceGroup。
   * - :doc:`yr.remove_resource_group`
     - 异步删除一个 ResourceGroup。
   * - :doc:`yr.ResourceGroup`
     - 创建 ResourceGroup 后返回的句柄。
   * - :doc:`yr.config.ResourceGroupOptions`
     - 资源组选项。
   * - :doc:`yr.config.SchedulingAffinityType`
     - 调度亲和性类型，用于定义资源组的调度策略。
   * - :doc:`yr.config.UserTLSConfig`
     - 用户 SSL/TLS 配置。
   * - :doc:`yr.config.DeploymentConfig`
     - 自动部署配置类。


KV 缓存 API
-------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`yr.kv_write`
     - 提供类 Redis 的 set 存储接口，支持保存二进制数据到数据系统。
   * - :doc:`yr.kv_write_with_param`
     - 提供支持自定义参数配置的存储接口。用于以更灵活的方式将二进制数据写入数据系统。
   * - :doc:`yr.kv_m_write_tx`
     - 提供一个类似 Redis 的集合存储接口，支持将一组二进制数据保存到数据系统。 
   * - :doc:`yr.kv_read`
     - 提供类 Redis 的 get 获取数据接口，支持同时获取多个数据。
   * - :doc:`yr.kv_del`
     - 提供类 Redis 的 del 删除数据接口，支持同时删除多个数据。
   * - :doc:`yr.kv_set`
     - 提供类似 Redis 的集合存储接口，支持二进制数据保存到数据系统。
   * - :doc:`yr.kv_get`
     - 提供类 Redis 的 get 获取数据接口，支持同时获取多个数据。
   * - :doc:`yr.kv_get_with_param`
     - 提供支持偏移读的接口 kv_get_with_param, 支持偏移读取数据。
   * - :doc:`yr.runtime.ExistenceOpt`
     - Kv 选项。
   * - :doc:`yr.runtime.WriteMode`
     - 对象写入模式。
   * - :doc:`yr.runtime.CacheType`
     - 缓存类型。
   * - :doc:`yr.runtime.ConsistencyType`
     - 一致性类型。
   * - :doc:`yr.runtime.GetParam`
     - 获取参数配置类。
   * - :doc:`yr.runtime.GetParams`
     - 获取参数的接口类。
   * - :doc:`yr.runtime.SetParam`
     - 设置参数。
   * - :doc:`yr.runtime.MSetParam`
     - 表示 `mset` 操作的参数配置类。
   * - :doc:`yr.runtime.CreateParam`
     - 创建参数。

可观测性 API
-------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`yr.Gauge`
     - 用于上报 Metrics 数据。
   * - :doc:`yr.Alarm`
     - 用于设置和管理告警信息。
   * - :doc:`yr.AlarmInfo`
     - 告警信息。
   * - :doc:`yr.AlarmSeverity`
     - 资源组选项。
   * - :doc:`yr.DoubleCounter`
     - 用于记录双精度浮点数计数器的指标。
   * - :doc:`yr.UInt64Counter`
     - 用于记录 64 位无符号整数计数器的指标。
   * - :doc:`yr.resources`
     - 获取集群中节点的资源信息。

亲和调度
-------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`yr.affinity.AffinityType`
     - 亲和类型枚举。
   * - :doc:`yr.affinity.AffinityKind`
     - 亲和种类枚举。
   * - :doc:`yr.affinity.OperatorType`
     - 标签操作符类型枚举。
   * - :doc:`yr.affinity.LabelOperator`
     - 亲和的标签操作符。
   * - :doc:`yr.affinity.Affinity`
     - 表示一个亲和配置。

异常
-------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70
   
   * - :doc:`yr.exception.YRError`
     - YR 模块中所有自定义异常的基类。
   * - :doc:`yr.exception.CancelError`
     - 任务取消错误。
   * - :doc:`yr.exception.YRInvokeError`
     - 表示在调用期间发生的错误。
   * - :doc:`yr.exception.YRequestError`
     - 请求失败错误。
