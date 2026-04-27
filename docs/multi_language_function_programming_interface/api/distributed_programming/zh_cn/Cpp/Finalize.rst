Finalize
===========

.. cpp:function:: void YR::Finalize()

    清理 openYuanrong 系统。

    此函数负责释放程序执行期间创建的资源，例如函数实例和数据对象。它确保不会出现资源泄漏，从而避免在生产环境中出现问题。

    .. code-block:: cpp

       YR::Config conf;
       conf.mode = YR::Config::Mode::CLUSTER_MODE;
       YR::Init(conf);
       YR::Finalize();

    .. note::
        - 在集群部署场景中，如果工作进程退出并重新启动，可能会导致进程残留。在这种情况下，建议重新部署集群。像 Donau 或 SGE 这样的部署场景可以依赖资源调度平台的能力来回收进程。
        - 此函数应在调用 `Init` 函数完成系统初始化之后调用，否则会导致异常。

    抛出：
        :cpp:class:`Exception` - 如果在系统初始化之前调用了 `Finalize` ，会抛出“Please init YR first”的异常。