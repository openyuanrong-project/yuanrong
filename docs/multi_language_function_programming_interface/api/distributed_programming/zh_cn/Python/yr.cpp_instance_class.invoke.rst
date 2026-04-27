.. _invoke:

yr.cpp_instance_class.invoke
-----------------------------

.. py:method:: cpp_instance_class.invoke(*args, **kwargs) -> InstanceProxy

    创建一个 cpp class 的实例。

    参数：
        - **\*args** (tuple) – cpp 类名。
        - **\**kwargs** (dict) – cpp 类静态构造函数名。

    返回：
        对应的 cpp 实例句柄。

