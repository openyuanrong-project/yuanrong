FunctionHandler::SetUrn
=========================

.. cpp:function:: inline FunctionHandler<F> &YR::FunctionHandler::SetUrn(const std::string &urn)

    为当前函数调用设置 URN（统一资源名称），它可以与 `CppFunction` 或 `JavaFunction` 一起使用。

    .. code-block:: cpp

       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           auto r1 =
           YR::CppFunction<int>("PlusOne").SetUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").Invoke(2);
           YR::Get(r1);
           return 0;
       }

    模板参数：
        - **F** - 要执行的函数的类型。

    参数：
        - **urn** - 函数的 URN，它应该与配置中设置的租户 ID 匹配。

    返回：
        FunctionHandler<F>&：一个对函数处理器对象的引用，便于直接调用 `Invoke` 接口。