InstanceCreator::SetUrn
=========================

.. cpp:function:: inline InstanceCreator<Creator> &YR::internal::InstanceCreator::SetUrn(const std::string &urn)

    为创建的实例设置函数 URN。

    .. code-block:: cpp

       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
           auto cppIns = YR::Instance(cppCls)
                           .SetUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
                           .Invoke(1);
           auto obj = cppIns.CppFunction<int>("&Counter::Add").Invoke(1);

           return 0;
       }

    .. warning::
        确保提供的 URN 中的租户 ID 与配置中设置的租户 ID 一致。租户 ID 不匹配可能导致错误或意外行为。

    .. note::
        接口配合 ``CppInstanceClass`` 或 ``JavaInstanceClass`` 使用。URN 格式应遵循指定的结构，并且租户 ID 必须正确配置以确保功能正常运行。

    模板参数：
        - **Creator** - 创建者类型，用于构建有状态函数实例。

    参数：
        - **urn** - 函数 URN，在函数部署后获取。URN 中的租户 ID 必须与 :doc:`结构体说明 <./struct-Config>` 中 Config 配置的租户 ID 匹配。

    返回：
        InstanceCreator<Creator>&： ``InstanceCreator`` 对象的引用，便于直接调用 ``Invoke`` 方法。

