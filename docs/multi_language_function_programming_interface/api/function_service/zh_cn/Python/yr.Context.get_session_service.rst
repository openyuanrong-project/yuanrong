.. _get_session_service:

yr.Context.get_session_service
-------------------------------

.. py:method:: Context.get_session_service()

    获取 SessionService 实例，用于加载和修改当前调用的 Session 对象。

    Session 的创建、保存、锁管理全部由 runtime 自动完成。
    本方法只是获取 SessionService 实例，实际加载通过 load_session() 完成。

    返回：
        SessionService 实例。
