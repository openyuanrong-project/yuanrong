FunctionHandler::Options
=========================

.. cpp:function:: inline FunctionHandler<F>\
                  &YR::FunctionHandler::Options(const InvokeOptions &optsInput)

    为当前调用设置选项，包括资源等。此接口在本地模式下无效。

    无状态函数重试

    .. code-block:: cpp

       // 静态函数
       int AddOne(int x)
       {
           return x + 1;
       }

       void ThrowRuntimeError()
       {
           throw std::runtime_error("123");
       }

       YR_INVOKE(AddOne, ThrowRuntimeError);

       bool RetryChecker(const YR::Exception &e) noexcept
       {
           if (e.Code() == YR::ErrorCode::ERR_USER_FUNCTION_EXCEPTION) {
                std::string msg = e.what();
                if (msg.find("123") != std::string::npos) {
                    return true;
                }
           }
           return false;
       }
       int main(void) {
           YR::Config conf;
           YR::Init(conf);
           YR::InvokeOptions opts;

           opts.cpu = 1000;
           opts.memory = 1000;
           auto r1 = YR::Function(AddOne).Options(opts).Invoke(5);
           YR::Get(r1);

           opts.retryTimes = 1;
           opts.retryChecker = RetryChecker;
           auto r2 = YR::Function(ThrowRuntimeError).Options(opts).Invoke();
           YR::Get(r2);

           return 0;
       }

    无状态函数指定本地执行

    .. code-block:: cpp
   
       int CallLocal(int x)
       {
           YR::InvokeOptions opt;
           opt.alwaysLocalMode = true;
           auto obj = YR::Function(AddOne).Options(opt).Invoke(x);
           int ret = *YR::Get(obj);
           std::cout << "CallLocal result: " << ret << std::endl;
           return ret;
       }
    
       int CallCluster(int x)
       {
           auto obj = YR::Function(AddOne).Invoke(x);
           int ret = *YR::Get(obj);
           std::cout << "CallCluster result: " << ret << std::endl;
           return ret;
       }
    
       YR_INVOKE(CallLocal, CallCluster)
    
       void HybridClusterCallLocal()
       {
           auto obj = YR::Function(CallLocal).Invoke(1);
           std::cout << *YR::Get(obj) << " == " << 2 << std::endl;
       }
    
       void HybridLocalCallCluster()
       {
           YR::InvokeOptions opt;
           opt.alwaysLocalMode = true;
           auto obj = YR::Function(CallCluster).Options(opt).Invoke(1);
           std::cout << *YR::Get(obj) << " == " << 2 << std::endl;
       }
    
       int main()
       {
           HybridClusterCallLocal();
           HybridLocalCallCluster();
           return 0;
       }

    模板参数：
        - **F** - 要执行的函数的类型。

    参数：
        - **optsInput** - 调用选项。详细描述请参考 :doc:`结构体说明 <./struct-Config>`。

    返回：
        FunctionHandler<F>&：一个对函数处理器对象的引用，便于直接调用 `Invoke` 接口。