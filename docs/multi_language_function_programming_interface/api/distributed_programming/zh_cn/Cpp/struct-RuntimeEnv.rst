RuntimeEnv
==========

.. cpp:class:: RuntimeEnv

    此类提供用于为 actor 设置 runtime env 的接口。

    **公共函数**

    .. cpp:function:: template<typename T> inline void Set(const std::string& name, const T& value)

        设置运行时环境的名称和对象。

        .. code-block:: cpp

            #include <iostream>
            #include "yr/yr.h"
            #include "yr/api/runtime_env.h"
            int main(int argc, char **argv) {
                std::string pyFunctionUrn = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mypython:$latest";
                YR::Config conf;
                YR::Init(conf, argc, argv);
                YR::InvokeOptions opts;
                YR::RuntimeEnv runtimeEnv;
                runtimeEnv.Set<std::string>("conda", "pytorch_p39");
                runtimeEnv.Set<std::map<std::string, std::string>>("env_vars", {{"OMP_NUM_THREADS", "32"}, {"TF_WARNINGS", "none"}, {"YR_CONDA_HOME", "/home/snuser/.conda"}});
                opts.runtimeEnv = runtimeEnv;
                auto resFutureSquare = YR::PyFunction<int>("calculator", "square").SetUrn(pyFunctionUrn).Options(opts).Invoke(2);
                auto resSquare = *YR::Get(resFutureSquare);
                std::cout << resSquare << std::endl;
                return 0;
            }

        .. note::

            限制条件

            - :cpp:class:`RuntimeEnv` 仅处理以下命名键值，其他键值既不会生效也不会报错：

            - **pip** ``String[]/Iterable<String>`` : 指定环境的 Python 依赖（该键与 conda 键互斥）。

              .. code-block:: cpp

                  YR::RuntimeEnv env;
                  env.Set<std::vector<std::string>>("pip", {"numpy=2.3.0", "pandas"});

            - **working_dir** ``str`` : 指定的代码路径，当前仅支持 C++ actor。

              .. code-block:: cpp

                  YR::RuntimeEnv env;
                  env.Set<std::string>("working_dir", "/opt/mycode/cpp-invoke-python/calculator");

            - **env_vars** ``JSON`` : 环境变量（值必须为字符串类型）。

              .. code-block:: cpp

                  YR::RuntimeEnv runtimeEnv;
                  runtimeEnv.Set<std::map<std::string, std::string>>("env_vars", {{"OMP_NUM_THREADS", "32"}, {"TF_WARNINGS", "none"}});

            - **conda** ``str/JSON`` : Conda 配置（需要设置 YR_CONDA_HOME 环境变量）。

              (1). 指定调度到已存在的 conda 环境。

              .. code-block:: cpp

                  YR::RuntimeEnv env;
                  runtimeEnv.Set<std::string>("conda", "pytorch_p39");

              (2). 创建新的环境并指定环境的依赖以及环境名（可选）。

              .. code-block:: cpp

                  runtimeEnv.Set<nlohmann::json>("conda", {{"name", "pytorch_p39"},{"channels", {"conda-forge"}}, {"dependencies", {"python=3.9", "matplotlib", "msgpack-python=1.0.5", "protobuf", "libgcc-ng", "numpy", "pandas", "cloudpickle=2.0.0", "cython=3.0.10", "pyyaml=6.0.2"}}});
                  runtimeEnv.Set<std::map<std::string, std::string>>("env_vars", {{"OMP_NUM_THREADS", "32"}, {"TF_WARNINGS", "none"}, {"YR_CONDA_HOME", "/home/snuser/.conda"}});

              (3). 创建新的环境，并且环境依赖以及环境通过文件指定。

              .. code-block:: cpp

                  YR::RuntimeEnv runtimeEnv;
                  /* yaml file demo
                  * name: myenv3
                  * channels:
                  *  - conda-forge
                  *  - defaults
                  * dependencies:
                  *  - python=3.9
                  *  - numpy
                  *  - pandas
                  *  - cloudpickle=2.2.1
                  *  - msgpack-python=1.0.5
                  *  - protobuf
                  *  - cython=3.0.10
                  *  - pyyaml=6.0.2
                  */
                  runtimeEnv.Set<std::string>("conda", "/opt/conda/env-xpf.yaml");

            - **venv** ``JSON`` : Python 内置虚拟环境 venv 配置（该键与 pip 键、conda 键互斥）。

              (1). 通过 pip install 下载依赖包。

              .. code-block:: cpp

                  YR::RuntimeEnv runtimeEnv;
                  runtimeEnv.Set<nlohmann::json>("venv", {
                    // 指定虚拟环境名称，不填或者填""时，元戎会自动生成 uuid 作为虚拟环境名称
                    {"name", "testVenv"},
                  	// 配置 pip 下载到虚拟环境的参数，选填
                  	{"dependencies", {
                  		// 待 pip install 的包名，选填
                  		// 支持纯包名(requests)、精确版本(requests==2.28.1)、范围约束(requests>=2.0,<3.0)，不支持填写 url
                  		{"pypi", {"pandas", "pyarrow", "requests", "cloudpickle==3.0.0"}},
                  		// 选填，将指定的主机（或"主机:端口"）标记为可信，即使它没有有效的 HTTPS 证书，或者完全使用  HTTP （非加密）协议，不支持填写多个地址
                  		{"trust_host", "mirrors.tools.nobody.com"},
                  		// 选填，指定 pip 安装包时使用的 PyPI 镜像源地址，以替代默认的官方源，不支持填写多个地址
                  		{"index_url", "http://mirrors.tools.nobody.com/pypi/simple/"}
                  	}}
                  });
                  
              (2). 通过指定 obs 地址，复用 site-packages，跳过 pip 安装包。

              .. code-block:: text

                 YR::InvokeOptions opts;
                 YR::RuntimeEnv runtimeEnv;

                 runtimeEnv.Set<nlohmann::json>("venv", {
                    // 指定虚拟环境名称，不填或者填""时，元戎会自动生成 uuid 作为虚拟环境名称
                    {"name", "testVenv"},
                    {"path", {
                 	     // 指定待下载的 site-packages 地址，以 obs:// 开头
                 	     // bucketId 与 objectId 需要和 opts.customExtensions["DELEGATE_DOWNLOAD"] 保持一致，否则无法正常下载
                 	     {"site_package_path", "obs://bucketId/objectId"}
                 	}},
                 });

                 opts.runtimeEnv = runtimeEnv;
                 // 配置 site_package_path 时必填，否则无法下载
                 // DELEGATE_DOWNLOAD 配置参考 InvokeOptions 结构体章节内容
                 opts.customExtensions["DELEGATE_DOWNLOAD"] = "{\"storage_type\":\"s3\",\"hostName\":\"obs.xxxx.com\",\"bucketId\":\"bucketId\",\"objectId\":\"objectid\",\"sha256\":\"xxxxx\",
                 \"temporaryAccessKey\":\"HST3UXZO1UWEXG6ZGUPV\",\"temporarySecretKey\":\"xxxxxx\",\"securityToken\":\"xxxxxxxx\"}";
              
              (3). 通过指定 obs 地址，复用 site-packages，增量安装 PyPI 包。

              .. code-block:: cpp

                  YR::InvokeOptions opts;
                  YR::RuntimeEnv runtimeEnv;

                  runtimeEnv.Set<nlohmann::json>("venv", {
                  	{"name", "testVenv"},
                  	{"dependencies", {
                  	      // 基于已存在的 site-packages 包，新增下载 pandas ，由用户保证包之间无版本冲突
                  	      {"pypi", {"pandas"},
                  	      {"trust_host", "mirrors.tools.nobody.com"},
                  	      {"index_url", "http://mirrors.tools.nobody.com/pypi/simple/"}
                  	}}
                  	{"path", {
                  	    {"site_package_path", "obs://bucketId/site-package.zip"}
                  	}},
                  });

                  opts.runtimeEnv = runtimeEnv;
                  opts.customExtensions["DELEGATE_DOWNLOAD"] = "{\"storage_type\":\"s3\",\"hostName\":\"obs.xxxx.com\",\"bucketId\":\"bucketId\",\"objectId\":\"site-package.zip\",\"sha256\":\"xxxxx\",\"temporaryAccessKey\":\"HST3UXZO1UWEXG6ZGUPV\",\"temporarySecretKey\":\"xxxxxx\",\"securityToken\":\"xxxxxxxx\"}";


        :模板参数:
            - **T** - 第二个入参的类型。
        :参数:
            - **name** - 运行时环境的参数名。
            - **value** - nlohmann/json 类型的可 json 的对象。

    .. cpp:function:: template<typename T> inline T Get(const std::string& name) const

        获取运行时环境字段值。

        :模板参数:
            - **T** - 返回值的类型。
        :参数:
            - **name** - 运行时环境的参数名。
        :返回:
            类型为 T 的运行时环境的参数值。

    .. cpp:function:: void SetJsonStr(const std::string& name, const std::string& jsonStr)

        通过 JSON 字符串设置运行时环境字段。

        :参数:
            - **name** - 要设置的运行时环境的参数名。
            - **jsonStr** - 表示字段值的 JSON 格式字符串。

    .. cpp:function:: std::string GetJsonStr(const std::string& name) const

        获取指定字段的 JSON 字符串表示。

        :参数:
            - **name** - 运行时环境的参数名。
        :返回:
            字段的 JSON 字符串表示。

    .. cpp:function:: bool Contains(const std::string& name) const

        检查是否包含指定字段。

        :参数:
            - **name** - 要检查的字段名称。
        :返回:
            如果包含返回 ``true`` ，否则返回 ``false`` 。

    .. cpp:function:: bool Remove(const std::string& name)

        移除指定字段。

        :参数:
            - **name** - 要移除的字段名称
        :返回:
            如果成功移除返回 ``true`` ，如果字段不存在返回 ``false`` 。

    .. cpp:function:: bool Empty() const

        检查运行时环境是否为空。

        :返回:
            如果没有设置任何字段返回 ``true`` ，否则返回 ``false`` 。