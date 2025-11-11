.. _set_DoubleCounter:

yr.DoubleCounter.set
------------------------------
.. py:method:: DoubleCounter.set(value: float)-> None

    将双精度计数器设置为给定的值。

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
        >>> class Actor:
        >>>     def __init__(self):
        >>>         try:
        >>>             self.data = yr.DoubleCounter(
        >>>                 "userFuncTime",
        >>>                 "user function cost time",
        >>>                 "ms",
        >>>                 {"runtime": "runtime1"}
        >>>            )
        >>>         except Exception as err:
        >>>             print(f"error: {err}")
        >>>
        >>>    def run(self, *args, **kwargs):
        >>>        try:
        >>>            self.data.set(1)
        >>>        except Exception as err:
        >>>            print(f"error: {err}")
