yr.instance
=====================

.. py:function:: yr.instance(*args, **kwargs) -> InstanceCreator

    定义一个有状态函数。
    
    用于装饰需要在 openYuanrong 系统上远程调用的类。

    参数:
        - **class** (class) - 需要被远程调用的类。
        - **invoke_options** (yr.InvokeOptions_) - 调用参数。

    返回:
        被装饰类的创建器。
        数据类型：InstanceCreator_ 。

    异常:
        - **RuntimeError** - 如果 `instance` 装饰的对象不是类。

    样例：
        简单调用示例：
            >>> import yr
            >>> yr.init()
            >>>
            >>> @yr.instance
            ... class Instance:
            ...     sum = 0
            ...     def add(self, a):
            ...         self.sum += a
            ...     def get(self):
            ...         return self.sum
            >>>
            >>> ins = Instance.invoke()
            >>> yr.get(ins.add.invoke(1))
            >>> print(yr.get(ins.get.invoke()))
            1
            >>> ins.terminate()
            >>>
            >>> yr.finalize()

        函数内调用示例：
            >>> import yr
            >>> yr.init()
            >>>
            >>> @yr.instance
            ... class Instance:
            ...     def __init__(self):
            ...         self.sum = 0
            ...     def add(self, a):
            ...         self.sum += a
            ...     def get(self):
            ...         return self.sum

            >>> @yr.instance
            ... class Instance2:
            ...    def __init__(self):
            ...        self.ins = Instance.invoke()
            ...    def add(self, a):
            ...        return self.ins.add.invoke(a)
            ...    def get(self):
            ...        return yr.get(self.ins.get.invoke())
            >>>
            >>> ins = Instance2.invoke()
            >>> yr.get(ins.add.invoke(2))
            >>> print(yr.get(ins.get.invoke()))
            2
            >>> ins.terminate()
            >>>
            >>> yr.finalize()

.. _yr.InvokeOptions: ../../zh_cn/Python/yr.InvokeOptions.html
.. _InstanceCreator: ../../zh_cn/Python/yr.InstanceCreator.html#yr.InstanceCreator
