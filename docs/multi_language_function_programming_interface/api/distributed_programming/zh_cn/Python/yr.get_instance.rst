yr.get_instance
=====================

.. py:function:: yr.get_instance(name: str, namespace: str = '', timeout: int = 60) -> InstanceProxy

    根据具名实例的 `name` 和 `namespace` 获取实例句柄。接口调用后会阻塞直到获取到实例句柄或者超时。 

    参数：
        - **name** (str) – 实例名（实例ID）。
        - **namespace** (str) – 命名空间。
        - **timeout** (int) – 超时时间，单位为秒。

    返回：
        一个 Python 具名实例句柄。
        数据类型：InstanceProxy。

    异常：
        - **TypeError** – 参数类型有误。
        - **RuntimeError** – 实例不存在。
        - **TimeoutError** – 超时错误。
	
    样例：
        >>> yr.get_instance("name")

.. _InstanceProxy: ../../zh_cn/Python/yr.decorator.instance_proxy.InstanceProxy.html#yr.decorator.instance_proxy.InstanceProxy