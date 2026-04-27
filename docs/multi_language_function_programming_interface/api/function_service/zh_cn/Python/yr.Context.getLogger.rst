.. _getLogger:

yr.Context.getLogger
---------------------

.. py:method:: Context.getLogger()

    获取供用户在标准输出中打印日志的 logger，Logger 接口必须在 SDK 中提供。

    返回：
        logger。

    样例：
        >>> log = context.getLogger()
        >>> log.info("test")
