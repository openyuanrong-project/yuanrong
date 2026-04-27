yr.runtime.ConsistencyType
============================

.. py:class:: yr.runtime.ConsistencyType

    基类：``Enum``

    一致性类型。在分布式场景中，可以配置不同的一致性语义级别。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 40 60

       * - :ref:`PRAM <pram_ct>`
         - 弱一致性。
       * - :ref:`CAUSAL <causal_ct>`
         - 因果一致性。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.runtime.ConsistencyType.PRAM
    yr.runtime.ConsistencyType.CAUSAL