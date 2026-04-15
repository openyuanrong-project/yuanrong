.. _wait_for_notify:

yr.Session.wait_for_notify
-------------------------------

.. py:method:: Session.wait_for_notify(timeout_ms)

    阻塞当前协程/线程，等待同一个会话的后续输入。在等待期间，当前线程会释放会话锁，允许其他请求（如 `notify`）进入。

    参数：
        - **timeout_ms** (int) - 等待超时时间（毫秒）。``-1`` 表示无限等待。

    返回：
        接收到的输入数据，如果超时则返回 ``None``。
        数据类型：dict。
