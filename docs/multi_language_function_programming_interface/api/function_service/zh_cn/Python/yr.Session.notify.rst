.. _notify:

yr.Session.notify
-------------------------------

.. py:method:: Session.notify(payload)

    唤醒正在 `wait_for_notify` 状态的线程，并将 `payload` 传递给它。

    参数：
        - **payload** (dict) - 要传递给等待线程的数据。

    返回：
        无。
