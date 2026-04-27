.. _real_id:

yr.InstanceProxy.real_id
-------------------------------------------------------

.. py:property:: InstanceProxy.real_id

    运行时分配给实例的真实 ID。

    ``instance_id`` 是内部使用的逻辑键。该属性会阻塞等待，直至运行时完成 Actor 的调度（默认超时时间为 30 秒），
    然后将其解析为物理实例 ID。

    抛出异常：
        - **TimeoutError** - 若 Actor 在 30 秒内未就绪，则抛出超时异常。

    返回：
        真实实例 ID。数据类型：str。

    示例：

    .. code-block:: python

        ins = MyActor.invoke()
        print(ins.real_id)
