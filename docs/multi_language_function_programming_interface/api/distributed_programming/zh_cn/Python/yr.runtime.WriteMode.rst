yr.runtime.WriteMode
=========================

.. py:class:: yr.runtime.WriteMode

    基类：``Enum``

    对象写入模式。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 40 60

       * - :ref:`NONE_L2_CACHE <none_l2_cache_wm>`
         - 不写入二级缓存。
       * - :ref:`WRITE_THROUGH_L2_CACHE <write_through_l2_cache_wm>`
         - 同步写数据到二级缓存，保证数据可靠性。
       * - :ref:`WRITE_BACK_L2_CACHE <write_back_l2_cache_wm>`
         - 异步写数据到二级缓存，保证数据可靠性。
       * - :ref:`NONE_L2_CACHE_EVICT <none_l2_cache_evict_wm>`
         - 不写入二级缓存，并且当系统资源不足时，会被系统主动驱逐。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.runtime.WriteMode.NONE_L2_CACHE
    yr.runtime.WriteMode.WRITE_THROUGH_L2_CACHE
    yr.runtime.WriteMode.WRITE_BACK_L2_CACHE
    yr.runtime.WriteMode.NONE_L2_CACHE_EVICT