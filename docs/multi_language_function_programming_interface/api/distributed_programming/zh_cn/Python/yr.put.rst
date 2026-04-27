yr.put
==========

.. py:function:: yr.put(obj: object, create_param: CreateParam = CreateParam()) -> ObjectRef

    保存数据对象到数据系统。
	
    .. note::
        1. 此方法应在调用 `yr.init()` 之后使用。
        2. 如果 put 的类型是 memoryview， bytearray， memoryview - meta-information，此时免去了序列化。
        3. 如果传递给 put 的对象类型为 memoryview，bytearray 或 bytes，则其长度不能为 ``0``。

    参数:
        - **obj** (object) - Python 对象，被序列化并保存到数据系统中。
        - **create_param** (CreateParam()，可选) - 这是为数据系统创建对象时的参数。

    返回:
        对象的引用。
        数据类型：ObjectRef_ 。

    异常:
        - **ValueError** - 如果输入的 obj 为 ``None`` 或长度为 ``0`` 的 ``bytes``, ``bytearray`` 或 ``memoryview`` 对象。
        - **TypeError** - 如果输入的 obj 已经是一个 `ObjectRef`。
        - **TypeError** - 如果输入的 obj 不可序列化，例如 `thread.RLock`。
        - **RuntimeError** - 在调用 `yr.init()` 之前调用了 `yr.put()`。
        - **RuntimeError** - 将对象放入数据系统失败。

    样例：
        >>> import yr
        >>> yr.init()
        >>> param = yr.CreateParam()
        >>> param.cache_type = yr.CacheType.DISK
        >>> bs = bytes(0)
        >>> mem = memoryview(bytes(100))
        >>> obj_ref2 = yr.put(mem)
        >>> print(yr.get(obj_ref2))
        >>> # 最后输出一个 memoryview 指针
        >>> byte_array = bytearray(20)
        >>> obj_ref3 = yr.put(byte_array)
        >>> print(yr.get(obj_ref3))
        >>> # 最后输出一个 memoryview 指针
        >>> obj_ref4 = yr.put(100)
        >>> print(yr.get(obj_ref4))

.. _ObjectRef: ../../zh_cn/Python/yr.object_ref.ObjectRef.html#yr.object_ref.ObjectRef
