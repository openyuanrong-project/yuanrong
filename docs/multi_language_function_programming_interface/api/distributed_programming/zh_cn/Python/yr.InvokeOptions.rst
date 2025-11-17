yr.InvokeOptions
==========================

.. py:class:: yr.InvokeOptions(cpu: int = 500, memory: int = 500, concurrency: int = 1, custom_resources: ~typing.Dict[str, float] = <factory>, custom_extensions: ~typing.Dict[str, str] = <factory>, pod_labels: ~typing.Dict[str, str] = <factory>, labels: ~typing.List[str] = <factory>, max_invoke_latency: int = 5000, min_instances: int = 0, max_instances: int = 0, recover_retry_times: int = 0, need_order: bool = False, name: str = '', namespace: str = '', schedule_affinities: ~typing.List[~yr.affinity.Affinity] = <factory>, resource_group_options: ~yr.config.ResourceGroupOptions = <factory>, function_group_options: ~yr.config.FunctionGroupOptions = <factory>, env_vars: ~typing.Dict[str, str] = <factory>, retry_times: int = 0, trace_id: str = '', alias_params: ~typing.Dict[str, str] = <factory>, runtime_env: ~typing.Dict = <factory>)

    基类：``object``

    用于设置调用选项。

    样例：

    .. code-block:: python

        import yr
        import time
        yr.init()
        opt = yr.InvokeOptions()
        opt.pod_labels["k1"] = "v1"
        @yr.invoke(invoke_options=opt)
        def func():
            time.sleep(100)
        ret = func.invoke()
        yr.finalize()

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`concurrency <concurrency_IO>`
         - 实例并发度。
       * - :ref:`cpu <cpu_IO>`
         - 所需 CPU 的大小。
       * - :ref:`max_instances <max_instances>`
         - 指定无状态函数的最大实例数。
       * - :ref:`max_invoke_latency <max_invoke_latency>`
         - 指定期望的异构函数调用完成的时间。
       * - :ref:`memory <memory_IO>`
         - 所需内存的大小。
       * - :ref:`min_instances <min_instances>`
         - 指定无状态函数的最小实例数。
       * - :ref:`name <name_IO>`
         - 用于指定实例的名称。
       * - :ref:`namespace <namespace_IO>`
         - 用于指定实例的命名空间。
       * - :ref:`need_order <need_order>`
         - 是否启用顺序保持。
       * - :ref:`preferred_anti_other_labels <preferred_anti_other_labels>`
         - 是否启用不可选资源的反亲和性。
       * - :ref:`preferred_priority <preferred_priority>`
         - 设置是否启用弱亲和优先级调度。
       * - :ref:`recover_retry_times <recover_retry_times_IO>`
         - 实例恢复次数（当实例异常退出时，实例将自动恢复到最新状态）。
       * - :ref:`required_priority <required_priority>`
         - 设置是否启用强亲和优先级调度。
       * - :ref:`retry_times <retry_times_IO>`
         - 无状态函数的重试次数。
       * - :ref:`trace_id <trace_id_IO>`
         - 为函数调用设置 traceId，用于链路追踪。
       * - :ref:`custom_resources <custom_resources>`
         - 自定义资源目前支持“GPU/XX/YY”和“NPU/XX/YY”，其中 XX 是卡型号，如 Ascend910B4，YY 可以是 count、latency 或 stream。
       * - :ref:`custom_extensions <custom_extensions>`
         - 指定用户自定义配置，例如函数并发度。
       * - :ref:`pod_labels <pod_labels_IO>`
         - Pod 标签仅在 Kubernetes 环境中使用。
       * - :ref:`labels <labels_IO>`
         - 实例标签。
       * - :ref:`schedule_affinities <schedule_affinities>`
         - 设置亲和条件列表。
       * - :ref:`resource_group_options <resource_group_options>`
         - 指定 ResourceGroup 选项，包括 resource_group_name 和 bundle_index。
       * - :ref:`function_group_options <function_group_options>`
         - 函数组选项。
       * - :ref:`env_vars <env_vars_IO>`
         - 实例启动时设置环境变量。
       * - :ref:`alias_params <alias_params>`
         - 在 FaaS 跨函数调用中，当通过指定的别名调用函数且该别名为规则别名时，此参数用于设置规则别名所依赖的 kv 参数。
       * - :ref:`runtime_env <runtime_env_IO>`
         - 使用 conda、pip、working_dir 和 env_vars 配置 actor/task 的运行时环境。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_IO>`
         - 
       * - :ref:`check_options_valid <check_options_valid>`
         - 检查选项是否有效。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.InvokeOptions.concurrency
    yr.InvokeOptions.cpu
    yr.InvokeOptions.max_instances
    yr.InvokeOptions.max_invoke_latency  
    yr.InvokeOptions.memory
    yr.InvokeOptions.min_instances  
    yr.InvokeOptions.name  
    yr.InvokeOptions.namespace  
    yr.InvokeOptions.need_order  
    yr.InvokeOptions.preferred_anti_other_labels  
    yr.InvokeOptions.preferred_priority
    yr.InvokeOptions.recover_retry_times
    yr.InvokeOptions.required_priority
    yr.InvokeOptions.retry_times 
    yr.InvokeOptions.trace_id 
    yr.InvokeOptions.custom_resources
    yr.InvokeOptions.custom_extensions
    yr.InvokeOptions.pod_labels
    yr.InvokeOptions.labels
    yr.InvokeOptions.schedule_affinities
    yr.InvokeOptions.resource_group_options
    yr.InvokeOptions.function_group_options  
    yr.InvokeOptions.env_vars
    yr.InvokeOptions.alias_params  
    yr.InvokeOptions.runtime_env
    yr.InvokeOptions.__init__ 
    yr.InvokeOptions.check_options_valid
    
