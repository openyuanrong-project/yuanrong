.. _get_session_id:

yr.Context.get_session_id
--------------------------

.. py:method:: Context.get_session_id()

    获取当前 session ID。

    Session ID 从请求头 X-Instance-Session 字段中解析。字段值格式为 {"sessionID":"xxx","sessionTTL":xx,"concurrency":xx}，
    本方法提取其中的 sessionID 值。若请求未携带该字段，则返回空字符串。

    返回：
        session ID 字符串，若无则为空字符串。
