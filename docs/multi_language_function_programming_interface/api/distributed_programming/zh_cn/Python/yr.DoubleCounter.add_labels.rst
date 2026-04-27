.. _add_labels_DoubleCounter:

yr.DoubleCounter.add_labels
------------------------------
.. py:method:: DoubleCounter.add_labels(labels: dict)-> None

    为 metrics 数据添加标签。

    参数：
        - **value** (float) - 要添加的标签字典。

    异常：
        - **ValueError** - 如果标签为空。

    样例：
        >>> import yr
        >>>
        >>> config = yr.Config(enable_metrics=True)
        >>> yr.init(config)
        >>>
        >>> @yr.instance
        >>> class Actor:
        >>>     def __init__(self):
        >>>         self.data = yr.DoubleCounter(
        >>>             "userFuncTime",
        >>>             "user function cost time",
        >>>             "ms",
        >>>             {"runtime": "runtime1"}
        >>>         )
        >>>         self.data.add_labels({"stage": "init"})
        >>>         print("Actor init:", self.data._DoubleCounter__double_counter_labels)
        >>>
        >>>     def run(self):
        >>>         self.data.set(5)
        >>>         self.data.add_labels({"phase": "run"})
        >>>         msg = f"Actor run: {self.data._DoubleCounter__double_counter_labels},
        >>>         value: {self.data.get_value()}"
        >>>         print(msg)
        >>>         return msg
        >>> actor1 = Actor.options(name="actor").invoke()
        >>> result = actor1.run.invoke()
        >>> print("run result:", yr.get(result))


