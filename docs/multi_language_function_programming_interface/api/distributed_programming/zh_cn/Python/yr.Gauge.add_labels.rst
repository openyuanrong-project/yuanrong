.. _add_labels:

yr.Gauge.add_labels
--------------------

.. py:method:: Gauge.add_labels(labels: Dict[str, str]) -> None

    添加 label。

    参数：
        - **labels** (Dict[str,str]) - the labels，kv 均要求为字符串，key不能以 \* 等特殊字符开头。

    异常：
        - **ValueError** - 无 label，或不满足 promethues 数据标准要求。

    样例：
        >>> import yr
        >>> config = yr.Config(enable_metrics=True)
        >>> yr.init(config)

        >>> @yr.instance
        ... class GaugeActor:
        ...     def __init__(self):
        ...         self.labels = {"key1": "value1"}
        ...         self.gauge = yr.Gauge(
        ...             name="test",
        ...             description="",
        ...             unit="ms"
        ...         )
        ...         self.gauge.add_labels(self.labels)
        ...         self.gauge.set(50)
        ...         print("Initial labels:", self.labels)
        ...
        ...     def set_value(self, value):
        ...         self.gauge.set(value)
        ...         return {
        ...             "value": value,
        ...             "labels": self.labels
        ...         }
        >>>
        >>> actor = GaugeActor.options(name="gauge_actor").invoke()
        >>> result = actor.set_value.invoke(75)
        >>> print("Driver got:", yr.get(result))

