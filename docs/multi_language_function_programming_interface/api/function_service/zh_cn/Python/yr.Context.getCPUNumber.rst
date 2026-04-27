.. _getCPUNumber:

yr.Context.getCPUNumber
------------------------

.. py:method:: Context.getCPUNumber()

    获取分配给运行中函数的 CPU 数量（CPU 数量以毫核为单位进行衡量），1 个 CPU 核心等于 1000 毫核。在函数运行时阶段，每个函数的基础资源为 200 毫核，
    并根据分配给函数的内存大小增加 CPU 资源。增加量约为 内存大小(M)/128 * 100。

    返回：
        函数占用的 CPU 资源。
