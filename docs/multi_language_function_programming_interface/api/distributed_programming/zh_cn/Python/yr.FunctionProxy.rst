yr.FunctionProxy
===========================================

.. py:class:: yr.FunctionProxy

    基类：``object``

    无状态函数句柄，使用它执行调用。

    **样例**：
        >>> import yr
        >>>
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def add(a, b):
        ...     return a + b
        >>>
        >>> ret = add.invoke(1, 2)
        >>> print(yr.get(ret))
        >>>
        >>> yr.finalize()

    **方法**：

    .. list-table::
       :widths: 40 60
       :header-rows: 0

       * - :ref:`__init__ <init_fp>`
         - 
       * - :ref:`get_original_func <get_original_func>`
         - 获取未包装的函数。
       * - :ref:`invoke <invoke_fp>`
         - 执行被装饰的远程函数。
       * - :ref:`options <options_fp>`
         - 动态修改被装饰函数的调用参数。
       * - :ref:`set_function_group_size <set_function_group_size>`
         - 设置函数组大小。
       * - :ref:`set_urn <set_urn>`
         - 设置函数 urn。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.FunctionProxy.__init__
    yr.FunctionProxy.get_original_func
    yr.FunctionProxy.invoke
    yr.FunctionProxy.options
    yr.FunctionProxy.set_function_group_size
    yr.FunctionProxy.set_urn