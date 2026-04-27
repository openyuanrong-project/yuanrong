.. _get_ObjectRef:

yr.object_ref.ObjectRef.get
------------------------------

.. py:method:: ObjectRef.get(timeout: int = 300) -> Any

    此函数用于 FaaS 模式中检索对象。

    参数：
        - **timeout**  (int，可选) - 等待检索接口对象的最大时间，单位为秒。默认为 ``300``。

    返回：
        检索到的对象。数据类型为任意类型。

    异常：
        - **ValueError** - 如果超时参数小于或等于 0 且不是 -1。
        - **YRInvokeError** - 如果检索结果是 `YRInvokeError` 的实例。
