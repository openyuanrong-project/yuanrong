yr.config.DeploymentConfig
============================

.. py:class:: yr.config.DeploymentConfig(cpu: int = 0, mem: int = 0, datamem: int = 0, spill_path: str = '', spill_limit: int = 0)

    基类：``object``

    自动部署配置类。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`cpu <deployment_cpu>`
         - 已获取的 CPU 资源，单位是 millicpu，默认值为 ``0``。
       * - :ref:`datamem <deployment_datamem>`
         - 已获取的数据系统内存资源，单位是 MB，默认值为 ``0``。
       * - :ref:`mem <deployment_mem>`
         - 已获取的内存资源，单位是 MB，默认值为 ``0``。
       * - :ref:`spill_limit <spill_limit>`
         - 磁盘溢出大小限制，单位是 MB，默认值为 ``0``。
       * - :ref:`spill_path <spill_path>`
         - 当内存不足时，数据将被刷新到磁盘的路径，默认值为空字符串。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <deployment_init>`
         -

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.config.DeploymentConfig.cpu
    yr.config.DeploymentConfig.datamem
    yr.config.DeploymentConfig.mem
    yr.config.DeploymentConfig.spill_limit
    yr.config.DeploymentConfig.spill_path
    yr.config.DeploymentConfig.__init__