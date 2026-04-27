.. _set_UInt64Counter:

yr.UInt64Counter.set
------------------------------
.. py:method:: UInt64Counter.set(value: int)-> None

     将 uint64 计数器设置为给定的值。

     参数：
        - **value** (float) - 要设置的值。

    异常：
        - **ValueError** - 如 driver 中调用。

    样例：
        >>> import yr
        >>> config = yr.Config(enable_metrics=True)
        >>> yr.init(config)
        >>>
        >>> @yr.instance
        ... class MyActor:
        ...     def __init__(self):
        ...         self.data = yr.UInt64Counter(
        ...             "userFuncTime",
        ...             "user function cost time",
        ...             "ms",
        ...             {"runtime": "runtime1"}
        ...         )
        ...         self.data.add_labels({"stage": "init"})
        ...         self.data.set(100)
        ...
        ...     def add(self, n=1, labels=None):
        ...         current = self.data.get_value()
        ...         self.data.set(current + n)
        ...         if labels:
        ...             self.data.add_labels(labels)
        ...         return self.get()
        ...
        ...     def get(self):
        ...         return {
        ...             "labels": self.data._UInt64Counter__uint_counter_labels,
        ...             "value": self.data.get_value()
        ...         }
        >>> actor = MyActor.options(name="actor").invoke()
        >>> result = actor.add.invoke(5)
        >>> print(yr.get(result))
