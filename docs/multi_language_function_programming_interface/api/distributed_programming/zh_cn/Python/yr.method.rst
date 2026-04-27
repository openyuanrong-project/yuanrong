yr.method
=====================

.. py:function:: yr.method(*args, **kwargs)

    定义有状态函数的方法。

    参数:
        - **return_nums** (int) - 成员函数返回值数量，限制：大于 0，该参数不建议设置过大。

    返回:
        装饰后的函数。
        数据类型：FunctionType。

    异常:
        - **ValueError** - 如果传入的参数错误。
        - **TypeError** - 如果传入的参数类型错误。

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
        ...     @yr.method(return_nums=2)
        ...     def detail(self, a, b):
        ...         return a, b
        ...
        >>> ins = Instance.invoke()
        >>> res1, res2 = ins.detail.invoke(0, 1)
        >>> print("detail result1:", yr.get(res1))
        detail result1: 0
        >>> print("detail result2:", yr.get(res2))
        detail result2: 1
        >>> ins.terminate()
        >>>
        >>> yr.finalize()

