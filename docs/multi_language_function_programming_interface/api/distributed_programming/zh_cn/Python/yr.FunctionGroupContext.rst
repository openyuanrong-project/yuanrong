yr.FunctionGroupContext
=======================

.. py:class:: yr.FunctionGroupContext(rank_id: int = 0, world_size: int = 0, server_list: List[yr.config.ServerInfo] = <factory>, device_name: str = '')

    基类：``object``

    用于管理函数组信息的上下文。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`device_name <device_name>`
         - 挂载到此函数实例的设备的名称，例如，NPU/Ascend910B。
       * - :ref:`rank_id <rank_id>`
         - 本函数实例在函数组中的 id。
       * - :ref:`world_size <world_size>`
         - 函数组中函数实例数。
       * - :ref:`server_list <server_list>`
         - 函数组中每个函数实例集合通信的 server 信息列表。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_FunctionGroupContext>`
         -

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.FunctionGroupContext.rank_id
    yr.FunctionGroupContext.world_size
    yr.FunctionGroupContext.server_list
    yr.FunctionGroupContext.device_name
    yr.FunctionGroupContext.__init__

