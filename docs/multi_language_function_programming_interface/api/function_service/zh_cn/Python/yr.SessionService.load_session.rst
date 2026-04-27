.. _load_session:

yr.SessionService.load_session
-------------------------------

.. py:method:: SessionService.load_session(session_id)

    加载当前请求关联的会话对象。

    参数：
        - **session_id** (str) - 会话 ID。

    返回：
        Session 实例，如果未找到会话或 ``enable_agent_session=false`` 则返回 ``None``。
        数据类型：yr.Session。
