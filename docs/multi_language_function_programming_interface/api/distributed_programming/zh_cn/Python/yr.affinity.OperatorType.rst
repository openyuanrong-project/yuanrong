yr.affinity.OperatorType
============================

.. py:class:: yr.affinity.OperatorType(value)

    基类：``Enum``

    枚举类型，定义了可用于亲和配置的不同标签操作符类型。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`LABEL_IN <LABEL_IN>`
         - 资源存在指定键和值在特定列表中的标签。
       * - :ref:`LABEL_NOT_IN <LABEL_NOT_IN>`
         - 资源存在指定键，但其值不在指定列表中。
       * - :ref:`LABEL_EXISTS <LABEL_EXISTS>`
         - 资源存在指定键的标签。
       * - :ref:`LABEL_NOT_EXISTS <LABEL_NOT_EXISTS>`
         - 资源存在指定键，但其值不在指定列表中。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.affinity.OperatorType.LABEL_IN
    yr.affinity.OperatorType.LABEL_NOT_IN
    yr.affinity.OperatorType.LABEL_EXISTS
    yr.affinity.OperatorType.LABEL_NOT_EXISTS


