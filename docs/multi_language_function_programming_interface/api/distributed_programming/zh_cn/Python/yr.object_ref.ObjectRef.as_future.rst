.. _as_future:

yr.object_ref.ObjectRef.as_future
------------------------------------------------

.. py:method:: ObjectRef.as_future() -> Future

    使用 asyncio.Future 包装 ObjectRef。

    需要注意的是，当 ObjectRef 表示任务的返回对象时，Future 的取消不会取消对应的任务。

    返回：
        asyncio.Future 类型的对象，用于包装 ObjectRef。
