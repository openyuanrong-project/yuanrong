yr.resources
=====================

.. py:function:: yr.resources() -> List[dict]

    获取集群中节点的资源信息。

    请求资源信息时需要在 `yr.Config` 中配置 `master_addr_list`。

    返回:
        list[dict]，集群中节点的资源信息。其中 `dict` 中的关键字包括以下部分，
        
        - `id`: 节点名。
        - `capacity`: 节点总资源。
        - `allocatable`: 节点可用资源，
        - `status`: 节点状态，`0` 表示正常。

    异常:
        - **RuntimeError** - 如果从 `functionsystem master` 获取信息失败。

    样例：
        >>> import yr
        >>> yr.init()
        >>>
        >>> res = yr.resources()
        >>> print(res)
        [{'id': 'function-agent-172.17.0.2-25742','status': 0, 'capacity': {'CPU': 1000.0, 'Memory': 8192.0},
        'allocatable': {'CPU': 500.0, 'Memory': 4096.0}}]
        >>>
        >>> yr.finalize()