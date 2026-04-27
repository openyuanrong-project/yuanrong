yr.list_named_instances
=========================

.. py:function:: yr.list_named_instances() -> None

    列出所有的具名实例。

    .. note::
        该接口不支持查询指定命名空间下的具名实例，如实例配置了命名空间，会通过 `-` 连接命名空间和实例名称。

    异常：
        - **RuntimeError** - 如果具名实例查询失败，会抛出该异常。

    返回：
        None。

    样例：
        >>> import yr
        >>> yr.init()
        >>> named_instances = yr.list_named_instances()
        >>> print(named_instances)
        >>> yr.finalize()
