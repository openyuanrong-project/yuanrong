Config 结构体
==============

.. cpp:struct:: Config

    Init 接口入参，用于初始化 openYuanrong 系统。

    .. note::
        当同时设置了配置字段和环境变量时，配置参数优先于环境变量。

    **公共成员**
 
    .. cpp:member:: Mode mode = Mode::CLUSTER_MODE

       部署模式。

       必需。支持的值：``LOCAL_MODE``：本地模式（单机多线程），``CLUSTER_MODE``：集群模式（多机多进程）。

    .. cpp:member:: std::vector<std::string> loadPaths

       指定 `so` 路径。
       
       默认不填则由 `services.yaml` 指定。

    .. cpp:member:: std::string functionUrn = ""

       部署 C++ 函数后返回的函数 ID。

       在 `CLUSTER_MODE` 中必需。对应的环境变量为 `YRFUNCID`。

    .. cpp:member:: std::string pythonFunctionUrn = ""

       部署 Python 函数后返回的函数 ID。

       在 `CLUSTER_MODE` 中可选。对应的环境变量为 `YR_PYTHON_FUNCID`。

    .. cpp:member:: std::string javaFunctionUrn = ""

       部署 Java 函数后返回的函数 ID。

       在 `CLUSTER_MODE` 中可选。对应的环境变量为 `YR_JAVA_FUNCID`。

    .. cpp:member:: std::string serverAddr = ""

       openYuanrong集群地址。

       在 `CLUSTER_MODE` 中必需。对应的环境变量为 `YR_SERVER_ADDRESS`。

    .. cpp:member:: std::string dataSystemAddr = ""

       集群内数据系统服务器地址。

       在 `CLUSTER_MODE` 中必需。对应的环境变量为 `YR_DS_ADDRESS`。

    .. cpp:member:: bool enableServerMode = false

       ``true``：函数代理作为服务器，``false``：运行时作为服务器运行。 默认 ``false``。

    .. cpp:member:: int threadPoolSize = 0

       线程池大小。

       在 `LOCAL_MODE` 中必需。有效范围：``1~64``。如果超出范围，则默认为 CPU 核心数。仅在 `ParallelFor` 中使用。

    .. cpp:member:: uint32_t localThreadPoolSize = 10

       本地线程池大小。

       在 `LOCAL_MODE` 中必需。有效范围：``1-64``。如果超出范围，则默认为 CPU 核心数。默认 ``10``。

    .. cpp:member:: int recycleTime = DEFAULT_RECYCLETIME

       实例最大空闲时间。

       如果空闲时间超过此持续时间，实例将被终止。单位：秒。有效范围：``1~3000``。如果未配置，则默认 ``2``。

    .. cpp:member:: bool enableMTLS = false

       是否开启云外客户端双向认证。默认 ``false``。

    .. cpp:member:: std::string certificateFilePath = ""

       客户端证书文件路径。

    .. cpp:member:: std::string privateKeyPath = ""

       客户端私钥文件路径。

    .. cpp:member:: std::string verifyFilePath = ""

       服务器证书文件路径。

    .. cpp:member:: std::string serverName = ""

       TLS 的服务器名称。

    .. cpp:member:: bool enableDsAuth = false

       是否开启数据系统鉴权，默认 ``false``，为关闭。

    .. cpp:member:: bool enableDsEncrypt = false

       ``true`` 表示数据系统加解密开启，需配置 `dsPublicKeyContext`、`runtimePublicKeyContext` 和 `runtimePrivateKeyContext`；``false`` 表示关闭。

    .. cpp:member:: std::string dsPublicKeyContextPath = ""

       数据系统 TLS 认证的 worker 进程公钥路径。
    
       如果 `enableDsEncrypt` 为 ``true`` 时，该参数不能为空，否则会抛出异常。

    .. cpp:member:: std::string runtimePublicKeyContextPath = ""

       数据系统 TLS 认证的客户端公钥路径。
    
       如果 `enableDsEncrypt` 为 ``true`` 时，该参数不能为空，否则会抛出异常。

    .. cpp:member:: std::string runtimePrivateKeyContextPath = ""

       数据系统 TLS 认证的客户端私钥路径。
       
       如果 `enableDsEncrypt` 为 ``true`` 时，该参数不能为空，否则会抛出异常。

    .. cpp:member:: int maxTaskInstanceNum = -1

       限制无状态函数创建的最大实例数量。

       有效范围：``1~65536``。如果未配置，则默认 ``-1``；如果值无效，`Init` 接口将抛出异常。

    .. cpp:member:: std::string metricsLogPath = ""

       自定义 metrics 日志路径。

       对应的环境变量为 ``YR_METRICS_LOG_PATH``。
    
    .. cpp:member:: bool enableMetrics = false

       是否启用 metrics 收集。

       ``false`` 表示禁用，``true`` 表示启用。仅在集群中调用时生效。默认 ``false``。对应的环境变量为 `YR_ENABLE_METRICS`。

    .. cpp:member:: int maxConcurrencyCreateNum = 100

       设置无状态实例的最大并发创建数量。

       必须大于 ``0``。默认 ``100``。

    .. cpp:member:: uint32_t maxLogSizeMb = 0

       可选。客户端日志的单个文件最大体积，单位：MB。
    
       默认 ``0`` （若为 ``0`` 则最终会被置为 ``40``）。超出大小限制后滚动切分，每30秒检查一次。负数会抛异常。对应的环境变量为 `YR_MAX_LOG_SIZE_MB`。

    .. cpp:member:: uint32_t maxLogFileNum = 0

       可选。客户端日志滚动后最多保留的文件数。
       
       默认 ``0`` （若为 ``0`` 则最终会被置为 ``20``）。超出数量限制后从最旧的日志开始删除，30秒检查一次。对应的环境变量为 `YR_MAX_LOG_FILE_NUM`。

    .. cpp:member:: bool logCompress = true

       可选。设置是否对滚动切分的日志文件进行压缩。

       默认 ``true``。对应的环境变量为 `YR_LOG_COMPRESS`。

    .. cpp:member:: std::string logLevel = ""

       可选。日志级别：``DEBUG``、``INFO``、``WARN``、``ERROR``。

       无效值将默认 ``INFO``。对应的环境变量为 `YR_LOG_LEVEL`。

    .. cpp:member:: std::int32_t rpcTimeout = 30 * 60

       客户端 RPC 超时时间（秒）。

       需大于 ``10``。默认 ``1800``。

    .. cpp:member:: std::string logDir = ""

       可选。客户端日志目录。

       默认是当前目录。若目录不存在则创建目录并在该目录下生成日志文件。 

    .. cpp:member:: std::string logPath = ""

       （已弃用，使用 `logDir`）备用日志目录。

    .. cpp:member:: std::string workingdir = ""

       openYuanrong函数目录的绝对路径（`service.yaml` 所在位置）。

       默认为空。

    .. cpp:member:: std::string ns = ""

       客户端函数的默认命名空间。

    .. cpp:member:: std::unordered_map<std::string, std::string> customEnvs
        
       运行时的自定义环境变量（目前仅支持 `LD_LIBRARY_PATH`）。

    .. cpp:member:: bool isLowReliabilityTask = false

       为无状态实例启用低可靠性模式（在大规模场景中提高创建性能）。

    .. cpp:member:: bool attach = false
        
       在初始化期间将 libruntime 实例附加到现有实例（仅支持 KV API）。

       默认 ``false``。

    .. cpp:member:: std::string tenantId = ""

       租戶 ID。

.. cpp:struct:: ClientInfo

    由 `Init` 返回的对象，包含运行时上下文的元数据。

    **公共成员**
 
    .. cpp:member:: std::string jobId

       当前上下文的作业 ID。

       每次 `Init` 都会生成一个随机的作业 ID。

    .. cpp:member:: std::string version

       当前openYuanrong SDK 的版本号。

    .. cpp:member:: std::string serverVersion

       成功初始化连接后获取的服务器版本号。
