.. _close_consumer:

yr.fnruntime.Consumer.close
--------------------------------

.. py:method:: Consumer.close(self)

    关闭 consumer，会自动触发 Ack。

    异常：
        - **RuntimeError** – 如果关闭操作失败。

