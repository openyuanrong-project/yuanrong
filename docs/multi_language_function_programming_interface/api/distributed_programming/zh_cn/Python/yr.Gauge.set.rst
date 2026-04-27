.. _set_Gauge:

yr.Gauge.set
--------------------

.. py:method:: Gauge.set(value: float) -> None

    设置值。

    参数：
        - **value** (float) - 目标值。

    异常：
        - **ValueError** - 如 driver 中调用，或不满足数据类型要求（比如上报非 float 数据）。
        - **Exception** - 其他异常。

    样例：
        >>> import yr
        >>> config = yr.Config(enable_metrics=True)
        >>> yr.init(config)
        >>>
        >>> @yr.instance
        ... class GaugeActor:
        ...     def __init__(self):
        ...         self.labels = {"key1": "value1"}
        ...         self.gauge = yr.Gauge(
        ...             name="test",
        ...             description="",
        ...             unit="ms",
        ...             labels=self.labels
        ...         )
        ...         self.gauge.set(50)
        ...
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
        >>>
        >>> result = actor.set_value.invoke(75)
        >>> print(yr.get(result))

