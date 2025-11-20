.. _ack:

yr.fnruntime.Consumer.ack
--------------------------------

.. py:method:: Consumer.ack(self, int element_id: int) → None

    消费者使用完某 elementId 标识的 element 后，需要确认已消费完，使得各个 worker 上可以获取到是否所有消费者都已经消费完的信息，若已经消费完某个 Page 可以触发内部的内存回收机制。
    若不 Ack，则在消费者退出时候才会自动 Ack。

    参数：
        - **element_id** (int) – 要确认的元素的ID。

    异常：
        - **RuntimeError** – 如果发送操作失败。

    返回：
        None。
