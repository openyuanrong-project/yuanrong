.. _add_labels_UInt64Counter:

yr.UInt64Counter.add_labels
------------------------------
.. py:method:: UInt64Counter.add_labels(labels: dict)-> None

    为 metrics 数据添加标签。

    参数：
        - **labels** (dict) - 一个字典，其中键和值都必须是字符串。

    异常：
        - **ValueError** -  如果标签为空或不满足数据类型要求。

    样例：
        >>> import yr
        >>>
        >>> config = yr.Config(enable_metrics=True)
        >>> yr.init(config)
        >>>
        >>> @yr.instance
        >>> class Actor:
        ...     def __init__(self):
        ...         labels = {"key1": "value1", "key2": "value2"}
        ...         self.data = yr.UInt64Counter(
        ...             name="test",
        ...             description="",
        ...             unit="ms",
        ...             labels=labels
        ...         )
        ...         self.data.add_labels({"key3": "value3"})
        ...         print("Actor init done")
        ...
        ...     def run(self):
        ...         self.data.set(5)
        ...         self.data.add_labels({"phase": "run"})
        ...         msg = f"Actor run: {self.data._UInt64Counter__uint_counter_labels}, value: {self.data.get_value()}"
        ...         print(msg)
        ...         return msg
        >>>
        >>> actor1 = Actor.options(name="actor").invoke()
        >>> print(yr.get(actor1.run.invoke()))
