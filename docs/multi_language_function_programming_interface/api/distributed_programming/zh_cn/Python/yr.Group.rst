yr.Group
===========================================

.. py:class:: yr.Group

    基类：``object``

    一个用于管理实例分组生命周期的类。

    `Group` 类负责管理一组实例的完整生命周期，包括它们的创建与销毁。
    该类遵循“命运共享（fate-sharing）”原则，即组内的所有实例要么一起创建，
    要么一起销毁。

    `Group` 类提供了用于创建、终止、挂起（suspend）、恢复（resume）以及管理
    分组实例的方法。它确保组内的所有实例被视为一个整体进行管理，并且在组创建
    过程中如果发生任何失败，都会回滚整个实例组。

    样例：
        >>> import yr
        >>>
        >>> yr.init()
        >>>
        >>> @yr.instance
        ... class Counter:
        ...     sum = 0
        ...
        ...     def add(self, a):
        ...         self.sum += a
        ...
        >>> group_opts = yr.GroupOptions()
        >>> group_name = "test"
        >>> g = yr.Group(group_name, group_opts)
        >>> opts = yr.InvokeOptions()
        >>> opts.group_name = group_name
        >>> ins = Counter.options(opts).invoke()
        >>> g.invoke()
        >>> res = ins.add.invoke()
        >>> print(yr.get(res))
        >>> g.terminate()
        >>>
        >>> yr.finalize()

    方法：

    .. list-table::
       :widths: 40 60
       :header-rows: 0

       * - :ref:`__init__ <init_Group>`
         -
       * - :ref:`invoke <invoke_g>`
         - 按照命运共享（fate-sharing）原则执行一组实例的创建。
       * - :ref:`terminate <terminate_g>`
         - 终止一组实例。
       * - :ref:`suspend <suspend>`
         - 挂起（暂停）由 Group 管理的所有函数实例。
       * - :ref:`resume <resume>`
         - 恢复由 Group 管理的所有函数实例。

.. toctree::
    :maxdepth: 1
    :hidden:
    yr.Group.__init__
    yr.Group.invoke
    yr.Group.terminate
    yr.Group.suspend
    yr.Group.resume