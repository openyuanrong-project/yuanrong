yr.delete_stream
==========================

.. py:function:: yr.delete_stream(stream_name: str) -> None

    删除数据流。在全局生产者和消费者个数为 0 时，该数据流不再使用，清理各个 worker 上及 master 上与该数据流相关的元信息。该函数可在任意的 Host 节点上调用。
	
    参数：
        - **stream_name** (str) - 流的名称。长度必须小于 256 个字符且仅含有下列字符 `(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)` 。

    返回：
        None。

    异常：
        - **RuntimeError** - 如果数据系统中无法删除流，则会抛出此异常。

    样例：
        >>> try:
        ...     delete_stream("streamName")
        ... except RuntimeError as exp:
        ...     pass