yr.save_state
===============

.. py:function:: yr.save_state(timeout_sec: int = _DEFAULT_SAVE_LOAD_STATE_TIMEOUT) -> None

    保存有状态函数实例的状态。

    参数:
        - **timeout_sec** (int，可选) - 以秒为单位的超时时间。默认为 ``30``。

    返回:
        None。

    异常:
        - **RuntimeError** - 如果 `save_state` 在本地模式下被调用，将引发一个异常，消息为：`在本地模式下未调用save_state`。
        - **RuntimeError** - 如果在云模式下调用 `save_state` 而不使用 POSIX API，会引发异常，消息为：`save_state仅在具有POSIX API的云上调用`。
        - **RuntimeError** - 如果 `save_state` 检索保存的实例状态失败，会引发异常，消息为：`Failed to save state`。

    样例：
        >>> @yr.instance
        ... class Counter:
        ...     def __init__(self):
        ...         self.cnt = 0
        ...
        ...     def add(self):
        ...         self.cnt += 1
        ...         return self.cnt
        ...
        ...     def get(self):
        ...         return self.cnt
        ...
        ...     def save(self, timeout=30):
        ...         yr.save_state(timeout)
        ...
        ...     def load(self, timeout=30):
        ...         yr.load_state(timeout)
        ...
        >>> counter = Counter.invoke()
        >>> print(f"member value before save state: {yr.get(counter.get.invoke())}")
        >>> counter.save.invoke()
        >>>
        >>> counter.add.invoke()
        >>> print(f"member value after add one: {yr.get(counter.get.invoke())}")
        >>>
        >>> counter.load.invoke()
        >>> print(f"member value after load state(back to 0): {yr.get(counter.get.invoke())}")