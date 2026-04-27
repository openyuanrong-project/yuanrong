yr.FunctionGroupOptions
=======================

.. py:class:: yr.FunctionGroupOptions(cpu: int | None = None, memory: int | None = None, resources: ~typing.Dict[str, float] = <factory>, scheduling_affinity_type: ~yr.config.SchedulingAffinityType | None = None, scheduling_affinity_each_bundle_size: int | None = None, timeout: int | None = None, concurrency: int | None = None, recover_retry_times: int = 0)

    基类：``object``

    函数组的配置项。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`concurrency <concurrency>`
         - 实例并发度。 限制范围：[1,1000]。
       * - :ref:`cpu <cpu>`
         - 需要使用的 cpu 大小。 单位：m（毫核），cpu 限制[300,16000]。
       * - :ref:`memory <memory_FGO>`
         - 需要使用的内存大小。 单位：MB，memory 限制： [128,65536]。
       * - :ref:`recover_retry_times <recover_retry_times>`
         - 恢复重试次数，用于在实例恢复失败时的重试。默认值为 ``0`` 。
       * - :ref:`scheduling_affinity_each_bundle_size <scheduling_affinity_each_bundle_size>`
         - 每个 bundle 中的函数实例数。
       * - :ref:`scheduling_affinity_type <scheduling_affinity_type>`
         - bundle 内实例亲和类型。
       * - :ref:`timeout <timeout_FGO>`
         - 超时时间（以秒为单位），有效值为 ``-1`` 或在 ``[0,0x7fffffff]`` 内。
       * - :ref:`resources <resources>`
         - 自定义资源，目前支持 ``NPU/XX/YY``，其中 XX 为插卡型号，如 ``Ascend910B4`` ， YY 为 ``count``，``latency``，或者 ``stream``。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_FunctionGroupOptions>`
         -





.. toctree::
    :maxdepth: 1
    :hidden:

    yr.FunctionGroupOptions.concurrency
    yr.FunctionGroupOptions.cpu
    yr.FunctionGroupOptions.memory
    yr.FunctionGroupOptions.recover_retry_times
    yr.FunctionGroupOptions.scheduling_affinity_each_bundle_size
    yr.FunctionGroupOptions.scheduling_affinity_type
    yr.FunctionGroupOptions.timeout
    yr.FunctionGroupOptions.resources
    yr.FunctionGroupOptions.__init__
    
    
    

