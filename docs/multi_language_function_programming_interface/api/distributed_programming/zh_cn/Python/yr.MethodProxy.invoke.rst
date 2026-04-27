.. _invoke_mp:

yr.MethodProxy.invoke
------------------------------------------------

.. py:method:: MethodProxy.invoke(*args, **kwargs) -> "yr.ObjectRef"

    执行用户函数的远程调用。

    参数：
        - **\*args** - 可变参数，用于传递非关键字参数。
        - **\*\*kwargs** - 可变参数，用于传递关键字参数。

    返回：
        数据对象的引用。数据类型：ObjectRef。

    异常：
        - **TypeError** - 如果参数类型不正确。

    样例：
        >>> import yr
        >>> yr.init()
        >>>
        >>> @yr.instance
        ... class Instance:
        ...     sum = 0
        ...
        ...     def add(self, a):
        ...         self.sum += a
        ...
        ...     def get(self):
        ...         return self.sum
        ...
        >>> ins = Instance.invoke()
        >>> yr.get(ins.add.invoke(1))
        >>>
        >>> print(yr.get(ins.get.invoke()))
        >>>
        >>> ins.terminate()
        >>>
        >>> yr.finalize()
