yr.load_state
=========================

.. py:function:: yr.load_state(timeout_sec: int = 30) -> None

    导入有状态函数实例的状态。

    参数:
        - **timeout_sec** (int，可选) - 超时时间，单位：秒。默认为 ``30``。

    返回：
        None。

    异常:
        - **RuntimeError** - 如果本地模式不支持 `load_state`，会抛出异常 `load_state is not supported in local mode`。
        - **RuntimeError** - 如果集群模式仅支持在远端代码中使用，本地代码会抛出 `load_state is only supported on cloud with posix api` 异常。
        - **RuntimeError** - 如果 `load_state` 获取不到已保存的实例状态，会抛出 `Failed to load state` 异常。

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
