.. _Alarm_set:

yr.Alarm.set
--------------------

.. py:method:: Alarm.set(alarm_info: AlarmInfo) -> None

    设置告警信息。

    参数：
        - **value** (float) - 包含告警详细信息的对象。

    异常：
        - **ValueError** - 如 driver 中调用。
        - **ValueError** - 如果 alarm_name 值为空。

    样例：
        >>> import yr
        >>> import time
        >>> config = yr.Config(enable_metrics=True)
        >>> config.log_level="DEBUG"
        >>> yr.init(config)

        >>> @yr.instance
        >>> class Actor:
        >>>    def __init__(self):
        >>>        self.state = 0
        >>>        self.name = "aa"

        >>>    def add(self, value):
        >>>        self.state += value
        >>>        if self.state > 10:
        >>>            alarm_info = yr.AlarmInfo(alarm_name="aad")
        >>>            alarm = yr.Alarm(self.name, description="")
        >>>            alarm.set(alarm_info)
        >>>        return self.state
        >>>
        >>>    def get(self):
        >>>        return self.state
        >>>
        >>> actor1 = Actor.options(name="actor1").invoke()
        >>>
        >>> print("actor1 add 5:", yr.get(actor1.add.invoke(5)))
        >>> print("actor1 add 7:", yr.get(actor1.add.invoke(7)))
        >>> print("actor1 state:", yr.get(actor1.get.invoke()))

