yr.Config
==========================

.. py:class:: yr.Config(function_id: str = '', cpp_function_id: str = '', server_address: str = '', ds_address: str = '', is_driver: bool = True, log_level: str | int = 'WARNING', invoke_timeout: int = 900, local_mode: bool = False, code_dir: str = '', connection_nums: int = 100, recycle_time: int = 2, job_id: str = '', tls_config: ~yr.config.UserTLSConfig | None = None, auto: bool = False, deployment_config: ~yr.config.DeploymentConfig | None = None, rt_server_address: str = '', log_dir: str = './', log_file_size_max: int = 0, log_file_num_max: int = 0, log_flush_interval: int = 5, runtime_id: str = 'driver', max_task_instance_num: int = -1, load_paths: list = <factory>, rpc_timeout: int = 1800, enable_mtls: bool = False, private_key_path: str = '', certificate_file_path: str = '', verify_file_path: str = '', server_name: str = '', ns: str = '', enable_metrics: bool = False, custom_envs: ~typing.Dict[str, str] = <factory>, master_addr_list: list = <factory>, working_dir: str = '', enable_ds_encrypt: bool = False, ds_public_key_path: str = '', runtime_public_key_path: str = '', runtime_private_key_path: str = '')

    基类：``object``

    init 接口使用的配置参数。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`auto <auto_cf>`
         - 在 `yr.init` 时自动启动分布式执行器，并在 `yr.finalize` 时自动停止分布式执行器。
       * - :ref:`certificate_file_path <certificate_file_path>`
         - 客户端证书文件路径。
       * - :ref:`code_dir <code_dir_cf>`
         - 需要在 runtime 初始化时设置的代码目录路径。
       * - :ref:`connection_nums <connection_nums>`
         - HTTP 客户端连接数。
       * - :ref:`cpp_function_id <cpp_function_id>`
         - 部署的 C++ 函数 ID，默认通过环境变量 `YR_CPP_FUNCID` 获取。
       * - :ref:`deployment_config <deployment_config>`
         - 当 `auto=True` 时，用于定义部署细节。
       * - :ref:`ds_address <ds_address>`
         - 数据系统地址，默认通过环境变量 `YR_DS_ADDRESS` 获取。
       * - :ref:`ds_public_key_path <ds_public_key_path>`
         - 工作进程公钥路径，用于数据系统 TLS 认证。
       * - :ref:`enable_ds_encrypt <enable_ds_encrypt>`
         - 是否启用数据系统 TLS 认证。
       * - :ref:`enable_metrics <enable_metrics>`
         - 是否启用指标收集。
       * - :ref:`enable_mtls <enable_mtls>`
         - 是否启用客户端双向认证，默认 ``False``。
       * - :ref:`function_id <function_id_cf>`
         - 部署的函数 ID，默认通过环境变量 `YRFUNCID` 获取。
       * - :ref:`invoke_timeout <invoke_timeout_cf>`
         - HTTP 客户端读取超时时间（秒），默认 ``900``。
       * - :ref:`is_driver <is_driver_cf>`
         - 仅在运行时初始化时为 ``False``，默认 ``True``。
       * - :ref:`job_id <job_id_cf>`
         - 由 `init` 自动生成。
       * - :ref:`local_mode <local_mode_cf>`
         - 在本地运行代码，默认 ``False``。
       * - :ref:`log_dir <log_dir_cf>`
         - 日志目录，指定日志文件存储的路径。
       * - :ref:`log_file_num_max <log_file_num_max>`
         - 日志文件的最大数量，默认 ``0``。
       * - :ref:`log_file_size_max <log_file_size_max>`
         - 日志文件的最大大小，默认 ``0``。
       * - :ref:`log_flush_interval <log_flush_interval>`
         - 日志刷新间隔，默认 ``5``。
       * - :ref:`log_level <log_level_cf>`
         - YR API 日志级别有 ``ERROR/WARNING/INFO/DEBUG``，默认 ``WARNING``。
       * - :ref:`max_task_instance_num <max_task_instance_num>`
         - 任务的最大实例数。
       * - :ref:`ns <ns_cf>`
         - 命名空间，用于组织和隔离配置或资源。
       * - :ref:`private_key_path <private_key_path>`
         - 客户端私钥文件路径。
       * - :ref:`recycle_time <recycle_time_cf>`
         - 实例回收周期（秒）。
       * - :ref:`rpc_timeout <rpc_timeout_cf>`
         - RPC 超时时间（秒）。
       * - :ref:`rt_server_address <rt_server_address>`
         - runtime 服务地址，在 driver 中保持默认值。
       * - :ref:`runtime_id <runtime_id_cf>`
         - 运行时 ID，在驱动程序中保持默认值。
       * - :ref:`runtime_private_key_path <runtime_private_key_path>`
         - 客户端私钥路径，用于数据系统 TLS 认证。
       * - :ref:`runtime_public_key_path <runtime_public_key_path>`
         - 客户端公钥路径，用于数据系统 TLS 认证。
       * - :ref:`server_address <server_address_cf>`
         - 系统集群地址，默认通过环境变量 `YR_SERVER_ADDRESS` 获取。
       * - :ref:`server_name <server_name_cf>`
         - 服务器名称，用于识别和连接到特定的服务器实例。
       * - :ref:`tls_config <tls_config_cf>`
         - 用于外部集群的 HTTPS SSL。
       * - :ref:`verify_file_path <verify_file_path>`
         - 服务器证书文件路径。
       * - :ref:`working_dir <working_dir_cf>`
         - 指定用户代码或其依赖的本地路径位置，必须是绝对路径，并确保在集群的所有节点上都存在。
       * - :ref:`load_paths <load_paths_cf>`
         - 代码加载路径。
       * - :ref:`custom_envs <custom_envs_cf>`
         - 用于为运行时设置自定义环境变量。
       * - :ref:`master_addr_list <master_addr_list>`
         - 函数主节点地址列表。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_cf>`
         - 

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.Config.auto
    yr.Config.certificate_file_path
    yr.Config.code_dir
    yr.Config.connection_nums
    yr.Config.cpp_function_id
    yr.Config.deployment_config
    yr.Config.ds_address
    yr.Config.ds_public_key_path
    yr.Config.enable_ds_encrypt
    yr.Config.enable_metrics
    yr.Config.enable_mtls
    yr.Config.function_id
    yr.Config.invoke_timeout
    yr.Config.is_driver 
    yr.Config.job_id
    yr.Config.local_mode
    yr.Config.log_dir 
    yr.Config.log_file_num_max
    yr.Config.log_file_size_max
    yr.Config.log_flush_interval
    yr.Config.log_level
    yr.Config.max_task_instance_num
    yr.Config.ns
    yr.Config.private_key_path
    yr.Config.recycle_time
    yr.Config.rpc_timeout 
    yr.Config.rt_server_address
    yr.Config.runtime_id
    yr.Config.runtime_private_key_path
    yr.Config.runtime_public_key_path
    yr.Config.server_address
    yr.Config.server_name
    yr.Config.tls_config 
    yr.Config.verify_file_path
    yr.Config.working_dir
    yr.Config.load_paths
    yr.Config.custom_envs
    yr.Config.master_addr_list
    yr.Config.__init__