.. _custom_extensions:

yr.InvokeOptions.custom_extensions
----------------------------------------------

.. py:attribute:: FunctionGroupOptions.custom_extensions
   :type: Dict[str, str]

   指定用户自定义配置，如函数并发度。同时也可以作为 metrics 的用户自定义 tag，用于采集用户信息。

   .. list-table:: 常见的 custom_extensions 配置

      * - "Concurrency"
        - 并发度。范围：[1,1000]。
      * - "lifecycle"
        - detached，支持 detached 模式。
      * - "DELEGATE_DIRECTORY_INFO"
        - 自定义目录支持创建和删除子目录的能力。当实例被创建时，如果用户定义的目录存在并且有读写权限，那么会在该目录下创建一个子目录作为工作目录；否则，会在 `/tmp` 目录下创建一个子目录作为工作目录。当实例被销毁时，工作目录也会被销毁。用户函数可以通过 `INSTANCE_WORK_DIR` 环境变量获取工作目录。
      * - "DELEGATE_DIRECTORY_QUOTA"
        - 子目录配额大小，取值范围大于 ``0 M`` 且小于 ``1 TB``。如果未设置此配置，默认值为 ``512 M``。如果配置为 ``-1``，则不进行监控。单位：MB。
      * - "GRACEFUL_SHUTDOWN_TIME"
        - 自定义优雅退出时间，单位为秒。限制：``>=0``，``0`` 表示立即退出，不保证能够完成用户的优雅退出函数；如果配置小于 ``0``，则使用部署时的系统配置作为超时时间。
      * - "RECOVER_RETRY_TIMEOUT"
        - 自定义恢复超时时间。实例恢复超时时间，单位为毫秒。限制：``>0``，默认值为 ``10 * 60 * 1000``。

   当用作指标的用户自定义标签时：

   .. code-block:: python

       import yr
       yr.init()
       opt = yr.InvokeOptions()
       opt.custom_extensions["YR_Metrics"] = "{"endpoint":"127.0.0.1", "project_id":"my_project_id"}"

   在 Prometheus 中，选择指标名称为 `yr_app_instance_billing_invoke_latency`，您可以在收集到的调用信息中找到自定义标签信息：

   .. code-block:: python

       yr_app_instance_billing_invoke_latency{
       ...
       endpoint="127.0.0.1",
       ...}
