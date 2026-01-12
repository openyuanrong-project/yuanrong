# 部署参数表

## 通用配置

| 参数名                             | 说明                       | 默认值       | 备注 |
|---------------------------------|--------------------------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-a,--ip_address`               | 本机 ip。                                        | ``本机 ip``   | 选填。若不配置，将自动获取当前机器的 ip。 |
| `-c,--cpu_num`                  | 数据面总可用 CPU（单位：``1/1000`` 核）。              | ``0``        | 选填。yr_master 默认占用 ``1`` 毫核，配置值建议大于 ``0`` （ yr start 启动时未配置该参数将默认使用所有系统 cpu 资源作为可用资源）。|
| `-d,--deploy_path`              | 部署路径（支持相对路径，或者绝对路径）。            | ``../deploy`` | 选填。建议指定目录，并确保空目录（yr start 启动时未配置该参数将默认使用 ``/tmp/yr_sessions/{current timestamp}`` 作为部署路径）。 |
| `-e,--enable_multi_master`      | 是否使能多 master 部署。                          | 无参      | 选填。开启后，etcd 集群 ip 和端口需要提前传入，一个节点最多允许拉起 1 个 master。 |
| `-m,--memory_num`               | 数据面总可用内存（单位：MB）。                       | ``0`` |选填（yr_master 默认 ``shared_memory_num + 1``）。配置值建议大于 ``shared_memory_num``。当开启数据系统流缓存功能时，需额外增加 `sc_local_cache_memory`（yr start 启动时未配置该参数将使用从 `/proc/meminfo` 采集到内存信息作为可用内存）。 |
| `-n,--node_id`                  | 数据面标识。                                      | 内部生成  | 选填。如手动配置，请保证数据面 ID 唯一。 |
| `-s,--shared_memory_num`        | 数据面数据系统可用内存（单位：MB）。                  | ``0``        | 选填，yr_master 默认 ``2048``。配置值建议大于 ``0`` （yr start 启动时未配置该参数将默认使用从 /proc/meminfo 采集到内存的 1/3 作为可用共享内存）。 |
| `-o,--master_info_output`       | 控制面组件启动 IP、端口信息输出文件路径。            | ``""``     | 选填。默认输出在部署目录下，yr_agent 不需要配置。 |
| `-w,--ds_worker_unique_enable`  | 是否打开 ds-worker 节点唯一。                      | 无参     | 选填。 |
| `-p,--services_path`            | 函数包 services.yaml 路径。                        | ``""``     | 选填。建议配置绝对路径，并确保文件存在，否则可能无法正确加载函数元数据。 |
| `-f,--function_proxy_unique_enable` | 是否打开 function-proxy 节点唯一。             | 无参      | 选填。 |
| `--master`                      | 启动 master 节点。                                | 无参      | 必填。yr_master 启动默认加此参数。 |
| `--master_info`                 | 控制面组件启动 IP、端口信息。                       | ``""``     | 部署从节点时**必填**，部署主节点时**不填**。优先级高于直接配置 etcd_ip、master_ip 等。 |
| `--master_ip`                   | 控制面组件 IP。                                    | ``""``     | 选填。若配置了 `--master`，该参数配置忽略。 |
| `--cpu_reserved`                | 数据系统 ds-worker 预留 cpu 核数。                 | ``0``         | 选填。 |
| `--status_collect_enable`       | 是否开启进程状态收集，记录日志。                    | ``false``     | 选填，取值：``true``，``false``。开启后，将定期收集进程状态信息，并记录到日志文件。 |
| `--status_collect_interval`     | 进程状态收集间隔（s）。                             | ``300``       | 选填。status_collect_enable 为 ``true`` 生效。 |
| `--enable_trace`                | 是否开启 trace。                                  | ``false``     | 选填，取值：``true``，``false``。 |
| `--runtime_trace_config`        | runtime 调用链配置，预期是 json 字符串。                 | `""`      | 选填。 |
| `--enable_metrics`              | 是否开启 metrics。                                | ``true``      | 选填，取值：``true``，``false``。 |
| `--metrics_config`              | metrics 配置，预期是 json 字符串。                 | `""`      | 选填，格式参照 metrics_config_file。 |
| `--metrics_config_file`         | metrics 配置文件，预期是绝对路径。                  | ``""``     | 选填。  |
| `--runtime_metrics_file`                                |     runtime_metrics配置                       |   ``""``              | 选填。   |
| `--pull_resource_interval`      | 设置拉取资源的周期（单位：ms）。                    | ``1000``      | 选填。 |
| `--custom_resources`            | 数据面自定义资源设置，预期是 json 字符串。           | ``""``      | 选填。 |
| `--etcd_addr_list`              | etcd 集群地址。                                    | 本机 IP   | 选填。传入格式如 ip1,ip2（端口根据 etcd_port,etcd_peer_port 参数确定）；或 ip1:port1,ip2:port2（peer 端口根据 etcd_peer_port 参数确定）；或 ip1:port1:peer_port1,ip2:port2:peer_port2。 |
| `--block`                       | 阻塞模式。                                         | ``false``     | 选填。启用阻塞模式时，部署脚本进程不会退出。 |
| `--etcd_mode`                   | etcd 部署模式。                                    | ``inner``     | 选填。取值 ``inner``，``outter``。当配置 ``outter`` 时，脚本不会启动 etcd，需要结合配置 `--etcd_addr_list` 配置 etcd 地址，并需要开启多 master，一起使用才能生效。 |
| `--etcd_table_prefix`           | etcd key 前缀配置，用于多套集群对接同一个 etcd 集群。 | ``""``      | 选填。配置不能包含空格，建议不同的集群使用不同的前缀。 |
| `--prometheus_address`              | prometheus 地址配置，用于连接 prometheus 并获取时序数据。 | ``""``      | 选填。传入格式如 prometheus_ip:prometheus_port。 |
| `--labels`                      | 数据面自定义标签。                                  | ``""``      | 选填。 |

* node_id 默认脚本内自动生成，生成格式 `${hostName}-${PID}`，PID 为部署脚本进程（`deploy.sh`） ID。
* services_path 默认值为 ``/home/sn/config/services.yaml``，建议部署环境前确保该文件存在。
* master_info_output 默认输出在部署目录下，文件名：master.info。
* cpu_reserved 只有当前数据面需要启动 ds-worker 的时候才会进行扣减，最终可用 CPU 为 cpu_num - 1000*cpu_reserved （如果该值小于 ``0``，则忽略 cpu_reserved 参数）。
* ds_worker_unique_enable 如不配置，则一个数据面拉起一个 ds-worker；若配置，单节点数据面共用一个。
* function_proxy_unique_enable 如不配置，则一个数据面拉起一个 function-proxy；若配置，单节点数据面共用一个。
* custom_resources，预期 json 字符串。遵从 `key : str value : double` 格式，若不满足则忽略对应资源配置，重名则被覆盖。
* labels，预期 json 字符串。遵从 `key : str value : str` 格式，若不满足则忽略对应标签配置，重名则被覆盖。
* block，当前在非 master 模式时，不会落盘 ``/tmp/yr_sessions/yr_current_master_info`` 文件，暂不支持 SDK 的无参自动初始化。

## 日志配置

|  参数名  | 说明    |  默认值 |  备注 |
| ------- | ------- | ------- | ------------------- |
| `-l,--log_level` | 全局日志级别。 | ``INFO`` | 选填，取值：``DEBUG``、``INFO``、``WARN``、``ERROR``。 |
| `--log_root` | 日志根目录配置。 | ``default`` | 选填，绝对路径。 |
| `--fs_log_level` | function system 日志级别。 | ``INFO`` |选填，取值：``DEBUG``、``INFO``、``WARN``、``ERROR``。 |
| `--fs_log_path` | function system 日志子目录。 | ``""``  | 选填，配置值不能包含 ``..``。 |
| `--fs_log_compress_enable` | function system 是否开启日志压缩。 | ``true`` | 选填，取值：``true``、``false``。 |
| `--fs_log_rolling_max_size` | function system 滚动日志配置，单个日志文件大小（单位：MB）。 | ``40`` | 选填，值类型为 int。当输入值小于 ``0`` 时，部署报错；当输入值等于 ``0`` 或为异常值时，单个日志文件大小配置为 ``100`` MB。 |
| `--fs_log_rolling_max_files` | function system 滚动日志，保留最大文件数。 | ``10`` | 选填，值类型为 int。当输入值小于 ``0`` 时，部署报错；当输入值等于 ``0`` 或为异常值时，保留最大文件数配置为 ``3`` 个。 |
| `--fs_log_rolling_retention_days` | function system 滚动日志，保留最大天数。 | ``30`` | 选填，值类型为 int。当输入值小于 ``0`` 时，部署报错；当输入值等于 ``0`` 或为异常值时，保留最大天数为 ``30`` 天。 |
| `--fs_log_async_log_buf_secs` | 组件日志缓存时间（单位：s）。 | ``30`` | 选填，值类型为 int。当输入值小于 ``0`` 时，部署报错；当输入值等于 ``0`` 或为异常值时，组件日志缓存时间为 ``30`` s。 |
| `--fs_log_async_log_max_queue_size` | 组件缓存最大的日志条数，超过则落盘。 | ``51200`` | 选填，值类型为 int。当输入值小于 ``0`` 时，部署报错；当输入值等于 ``0`` 或为异常值时，组件缓存最大的日志条数为 ``51200`` 条。 |
| `--fs_log_async_log_thread_count` | function system 异步记录日志线程数。 | ``1`` | 选填，值类型为 int。当输入值小于 ``0`` 时，部署报错；当输入值等于 ``0`` 或为异常值时，异步记录日志线程数为 ``1``。 |
| `--fs_also_log_to_stderr` | 组件日志是否同时输出到标准错误输出。 | ``false`` |选填。 |
| `--runtime_log_level` | runtime 日志级别。 | ``INFO`` |选填。  |
| `--runtime_log_rolling_max_size` | runtime 日志压缩阈值（单位：MB）。 | ``40`` |选填。  |
| `--runtime_log_rolling_max_files` | runtime 日志压缩保留文件数。 | ``20`` |选填。  |
| `--runtime_log_path` | runtime 日志子目录。 | ``""`` | 选填，配置值不能包含 ``..`` |
| `--ds_log_level` | 数据系统日志级别。 | ``INFO`` |选填，取值：``DEBUG``、``INFO``、``WARN``、``ERROR``。  |
| `--ds_log_path` | 数据系统日志子目录。 | ``""``  | 选填，配置值不能包含 ``..``。 |
| `--ds_log_rolling_max_size` | 数据系统组件单个日志文件最大大小（单位：MB）。 | ``40`` | 选填。 |
| `--ds_log_rolling_max_files` | 数据系统组件压缩日志文件保留最大文件数量 | ``32`` |选填。  |
| `--ds_log_monitor_enable` | 数据系统是否开启资源日志监控。 | ``false`` |选填。  |
| `--etcd_log_path` | etcd 日志子目录。 | ``""``  |选填，配置值不能包含 ``..``。  |
| `--user_log_export_mode` | runtime 标准输出日志导出方式。 | ``file`` | 选填，runtime 标准输出日志导出方式。取值：``file``（导出至文件）、``std``（导出至标准输出流），默认导出至文件。 |
| `--enable_separated_redirect_runtime_std` | 开启 runtime 日志分进程落盘。 | ``true`` | 选填，为 ``true`` 时，在 runtime 标准输出日志导出方式为文件时生效。不同的 runtime 进程的 std 输出会重定向到不同的文件，输出文件名为 ``{runtimeID}.out {runtimeID}.err``。此模式不支持 runtime 日志的压缩和轮转等功能。 |
| `--runtime_std_rolling_enable` | 开启 runtime 标准输出日志的滚动压缩。 | ``false`` | 选填。仅在 runtime 日志分进程落盘时生效，会同时开始标准输出的日志的滚动和压缩。 |

* `${deploy_path}` 为部署脚本指定的部署路径。
* `${NODE_ID}` 为数据面组件的节点标识，可以通过传参指定（`yr_agent.sh -n xx`），如未指定，则脚本内生成。
* log_root 配置为 ``default`` 时，对于 yr_master 启动默认日志目录为 ``${deploy_path}/log``，yr_agent 启动默认日志目录 ``${deploy_path}/${NODE_ID}/log``。
* 配置指定的日志根目录/子目录，对于已存在的目录，需要保证当前用户有读写权限，建议权限设置不大于 ``750`` 。目录不存在情况下，新建目录。
* `--runtime_std_rolling_enable` 依赖开源软件 logrotate。如需此功能，需用户预先安装 logrotate 工具。默认开启压缩。会使用 `runtime_log_rolling_max_size`， `runtime_log_rolling_max_files` 配置 logrotate 使用的轮转压缩参数，每 ``5``s 调用一次。注意：由于 logrotate 限制，压缩文件大小并不能严格保证为 `runtime_log_rolling_max_size` 大小，会根据用户日志输出情况存在波动。另外，由于 logrotate 本身限制，存在日志丢失可能，该选项与 log-to-driver 功能混用可能导致部分日志无法正常回传到 driver 侧。

## 端口配置

|  参数名  | 说明    |  默认值 |  备注 |
| ------- | ------- | ------- | ------------------- |
| `--port_policy` | 端口分配策略。 | ``RANDOM`` | 选填。``FIX``：固定，``RANDOM``：从可用端口范围随机选取。 |
| `--control_plane_port_min` | 控制面分配端口范围最小值。 | ``10000`` | 选填。当 port_policy 配置为 ``RANDOM`` 生效，自动分配组件端口。 |
| `--control_plane_port_max` | 控制面分配端口范围最大值。 | ``20000`` | 选填。当 port_policy 配置为 ``RANDOM`` 生效，自动分配组件端口。 |
| `--data_plane_port_min` | 数据面分配端口范围最小值。 | ``20000`` | 选填。当 port_policy 配置为 ``RANDOM`` 生效，自动分配组件端口。 |
| `--data_plane_port_max` | 数据面分配端口范围最大值。 | ``40000`` | 选填。当 port_policy 配置为 ``RANDOM`` 生效，自动分配组件端口。 |
| `--function_agent_port` | function-agent 监听端口。 | ``58866`` | 选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--function_proxy_port` | function-proxy 监听端口。 | ``22772`` | 选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--function_proxy_grpc_port` | function-proxy 监听端口，用于 driver 注册。 | ``22773`` | 选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--global_scheduler_port` | global scheduler 监听端口。 | ``22770`` | 选填。对于启动控制面组件，当 port_policy 配置为 ``FIX`` 生效；对于启动数据面组件，master_info 未配置时生效。 |
| `--ds_master_port` | ds-master 监听端口。 | ``12123`` |选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--ds_worker_port` | ds-worker 监听端口。 | ``31501`` |选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--dashboard_port`           | dashboard 监听端口。                   | ``9080``   |选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--dashboard_grpc_port`      | dashboard grpc 监听端口。              | ``9081``   |选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--faas_frontend_http_port`  | frontend http 监听端口。               | ``8888``   |选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--faas_frontend_grpc_port`  | frontend grpc 监听端口。              | ``31223``  |选填。当 port_policy 配置为 ``FIX`` 生效。 |
| `--disable_nc_check` | 是否禁用 nc 工具进行端口探测，防止组件端口冲突。 | 无参 |选填。默认不配置此参数。 |

:::{Note}

如果手动配置端口，需要自行保证端口不会冲突，并且端口在有效范围内。

:::

## function system 配置

|  参数名  | 说明    | 默认值                                                  |  备注 |
| ------- | ------- |------------------------------------------------------| ------------------- |
| `-g,--driver_gateway_enable` | function-proxy 开启端口，支持 driver 直连。 | ``false``                                                | 选填，取值：``true``、``false``，yr_master.sh 启动固定为 ``true``。 |
| `--state_storage_type` | 状态存储类型。 | ``datasystem``                                              | 选填，取值：``datasystem``、``disable``。 |
| `--min_instance_cpu_size` | 函数实例支持请求最小 cpu（单位：``1/1000`` 核）。 | ``300``                                                  |选填。 |
| `--min_instance_memory_size` | 函数实例支持请求最小 memory（单位：MB）。 | ``128``                                                  | 选填。|
| `--max_instance_cpu_size` | 函数实例支持请求最大 cpu（单位：``1/1000`` 核）。 | ``16000``                                                |选填。最大值：``1073741824``。 |
| `--max_instance_memory_size` | 函数实例支持请求最大 memory（单位：MB）。 | ``128``                                           | 选填。最大值：``1073741824``。|
| `--metrics_collector_type` | metric 收集类型。 | ``proc``                                                 |选填。取值：``proc``、``node``。 |
| `--merge_process_enable` | 是否开启进程合并部署。 | ``true``                                                 | 选填。runtime-manager 与 function-agent 组件进程合并部署，取值：``true``、``false``。 |
| `--is_schedule_tolerate_abnormal` | 调度是否容忍异常。 | ``false``                                                |选填。  |
| `--system_timeout` | 系统超时时间（单位:ms）。 | ``1800000``                                              |选填。大于 ``0``，设置值小于 ``12000`` 时，重置为 ``12000``。 |
| `--npu_collection_mode` | 采集 npu 信息模式。  | ``all``                                                  |选填。取值：``all``、 ``count``、 ``hbm``、``sfmd``、``topo``、``off``。  |
| `--gpu_collection_enable` | 是否采集 gpu 信息。 | ``false``                                                |选填。 |
| `--function_agent_litebus_thread` | function-agent litebus 线程数。 | ``20``                                                   | 选填。|
| `--function_proxy_litebus_thread` | function-proxy litebus 线程数。 | ``20``                                                   | 选填。|
| `--function_master_litebus_thread` | function-master litebus 线程数。                      | ``20``                                                   |选填。 |
| `--function_agent_alias` | function agent 别名。 | ``""``                                                   | 选填。|
| `--enable_print_perf` | 是否打开 function proxy 的 perf 打印。 | ``false``                                                | 选填。|
| `--enable_meta_store` | 是否打开 meta store。 | ``false``                                                | 选填。|
| `--enable_dashboard`                 | 是否打开 dashboard。 | ``false``                                                | 选填。|
| `--enable_collector`                 | 是否打开 collector。 | ``false``                                                | 选填。|
| `--enable_faas_frontend`             | 是否打开 frontend。 | ``false``                                      | 选填。|
| `--enable_jemalloc` | 是否打开 function master 和 function proxy 的 jemalloc。 | ``true``                                                 | 选填。|
| `--massif_enable` | runtime 是否启动 massif。  | ``false``                                                | 选填。|
| `--is_partial_watch_instances` | function-proxy 是否开启部分 watch。 | ``false``                                                | 选填。|
| `--enable_persistence` | 是否打开 meta store 持久化到 etcd 以支持 meta store 的数据可靠性。打开后支持 function master 重启容灾。 | ``false``                                                | 选填。|
| `--meta_store_mode` | meta store 运行模式。local 模式：数据保存在本地；etcd 模式：数据保存在 etcd 中。 | ``local``                                                | 选填，取值：``local``、``etcd``。                                                                                                         |
| `--schedule_relaxed`    | 大于 ``0`` 表示启用宽松调度策略。当调度选中 scheduleRelaxed 个节点或者 pod 后即退出流程，不遍历所有节点或者 pod 。配置为 ``-1`` 或 ``0``  表示需要遍历所有节点或 pod ，未配置该参数时默认为 ``-1``。 warn ：当 > ``0`` 时会降低调度的准确性。 | ``1``                                                    | 选填。 |
| `--max_priority`    | 为 ``0`` 时关闭实例优先级调度。  当 > ``0`` 时使能实例优先级调度，maxPriority 表示集群支持的实例最大优先级，最大值为 ``65535``。   | ``0``                                                    | 选填。 |
| `--enable_preemption`    | 使能调度抢占功能（高优先级实例可以抢占低优先级被配置为可抢占的实例），仅 maxPriority > 0 生效。   | ``false``                                                | 选填。 |
| `--local_schedule_plugins` | local scheduler 调度插件启用列表。 | 默认开启所有调度插件。                                           | 选填，支持 ``Default``，``Label``，``ResourceSelector``，``Heterogeneous``，自定义插件时，必选 ``Default``，其他三种插件可选；启用 Heterogeneous 插件需要 npu_collection_mode 值不为 ``off``。|
| `--domain_schedule_plugins` | domain scheduler 调度插件启用列表。 | 默认开启所有调度插件。                                           | 选填，支持 ``Default``，``Label``，``ResourceSelector``，``Heterogeneous``，自定义插件时，必选 ``Default``，其他三种插件可选；启用 Heterogeneous 插件需要 npu_collection_mode 值不为 ``off``。 |
| `--meta_store_max_flush_concurrency` | meta store 持久化到 etcd 允许最大请求并发数，当需要持久化的请求数超过该值后，会进行限流，延迟持久化。  | ``100``                                                  | 选填，大于 ``0``，建议以实际 etcd 能处理的能力为准。 |
| `--meta_store_max_flush_batch_size` | meta store 持久化到 etcd 一次性最多允许的条数，采用分批次持久化。  | ``50``                                                   | 选填，大于 ``0``，建议以实际 etcd 能处理的能力为准。 |
| `--fs_health_check_timeout` | 函数系统组件健康检查超时时间（单位：秒）。  | ``1``   | 选填，大于 ``0``。 |
| `--fs_health_check_retry_interval` | 函数系统组件健康检查重试间隔（单位：秒）。  | ``0``   | 选填，大于等于 ``0``。 |
| `--fs_health_check_retry_times` | 函数系统组件健康检查重试次数。 | ``60``   | 选填，大于 ``0``。 |

* npu_collection_mode：count 模式只采集 NPU 卡数量信息；hbm 模式采集 NPU 卡数量以及 HBM 容量等信息；sfmd 模式采集 NPU 卡数量、HBM 容量以及 NPU 卡 IP 信息；topo 模式采集 NPU 卡数量、HBM 容量以及 NPU 卡 topo 信息；all 模式采集以上所有 NPU 信息；off 模式不采集任何 NPU 信息。
* local_schedule_plugins 和 domain_schedule_plugins：用户自定义插件列表时，建议两个参数使用一致的调度插件列表，例如：--local_schedule_plugins='["Default", "Heterogeneous"]' --domain_schedule_plugins='["Default", "Heterogeneous"]'。

## runtime 配置

| 参数名                                   | 说明                                          | 默认值    |  备注 |
|---------------------------------------|---------------------------------------------|--------| ------------------- |
| `--runtime_init_port`                   | runtime 初始端口。                                | ``21006``  | 选填，当 port_policy 配置为 ``FIX`` 生效。 |
| `--runtime_heartbeat_enable`            | 是否开启 function-proxy 与 runtime 心跳。            | ``true``   | 选填，取值：``true``、``false``。 |
| `--runtime_heartbeat_timeout_ms`        | 心跳超时时间（单位：ms）。                               | ``100000`` |选填。 |
| `--runtime_max_heartbeat_timeout_times` | 心跳超时最大次数。                                    | ``18``     |选填。 |
| `--runtime_port_num`                    | runtime 端口数量。                                | ``65535``  |选填。|
| `--runtime_recover_enable`              | 是否开启 recover。                                | ``false``  | 选填，取值：``true``、``false``。 |
| `--runtime_init_call_timeout_seconds`   | init call 超时时间（单位：s）。                        | ``600``    | 选填。 |
| `--runtime_conn_timeout_s`              | 首次连接 runtime 超时时间（单位：s）。                     | ``30``     | 选填。 |
| `--is_protomsg_to_runtime`              | 传输到 runtime 数据格式是否为 proto，LibRuntime 版本需要开启。 | ``false``  |  选填，取值：``true``、``false``。   |
| `--enable_inherit_env`                  | 使能通过 runtime-manager 继承 env 来创建 runtime 进程。  | ``false`` | 选填。 |
| `--runtime_direct_connection_enable`    | runtime 直连，会从 runtime 端口数量里为 runtime 直连分配 server 端口。  | ``false`` | 选填。 |
| `--memory_detection_interval`    |  runtime 内存检测时间间隔，单位：毫秒（ms）。 | ``1000`` | 选填，最低支持 100ms 检测，检测越频繁会导致 CPU 底噪过高。用户需知该特性 10ms 间隔的检测频率约引入 60~100m CPU 的底噪，100ms 间隔的检测频率约引入 30m CPU 的底噪，挤占 runtime 能使用的资源。进程部署，没有资源隔离的情况下，共用宿主机的资源限制，该 CPU 底噪影响可能并不明显。
| `--oom_kill_enable`    |  是否开启基于 runtime 进程内存使用情况检测超过配置内存大小并 kill。 | ``false`` | 选填。 |
| `--oom_kill_control_limit`    |  基于进程内存使用配置运行时 OOM kill 的控制限（单位：MB）。如 limit=-5，设置的内存大小 -5 MB（例如 300m-128Mi 的函数，内存上限为 123MB），主动预防 OOM | ``0`` | 选填，对于进程部署，当系统负载不高时，可以适当设置控制限为正数，例如 300m-128Mi 的函数，limit=10，设置的内存大小 +10 MB，单个函数内存上限为 138MB，允许函数超分使用内存，提高函数性能和系统资源利用率。 |
| `--oom_consecutive_detection_count`    |  内存连续检测异常点的次数。参考控制图中的 3σ 原则。 | ``3`` | 选填。 |
| `--runtime_default_config`    | Java runtime 自定义 JVM 参数。 | ``""`` | 选填。 |
| `--kill_process_timeout_seconds`    | 支持用户配置优雅退出（SIGINT->SIGKILL）的超时时间（单位：s），默认立即删除，资源回收。 | ``0`` | 选填。 |
| `--runtime_home_dir`    | 支持用户配置 runtime 函数的 HOME 环境变量。 | 当为空时，使用 function_agent/runtime_manager 组件部署时的操作系统 HOME 环境变量。 | 若用户函数所属用户与 function_agent/runtime_manager 组件部署时的操作系统 HOME 环境变量仍有区别，需要指定该参数方便用户函数内使用 HOME 环境变量。 |
| `--runtime_instance_debug_enable`       | 是否开启 runtime 实例 Debug 模式。                  | ``false``  | 选填，取值：``true``、``false``。 |

* 对于 Python runtime 和 posix-custom-runtime 函数，默认设置环境变量 `PYTHONUNBUFFERED = 1`，便于用户函数日志实时打印。

* enable_inherit_env：使能情况下，LD_LIBRARY_PATH、PYTHONPATH 及 PATH 将以追加的方式被 runtime 继承，POSIX_LISTEN_ADDR、POD_IP、YR_RUNTIME_ID、INSTANCE_ID、DATA_SYSTEM_ADDR、DRIVER_SERVER_PORT、HOME、HOST_IP、FUNCTION_LIB_PATH、LAYER_LIB_PATH、CLUSTER_ID、NODE_ID 等以及其他与用户自定义同名的环境变量则不会被继承，请妥善配置环境变量避免冲突。

* runtime_direct_connection_enable: 使能情况下，1 个 runtime 占用 1 个端口数量。若开启 runtime 的 server 翻转功能， 复用 server 翻转分配的端口。 runtime-manager 会给 runtime 传递 2 个环境变量，`RUNTIME_DIRECT_CONNECTION_ENABLE` 表示 runtime 是否开启 runtime 直连，`DERICT_RUNTIME_SERVER_PORT` 表示该 runtime 直连 server 使用的端口。

* runtime_default_config 配置参考：`--runtime_default_config="{\"java1.8\": ["-XX:InitialRAMPercentage=35.0", "-XX:+UseConcMarkSweepGC", "-XX:+CMSClassUnloadingEnabled", "-XX:+CMSIncrementalMode", "-XX:+CMSScavengeBeforeRemark", "-XX:+UseCMSInitiatingOccupancyOnly", "-XX:CMSInitiatingOccupancyFraction=70", "-XX:CMSFullGCsBeforeCompaction=5", "-XX:MaxGCPauseMillis=200", "-XX:+ExplicitGCInvokesConcurrent", "-XX:+ExplicitGCInvokesConcurrentAndUnloadsClasses"]",
  \"java11\": ["-XX:+UseG1GC", "-XX:MaxRAMPercentage=80.0", "-XX:+TieredCompilation"],
  \"java17\": ["-XX:+UseZGC", "-XX:+AlwaysPreTouch", "-XX:+UseCountedLoopSafepoints", "-XX:+TieredCompilation", "--add-opens=java.base/java.util=ALL-UNNAMED", "--add-opens=java.base/java.lang=ALL-UNNAMED", "--add-opens=java.base/java.net=ALL-UNNAMED", "--add-opens=java.base/java.io=ALL-UNNAMED", "--add-opens=java.base/java.math=ALL-UNNAMED", "--add-opens=java.base/java.time=ALL-UNNAMED", "--add-opens=java.base/java.text=ALL-UNNAMED", "--enable-preview"],
  \"java21\": ["-XX:+UseZGC", "-XX:+ZGenerational", "-XX:+AlwaysPreTouch", "-XX:+UseCountedLoopSafepoints", "-XX:+TieredCompilation", "--add-opens=java.base/java.util=ALL-UNNAMED", "--add-opens=java.base/java.lang=ALL-UNNAMED", "--add-opens=java.base/java.net=ALL-UNNAMED", "--add-opens=java.base/java.io=ALL-UNNAMED", "--add-opens=java.base/java.math=ALL-UNNAMED", "--add-opens=java.base/java.time=ALL-UNNAMED", "--add-opens=java.base/java.text=ALL-UNNAMED", "--enable-preview"]}"`。

## data system 配置

|  参数名  | 说明    |  默认值 |  备注 |
| ------- | ------- | ------- | ------------------- |
| `--ds_spill_enable` | 是否内存不足的情况下写磁盘。 | ``false`` | 选填，取值：``true``、``false``。 |
| `--ds_spill_directory` | 写入磁盘路径。 | | 选填，spill_enable 为 ``true`` 时生效，默认为 ``install_path/spill``。|
| `--ds_spill_size_limit` | 写入磁盘大小限制（单位：MB）。 | ``20480`` |选填。 |
| `--ds_rpc_thread_num` | rpc 线程数量。 | ``32`` |选填。 |
| `--ds_node_dead_timeout_s` | ds-worker 与 ds-master 之间心跳的超时时间（秒），当心跳超时时 ds-master 会将 ds-worker 的资源清理掉。 | ``86400`` |选填。 |
| `--ds_client_dead_timeout_s` | ds-client 跟 ds-worker 之间心跳的超时时间（秒），当心跳超时时 ds-worker 会将 ds-client 的资源清理掉。 | ``60`` | 选填。|
| `--ds_node_timeout_s` | 数据系统节点超时时间（单位：s）。 | ``1800`` |选填。 |
| `--ds_heartbeat_interval_ms` | 数据系统心跳超时时间（单位：ms）。 | ``300000`` |选填。 |
| `--ds_max_client_num` | 单个数据系统 worker 最大 client 数。 | ``1000`` |选填。 |
| `--ds_memory_reclamation_time_second` | 数据系统 worker 内存恢复时间（单位：s）。 | ``5`` |选填。 |
| `--enable_distributed_master` | 数据系统 worker 是否使能分布式 master。 | ``true`` |选填，取值：``true``、``false``。 |
| `--ds_l2_cache_type` | 数据系统 l2 cache 后端类型。 | ``none`` | 选填，取值：``none``、``sfs``。|
| `--ds_sfs_path` | l2 cache 选择为 sfs 时，指定 sfs 路径。 | ``""`` | 选填。 |
| `--enable_lossless_data_exit_mode` | 数据系统是否开启无损退出模式。开启后 ds-worker 退出前会迁移数据到其它的 ds-worker 上。如果没有可用的 ds-worker 节点，直接关闭，此时数据仍然会丢失。 | ``false`` |选填。 |
| `--zmq_chunk_sz` | 数据面数据系统使用 ZMQ 通信时的消息分片大小（单位：byte）。 | ``1048576`` | 选填。 |
| `--ds_enable_huge_tlb` | 是否开启共享内存大页内存功能，它可以提高内存访问，减少页表的开销。 | ``false`` | 选填，取值：``true``、``false``。|
| `--ds_enable_fallocate` | 由于 Kubernetes(k8s) 的资源计算策略，共享内存有时会被计算两次，这可能会导致客户端 OOM 崩溃。为了解决这个问题，我们使用了 fallocate 来链接客户端和工作节点的共享内存，从而纠正内存计算错误。缺省情况下，fallocate 是使能的。启用 fallocate 会降低内存分配的效率。 | ``true`` | 选填，取值：``true``、``false``。 |
| `--ds_enable_thp` | 是否启用透明大页（ransparent Huge Page,THP）功能，启用透明大页可以提高性能，减少页表开销，但也可能导致 Pod 内存使用增加。 | ``false`` | 选填，取值：``true``、``false``。 |
| `--ds_arena_per_tenant` | 每个租户的共享内存分配器数量。多分配器可以提高第一次分配共享内存的性能，但每个分配器会多使用一个 fd，导致 fd 资源使用量上升。| ``16`` | 选填，取值范围：``[1, 32]``。 |

* `${install_path}` 为数据面的安装路径。

## etcd 配置

|  参数名  | 说明    |  默认值 |  备注 |
| ------- | ------- | ------- | ------------------ |
| `--etcd_ip` | etcd ip。 | ``""`` | 选填，部署控制面组件，默认为本地 IP。 |
| `--etcd_port` | etcd 端口。 | ``32379`` |选填。 |
| `--etcd_peer_port` | etcd 集群交互端口。 | ``32380`` |选填。|
| `--etcd_auth_type` | etcd 鉴权类型。 | ``Noauth`` | 选填，取值：``Noauth``、``TLS``。 |
| `--etcd_cert_file` | 经过 CA 签名的证书位置。 | ``""`` | 选填，etcd_auth_type 配置为 ``TLS`` 生效。 |
| `--etcd_key_file` | CA 机构私钥位置。 | ``""`` | 选填，etcd_auth_type 配置为 ``TLS`` 生效。 |
| `--etcd_ca_file` | 鉴权根证书位置。 | ``""`` | 选填，etcd_auth_type 配置为 ``TLS`` 生效。 |
| `--etcd_client_cert_file` | etcd 客户端的签名证书位置。 | ``""`` | 选填，etcd_auth_type 配置为 ``TLS`` 生效。 |
| `--etcd_client_key_file` | etcd 客户端的密钥位置。 | ``""`` | 选填，etcd_auth_type 配置为 ``TLS`` 生效。 |
| `--etcd_proxy_enable` | 是否开启 etcd proxy。 | ``false`` | 选填，取值：``true``、``false``，配置为 ``true``，且 etcd_proxy_nums 大于 ``0``，才会开启 proxy。 |
| `--etcd_proxy_nums` | proxy 数量。 | ``0`` | 选填。 |
| `--etcd_proxy_port` | proxy 端口。 | ``23790`` | 选填。|
| `--etcd_no_fsync` | 是否禁止 fsync。 | ``true`` | 选填。|
| `--etcd_compact_retention` | etcd 压缩的 revision 数量阈值。 | ``100000`` | 选填。|
| `--etcd_target_name_override`           | etcd TLS 证书域名。               | ``""`` | 用于 etcd TLS 证书校验，需要与实际的证书保存一致。                                                |

(security-params)=

## 安全配置

启动脚本配置参数如下：

|  参数名  | 说明    |  默认值 |  备注 |
| ------- | ------- | ------- | ------------------- |
| `--etcd_auth_type` | etcd 鉴权类型。 |``Noauth`` | 建议配置为 ``TLS``。 |
| `--etcd_ssl_base_path`| etcd 证书目录。 | ``""`` | 配置绝对路径，根据实际路径配置。 |
| `--etcd_ca_file` | etcd CA 证书文件名。 | ``ca.crt`` | 可保持默认配置。 |
| `--etcd_cert_file` | etcd 服务端证书文件名。 | ``server.crt`` | 可保持默认配置。 |
| `--etcd_key_file` |  etcd 服务端私钥文件名。 | ``server.key`` | 可保持默认配置。 |
| `--etcd_client_cert_file` | etcd 客户端证书文件名。 | ``client.crt`` | 可保持默认配置。 |
| `--etcd_client_key_file` |  etcd 客户端私钥文件名。 | ``client.key`` | 可保持默认配置。 |
| `--ssl_enable` | 函数系统组件是否开启 TLS。 | ``false`` | 指定配置为 ``true``，取值：``true``、``false``。 |
| `--ssl_base_path` | 函数系统组件证书目录。 | ``""``  | 配置绝对路径，根据实际路径配置。 |
| `--ssl_root_file` | 函数系统组件 CA 证书文件名。 | ``ca.crt``  | 可保持默认配置。 |
| `--ssl_cert_file` | 函数系统组件证书文件名。 | ``module.crt`` | 可保持默认配置。 |
| `--ssl_key_file` | 函数系统组件证书私钥文件名。 | ``module.key`` | 可保持默认配置。 |
| `--ds_component_auth_enable` | 数据系统组件是否开启鉴权。 | ``false`` | 指定配置为 ``true``，取值：``true``、``false``。 |
| `--curve_key_path`  | 数据系统 curve 证书目录。 | ``""`` | 配置绝对路径，根据实际路径配置。 |
| `--runtime_ds_auth_enable`  | 数据系统与 runtime 是否开启认证。 | ``false``  | 指定配置为 ``true``，取值：``true``、``false``。 |
| `--runtime_ds_encrypt_enable`  | 数据系统与 runtime 是否开启加密。 | ``false``  | 指定配置为 ``true``，取值：``true``、``false``。 |
| `--runtime_ds_connect_timeout`  | runtime 检测数据系统异常的超时时间，单位：秒（s）。 | ``1800``  | 可保持默认配置。 |
| `--cache_storage_auth_enable` | 是否开启 function-proxy 与数据系统间的连接认证。 | ``false`` | 指定配置为 ``true``，取值：``true``、``false``。 |
| `--cache_storage_auth_type` | function-proxy 与数据系统间的认证类型。 | ``Noauth`` | 可选的配置类型：``Noauth``、``ZMQ``、``AK/SK``，建议配置为 ``ZMQ``。 |
| `--dashboard_ssl_enable` | dashboard 是否开启单向 TLS 认证。 | ``false`` | 指定配置为 ``true``，取值：``true``、``false``。 |
| `--dashboard_ssl_base_path` | dashboard 证书目录。 | ``""``  | 配置绝对路径，根据实际路径配置。 |
| `--dashboard_ssl_cert_file` | dashboard 服务端证书文件名。 | ``server.crt`` | 可保持默认配置。 |
| `--dashboard_ssl_key_file` | dashboard 服务端证书私钥文件名。 | ``server.key`` | 可保持默认配置。 |
| `--prometheus_ssl_enable` | prometheus 是否开启 TLS。 | ``false`` | 指定配置为 ``true``，取值：``true``、``false``。 |
| `--prometheus_ssl_base_path` | prometheus 证书目录。 | ``""``  | 配置绝对路径，根据实际路径配置。 |
| `--prometheus_ssl_root_file` | prometheus CA 证书文件名。 | ``ca.crt``  | 可保持默认配置。 |
| `--prometheus_ssl_cert_file` | prometheus 客户端证书文件名。 | ``client.crt`` | 可保持默认配置。 |
| `--prometheus_ssl_key_file` | prometheus 客户端证书私钥文件名。 | ``client.key`` | 可保持默认配置。 |
