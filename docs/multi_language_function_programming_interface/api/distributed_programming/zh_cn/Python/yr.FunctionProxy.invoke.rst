yr.FunctionProxy.invoke
=======================================

.. py:method:: yr.FunctionProxy.invoke(*args, **kwargs)

    执行被装饰的远程函数。

    此方法触发被装饰函数在远程工作节点上的执行。函数将使用提供的参数执行，
    并返回一个 ObjectRef，可用于检索结果。

    **参数**：
        * **\*args** -- 传递给被装饰函数的变长参数。
        * **\*\*kwargs** -- 传递给被装饰函数的关键字参数。

    **返回**：
        ObjectRef 或 List[ObjectRef]：结果对象的引用。
        
        * 对于 return_nums=1 的函数，返回单个 ObjectRef。
        * 对于 return_nums>1 的函数，返回 ObjectRef 列表。  
        * 对于生成器函数，返回 ObjectRefGenerator。
        * 对于 return_nums=0 的函数，返回 None。

    **抛出**：
        * **TypeError** -- 如果提供的参数与函数签名不匹配。
        * **RuntimeError** -- 如果函数执行失败或运行时未初始化。

    **样例**：
        >>> import yr
        >>>
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def add(a, b):
        ...     return a + b
        >>>
        >>> result_ref = add.invoke(1, 2)
        >>> result = yr.get(result_ref)
        >>> print(result)  # 输出: 3
        >>>
        >>> # 对于多返回值函数
        >>> @yr.invoke(return_nums=2)
        ... def divmod_func(a, b):
        ...     return divmod(a, b)
        >>>
        >>> quotient_ref, remainder_ref = divmod_func.invoke(10, 3)
        >>> print(yr.get(quotient_ref))  # 输出: 3
        >>> print(yr.get(remainder_ref))  # 输出: 1
        >>>
        >>> yr.finalize()
