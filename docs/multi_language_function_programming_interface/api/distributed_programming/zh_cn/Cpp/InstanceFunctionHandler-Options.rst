InstanceFunctionHandler::Options
=================================

.. cpp:function:: inline InstanceFunctionHandler<F, T>\
                  &YR::InstanceFunctionHandler::Options(InvokeOptions &&optsInput)

    设置函数调用的选项，例如超时时间和重试次数。

    此方法允许你为函数调用配置特定选项，例如设置超时时间或指定重试次数。
    这些选项对于控制远程函数调用的行为至关重要。但请注意，此方法在本地模式下无效。

    .. code-block:: cpp

       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           YR::InvokeOptions opts;
           opts.retryTimes = 5;
           auto ins = YR::Instance(SimpleCaculator::Constructor).Invoke();
           auto r3 = ins.Function(&SimpleCaculator::Plus).Options(opts).Invoke(1, 1);

           return 0;
       }

    .. warning::
        此方法专为分布式环境设计，在本地模式下不起作用。请确保在适用调用选项的适当上下文中使用它。

    模板参数：
        - **Args** - 传递给函数的参数类型。

    参数：
        - **optsInput** - 调用选项，包括超时时间、重试次数等。详细描述请参考 `数据结构文档 <struct-Config.html>`_ 中的 struct InvokeOptions 章节。

    返回：
        InstanceFunctionHandler<F, T>&，返回 ``InstanceFunctionHandler`` 对象的引用，方便直接调用 ``Invoke`` 接口。