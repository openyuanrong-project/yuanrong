yr.config.ResourceGroupOptions
==================================

.. py:class:: yr.config.ResourceGroupOptions(resource_group_name: str = '', bundle_index: int = -1)

    基类：``object``

    资源组选项。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`resource_group_name <resource_group_name>`
         - 需要调度的资源组名称。默认为空，表示不调度到任何资源组；如果不为空，则调度到指定的资源组。
       * - :ref:`bundle_index <bundle_index>`
         - 资源组名称不为空时，要调度的 bundle 的索引才生效。取值范围是 [-1, 资源组中的 bundle 数量)。默认值是 ``-1``，表示不指定具体的 bundle；如果取值范围内的值不是 ``-1``，则表示将 bundle 调度到资源组中的相应索引；如果是其他任何值，将生成错误。
         


    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_ResourceGroupOptions>`
         -


.. toctree::
    :maxdepth: 1
    :hidden:

    yr.config.ResourceGroupOptions.__init__
    yr.config.ResourceGroupOptions.resource_group_name
    yr.config.ResourceGroupOptions.bundle_index

