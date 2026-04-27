.. _dedup_logs_cn:

yr.Config.dedup_logs
------------------------------------

.. py:attribute:: config.dedup_logs
   :type: bool
   :value: True

   默认值为 ``True``，用于对作业中函数进程的标准输出日志去重。第一次出现的日志模式会立即打印，之后相同模式的日志最多缓存 5 秒后批量打印。