yr.config.GroupOptions
==================================

.. py:class:: yr.config.GroupOptions(timeout: int = -1, same_lifecycle: bool = True, strategy: str = '')

    基类：``object``

    分组实例调度的配置选项。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`timeout <timeout_go>`
         - 当内核资源不足时，重新调度的超时时间，单位为秒。如果设置为 `-1`，内核将无限次尝试调度。如果设置为小于 `0` 的其他值，将抛出异常。默认值：``-1``。
       * - :ref:`same_lifecycle <same_lifecycle>`
         - 是否启用分组实例的命运共享配置。`True`（默认值）：组内实例将一起创建和销毁。`False`：实例可以拥有独立的生命周期。默认值：``True``。
       * - :ref:`strategy <strategy>`
         - 创建实例组的策略。None：不使用策略。SPREAD：尽可能将多个实例分散到不同节点上。STRICT_PACK：所有实例必须放置在同一节点，否则创建失败。PACK：尽可能将多个实例打包到同一节点。STRICT_SPREAD：所有实例必须放置在不同节点，否则创建失败。默认值：``None``。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_GroupOptions>`
         -

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.config.GroupOptions.__init__
    yr.config.GroupOptions.timeout
    yr.config.GroupOptions.same_lifecycle
    yr.config.GroupOptions.strategy

