.. _runtime_env_IO:

yr.InvokeOptions.runtime_env
------------------------------

.. py:attribute:: InvokeOptions.runtime_env
   :type: Dict

   - `conda` 为 actor 提供不同的 Python 运行时环境。

     - 指定一个现有的 conda 环境（该环境在所有节点上都存在）``runtime_env = {"conda":"pytorch_p39"}``
     - 通过配置创建并使用 conda 环境。``runtime_env["conda"] = {"name":"myenv","channels": ["conda-forge"], "dependencies": ["python=3.9",
       "msgpack-python=1.0.5", "protobuf", "libgcc-ng", "cloudpickle=2.0.0", "cython=3.0.10", "pyyaml=6.0.2"]}``
     - 通过 YAML 文件创建并使用 conda 环境（YAML 文件符合 conda 要求）。``runtime_env = {"conda":"/home/env.yaml"}``

   - `pip` 为 Python 运行时环境安装依赖项。
   - `working_dir` 配置作业的代码路径。
   - `env_vars` 配置进程级环境变量。``runtime_env = {"env_vars":{"OMP_NUM_THREADS": "32", "TF_WARNINGS": "none"}}``

   其中 `runtime_env` 的约束条件如下：

     - `runtime_env` 支持的键是 `conda`、`env_vars`、`pip`、`working_dir`。其他键不会生效，也不会导致错误。
     - 使用 conda 运行 yr 函数。环境需要包含 yr 及其第三方依赖项。建议用户先创建一个 conda 环境，然后通过 `runtime_env` 指定它，例如：
       ``runtime_env = {"conda":"pytorch_p39"}``
     - `runtime_env` 支持使用配置创建和切换 conda 环境。配置需要安装 yr 的第三方依赖项，例如：
       ``runtime_env["conda"] = {"name":"myenv","channels": ["conda-forge"], "dependencies": ["python=3.9",
       "msgpack-python=1.0.5", "protobuf", "libgcc-ng", "cloudpickle=2.0.0", "cython=3.0.10", "pyyaml=6.0.2"]}``
     - 用户需要清理在 `runtime_env` 中使用 conda 创建的环境。
     - 在 `runtime_env` 中，conda 可以使用 `pip` 安装依赖项，这些依赖项由 conda 直接管理。
       ``runtime_env = {"conda":{'name': 'my_project_env', 'channels': ['defaults', 'conda-forge'],
       'dependencies': ['python=3.9', {'pip': ['requests==2.25.1']}]}}``
     - 目前，提供了 Python 3.9 和 Python 3.11 SDK。conda 的 Python 版本需要与 SDK 版本一致。
     - 如果同时配置了 `InvokeOptions.env_vars` 和 `InvokeOptions.runtime_env["env_vars"]` 中的相同键，
       则使用 `InvokeOptions.env_vars` 中的配置。
     - 如果配置了 `InvokeOptions.runtime_env["working_dir"]`，则使用此配置，
       否则，使用 `YR.Config.working_dir`，最后使用 `InvokeOptions.env_vars` 中的配置。
     - 如果使用 conda，需要指定环境变量 `YR_CONDA_HOME` 指向安装路径。