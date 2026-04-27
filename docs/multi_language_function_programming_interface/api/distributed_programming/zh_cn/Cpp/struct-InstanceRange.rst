InstanceRange 结构体
--------------------------------

.. cpp:struct:: YR::InstanceRange

    配置函数实例数量范围，用于 ``Range`` 调度。

    .. note::
    
        ``Range`` 调度是一种原子资源分配策略。
        
        基于用户配置的 ``max`` 字段，当前调度实例数被设置为 now = ``max``。首先尝试调度 now 个实例，如果成功，则返回已调度的实例。\
        如果不成功，则将 now 减少 ``step`` 个实例（最多减少到 ``now - step`` 和 ``min`` 中的较大值），并再次尝试调度 now 个实例。\
        此过程将持续进行，直到调度成功、超时或所有尝试均失败（如果在降至 ``min`` 后调度仍失败，则视为失败）。

    **公共成员**
 
    .. cpp:member:: int min = DEFAULT_INSTANCE_RANGE_NUM
 
       允许的最小函数实例数量，默认值为 -1。
  
    .. cpp:member:: int max = DEFAULT_INSTANCE_RANGE_NUM
 
       允许的最大函数实例数量，默认值为 -1。
 
       ``min`` 和 ``max`` 全部配置为 -1 时，``Range`` 调度功能被禁用。当 1 < ``min`` <= ``max`` 时，启用 ``Range`` 调度功能。其他值均被视为无效，并抛出异常。
 
    .. cpp:member:: int step = DEFAULT_INSTANCE_RANGE_STEP
 
       ``max`` 向 ``min`` 递减的步长，默认值为 2。
 
       当启用 ``Range`` 调度时，``step`` 必须大于 0，否则抛出异常。如果 ``step`` 大于 ``max`` 与 ``min`` 的差，则直接从 ``max`` 递减到 ``min``。

    .. cpp:member:: bool sameLifecycle = true
 
       指定为 ``Range`` 调度配置的一组实例是否共享相同的生命周期，默认值为 true。

    .. cpp:member:: RangeOptions rangeOpts
 
       定义 ``Range`` 调度的生命周期参数，包括在内核资源不足时逐步调度的总超时时间。
    
.. cpp:struct:: YR::RangeOptions

    InstanceRange 结构体中的配置，它定义了 ``Range`` 调度的生命周期参数。

    .. warning::
    
       1. 单个 Range 在一个组内最多可以创建 256 个实例。

       2. 并发创建最多支持 12 个 group ，每个 group 最多成组创建 256 个实例。

       3. ``invoke()`` 接口在 `NamedInstance::Export()` 之后调用会导致当前线程卡住。

       4. 不调用 ``invoke()`` 接口，直接向有状态函数实例发起函数请求并 get 结果会导致当前线程卡住。

       5. 重复调用 ``invoke()`` 接口会导致异常抛出。

       6. Range 内的实例不支持独立指定生命周期。

    **公共成员**
 
    .. cpp:member:: int timeout = NO_TIMEOUT
 
       当 Eugeo 内核资源不足时，逐步调度的总超时时间，单位为秒。

       -1 表示无超时限制，调度将一直持续到成功或所有尝试均失败为止。值小于 0 会抛出异常。
 