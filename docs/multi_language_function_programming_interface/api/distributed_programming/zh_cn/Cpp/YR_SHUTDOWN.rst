YR_SHUTDOWN
==================

.. cpp:function:: YR_SHUTDOWN(...)
    
    用户可以使用此接口执行数据清理或数据持久化操作。

    在优雅关闭期间要执行的函数必须使用 `YR_SHUTDOWN` 进行装饰。该函数必须且只能有一个类型为 `uint64_t` 的参数。否则，由于参数不匹配，该函数将无法在云环境中执行。这些函数用于装饰在云中执行的参与者内的用户定义的优雅关闭成员函数。这些函数可以使用 openYuanrong 接口。使用 `YR_SHUTDOWN` 装饰的函数将在以下场景中执行：

    - 当运行时接收到关闭请求时（例如，当用户在云环境中调用终止接口时）。
    - 当运行时捕获到 `Sigterm` 信号时（例如，在集群缩容期间，运行时管理器捕获 `Sigterm` 信号并将其转发到运行时）。
  
    .. code-block:: cpp

       class Counter {
       public:
           int count;
    
           Counter() {}
           explicit Counter(int init) : count(init) {}
    
           static Counter *FactoryCreate(int init)
           {
               return new Counter(init);
           }
    
           void Counter::MyShutdown(uint64_t gracePeriodSecond)
           {
               YR::KV().Set("myKey", "myValue");
           }
    
           int Add(int x)
           {
               count += x;
               return count;
           }
       };
    
       YR_INVOKE(Counter::FactoryCreate, &Counter::Add)
       YR_SHUTDOWN(&Counter::MyShutdown);
    
       int main(int argc, char **argv)
       {
           YR::KV().Del("myKey");
           auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
           auto ret = counter.Function(&Counter::Add).Invoke(1);
           std::cout << *YR::Get(ret) << std::endl; // 2
           // 云下使用 kv 接口获取 my_key 映射的值将为 'my_value'
           counter.Terminate();
           std::string result = YR::KV().Get("myKey", 30);
           EXPECT_EQ(result, "myValue");
    
           YR::KV().Del("myKey");
           YR::InvokeOptions opt;
           opt.customExtensions.insert({"GRACEFUL_SHUTDOWN_TIME", "10"});
           auto counter2 = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke(1);
           auto ret2 = counter2.Function(&Counter::Add).Invoke(1);
           std::cout << *YR::Get(ret2) << std::endl; // 2
    
           counter2.Terminate();
           std::string result2 = YR::KV().Get("myKey", 30);
           EXPECT_EQ(result2, "myValue");
       }
    
    .. note::
        - 在定义自定义优雅关闭函数时，函数必须且只能有一个名为 `gracePeriodSeconds` 的参数，其类型为 `uint64_t`。否则，由于参数不匹配，该函数将无法在云环境中执行。
        - 如果自定义优雅关闭函数的执行时间超过 `gracePeriodSeconds` 秒，云参与者将不会等待该函数完成，而是会立即被回收。