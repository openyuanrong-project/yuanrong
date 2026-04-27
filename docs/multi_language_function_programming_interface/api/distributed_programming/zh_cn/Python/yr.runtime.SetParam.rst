yr.runtime.SetParam
=======================

.. py:class:: yr.runtime.SetParam(existence: ExistenceOpt = ExistenceOpt.NONE, write_mode: WriteMode = WriteMode.NONE_L2_CACHE, ttl_second: int = 0, cache_type: CacheType = CacheType.MEMORY)

    基类：``object``

    设置参数。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`cache_type <cache_type_setparam>`
         - 标识分配的存储介质类型（内存或磁盘）。
       * - :ref:`existence <existence_cp>`
         - 表示是否支持 Key 重复写入。
       * - :ref:`ttl_second <ttl_second_cp>`
         - 数据生命周期，超过会被删除。
       * - :ref:`write_mode <write_mode_sp>`
         - 设置数据的可靠性。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_SetParam>`
         -




.. toctree::
    :maxdepth: 1
    :hidden:

    yr.runtime.SetParam.cache_type
    yr.runtime.SetParam.existence
    yr.runtime.SetParam.ttl_second
    yr.runtime.SetParam.write_mode
    yr.runtime.SetParam.__init__
    


