yr.ResourceGroup
================

.. py:class:: yr.ResourceGroup(name:, request_id:, bundles: ~typing.List[~typing.Dict] | None = None)

    基类：``object``

    创建 ResourceGroup 后返回的句柄。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`bundle_count <bundle_count>`
         - 返回当前资源组中的 bundle 数量。
       * - :ref:`bundle_specs <bundle_specs>`
         - 返回当前资源组下的所有 bundle。
       * - :ref:`resource_group_name <resource_group_name_resource>`
         - 返回当前资源组的名称。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 35 70

       * - :ref:`__init__ <init_ResourceGroup>`
         - 
       * - :ref:`wait <wait_ResourceGroup>`
         - 阻塞并等待创建 ResourceGroup 的结果。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.ResourceGroup.__init__
    yr.ResourceGroup.wait
    yr.ResourceGroup.bundle_count
    yr.ResourceGroup.bundle_specs
    yr.ResourceGroup.resource_group_name


