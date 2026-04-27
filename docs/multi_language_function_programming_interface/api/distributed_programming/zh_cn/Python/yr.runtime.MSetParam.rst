yr.runtime.MSetParam
=======================

.. py:class:: yr.runtime.MSetParam(existence: ExistenceOpt = ExistenceOpt.NX, write_mode: WriteMode = WriteMode.NONE_L2_CACHE, ttl_second: int = 0, cache_type: CacheType = CacheType.MEMORY)

    基类：``object``

    表示 `mset` 操作的参数配置类。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`cache_type <cache_type_MSP>`
         - 指定使用的存储介质类型（内存或磁盘）。
       * - :ref:`existence <existence>`
         - 是否允许 Key 被重复写入。
       * - :ref:`ttl_second <ttl_second>`
         - 配置数据的生存时间（TTL），单位为秒。
       * - :ref:`write_mode <write_mode>`
         - 设置数据的写入可靠性模式。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_MSetParam>`
         -


.. toctree::
    :maxdepth: 1
    :hidden:

    yr.runtime.MSetParam.cache_type
    yr.runtime.MSetParam.existence
    yr.runtime.MSetParam.write_mode
    yr.runtime.MSetParam.ttl_second
    yr.runtime.MSetParam.__init__




