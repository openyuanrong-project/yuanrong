.. _send:

yr.fnruntime.Producer.send
-----------------------------

.. py:method:: Producer.send(self, element: Element, int timeout_ms: int = None) -> None

    Producer 发送数据，数据会首先放入缓冲中，根据配置的自动 flush 策略（发送间隔一段时间或者缓冲写满）去刷新缓冲，或者通过手动调用 flush 操作让数据对消费者可见。
    
    参数：
        - **element** (Element_) – 要发送的 Element 数据。
        - **timeout_ms** (int, optional) – 超时时间，以毫秒为单位。默认值为 ``None``。
        
    异常：
        - **RuntimeError** – 如果发送操作失败。
        
    .. _Element: ../../Python/generated/yr.Element.html#yr.Element