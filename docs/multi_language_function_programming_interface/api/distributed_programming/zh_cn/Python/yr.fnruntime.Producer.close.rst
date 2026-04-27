.. _close:

yr.fnruntime.Producer.close
--------------------------------

.. py:method:: Producer.close(self) -> None

    关闭生产者将触发数据缓冲区的自动刷新，并表明数据缓冲区不再使用。一旦关闭，生产者将无法再次使用。