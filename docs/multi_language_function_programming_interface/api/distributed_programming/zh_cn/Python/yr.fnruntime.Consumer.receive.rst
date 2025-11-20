.. _receive:

yr.fnruntime.Consumer.receive
--------------------------------

.. py:method:: Consumer.receive(self, timeout_ms: int, expect_num: int = None) -> List[Element]

    消费者接收数据带有订阅功能，接收数据会等待期望个数 expectNum 个 elements，当超时时间 timeout_ms 到达或者期望个数的数据可接受到时，调用返回。

    参数：
        - **timeout_ms** (int) – 超时时间。
        - **expect_num** (int，可选) – 期望接收 elements 的个数。

    异常：
        - **RuntimeError** – 如果接收数据失败。

    返回：
        实际接收到的 elements。
        数据类型：List[Element]。
