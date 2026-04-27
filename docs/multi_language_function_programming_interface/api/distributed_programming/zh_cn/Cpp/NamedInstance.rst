NamedInstance
===============

具名实例，可调用该类的 ``Function`` 方法构造实例所属类的成员函数，适用于面向对象编程场景。

.. cpp:class:: template<typename InstanceType> NamedInstance
        
    模板参数：
        - **InstanceType** - 实例类型。

    **公共方法**

    .. cpp:function:: template<typename F> InstanceFunctionHandler<F, InstanceType> Function(F memberFunc)

        构造具名实例中的成员函数，用于执行一次函数调用。

        .. code-block:: cpp

           auto counter = YR::Instance(Counter::FactoryCreate, "name_1").Invoke(100);
           auto c = counter.Function(&Counter::Add).Invoke(1);
           std::cout << "counter is " << *YR::Get(c) << std::endl;

        模板参数：
           - **F** - 成员函数类型。

        参数：
           - **memberFunc** - 实例中需要执行的函数。实例绑定了类对象时，该函数必须是类的成员函数。

        返回：
           具名实例的函数 ``InstanceFunctionHandler`` 类，该类提供成员方法 ``Invoke`` 用于执行函数调用。

    .. cpp:function:: template<typename R> InstanceFunctionHandler<CppClassMethod<R>, InstanceType> CppFunction(const std::string &functionName)

        将 cpp 类成员函数调用请求发送给远程后端，由后端执行本次用户函数调用。

        .. code-block:: cpp
           
           class Counter {
           public:
               int count;

               Counter() {}

               explicit Counter(int init) : count(init) {}

               static Counter *FactoryCreate()
               {
                   int c = 10;
                   return new Counter(c);
               }

               std::string RemoteVersion()
               {
                   return "RemoteActor v0";
               }
           };

           YR_INVOKE(Counter::FactoryCreate, &Counter::RemoteVersion)

           int main(void)
           {
               YR::Config conf;
               YR::Init(conf);

               auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
               auto cppIns = YR::Instance(cppCls).Invoke();
               auto obj = cppIns.CppFunction<std::string>("&Counter::RemoteVersion").Invoke();
               auto res = *YR::Get(obj);
               std::cout << "add one result is " << res << std::endl;
               return 0;
           }

        模板参数：
           - **R** - 返回值的类型。

        参数：
           - **functionName** - cpp 类成员函数名。

        返回：
           ``InstanceFunctionHandler`` 对象，可通过调用其成员函数 ``Invoke`` 进行实例方法的调用。

    .. cpp:function:: template<typename R> InstanceFunctionHandler<PyClassMethod<R>, InstanceType> PyFunction(const std::string &functionName)

        将 python 类成员函数调用请求发送给远程后端，由后端执行本次用户函数调用。

        .. code-block:: cpp

           int main(void)
           {
               // class Instance:
               //     sum = 0
               //
               //     def __init__(self, init):
               //       self.sum = init
               //
               //     def add(self, a):
               //       self.sum += a
               //
               //     def get(self):
               //       return self.sum
               YR::Config conf;
               YR::Init(conf);

               auto pyInstance = YR::PyInstanceClass::FactoryCreate("pycallee", "Instance");  // moduleName, className
               auto r1 = YR::Instance(pyInstance).Invoke(x);
               r1.PyFunction<void>("add").Invoke(1);  // returnType, memberFunctionName

               auto r2 = r1.PyFunction<int>("get").Invoke();
               auto res = *YR::Get(r2);

               std::cout << "PlusOneWithPyClass with result=" << res << std::endl;
               return res;

               return 0;
           }

        模板参数：
           - **R** - 返回值的类型。

        参数：
           - **functionName** - python 类成员函数名。

        返回：
           ``InstanceFunctionHandler`` 对象，可通过调用其成员函数 ``Invoke`` 进行实例方法的调用。

    .. cpp:function:: template<typename R> InstanceFunctionHandler<JavaClassMethod<R>, InstanceType> JavaFunction(const std::string &functionName)

        将 java 类成员函数调用请求发送给远程后端，由后端执行本次用户函数调用。

        .. code-block:: cpp

           int main(void)
           {
               // package io.yuanrong.demo;
               //
               // // A regular Java class.
               // public class Counter {
               //
               //     private int value = 0;
               //
               //     public int increment() {
               //         this.value += 1;
               //         return this.value;
               //     }
               // }
               YR::Config conf;
               YR::Init(conf);
               auto javaInstance = YR::JavaInstanceClass::FactoryCreate("io.yuanrong.demo.Counter");
               auto r1 = YR::Instance(javaInstance).Invoke();
               auto r2 = r1.JavaFunction<int>("increment").Invoke(1);
               auto res = *YR::Get(r2);
               std::cout << "PlusOneWithJavaClass with result=" << res << std::endl;
               return res;
           }

        模板参数：
           - **R** - 返回值的类型。

        参数：
           - **functionName** - java 类成员函数名。

        返回：
           ``InstanceFunctionHandler`` 对象，可通过调用其成员函数 ``Invoke`` 进行实例方法的调用。
    
    .. cpp:function:: void Terminate()

        对于实例句柄，表示退出已创建的函数实例。对于 ``Range`` 调度的句柄，表示退出已创建的一组函数实例。
        
        退出默认超时时间为 30 秒，在磁盘高负载、etcd 故障等场景下，处理时间可能超出，接口将抛出超时异常，用户可在捕获超时异常后选择是否重试。

        .. code-block:: cpp

           auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
           auto c = counter.Function(&Counter::Add).Invoke(1);
           std::cout << "counter is " << *YR::Get(c) << std::endl;
           counter.Terminate();
        
        抛出：
            :cpp:class:`Exception` - 函数实例删除失败时抛出并携带错误信息。例如删除已提前退出的函数实例、收到删除函数实例响应超时等。

    .. cpp:function:: void Terminate(bool isSync)

        同步或异步退出实例。对于实例句柄，表示退出已创建的函数实例。对于 ``Range`` 调度的句柄，表示退出已创建的一组函数实例。
        
        异步退出默认超时时间为 30 秒，在磁盘高负载、etcd 故障等场景下，处理时间可能超出，接口将抛出超时异常，用户可在捕获超时异常后选择是否重试。同步退出时该接口将阻塞直至实例完全退出。

        .. code-block:: cpp

           auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
           auto c = counter.Function(&Counter::Add).Invoke(1);
           std::cout << "counter is " << *YR::Get(c) << std::endl;
           counter.Terminate(true);

        参数：
           - **isSync** - 是否开启同步。为 true 时表示发送信号量为 killInstanceSync 的请求，分布式内核将同步删除实例；若为 false 表示发送信号量为 killInstance 的请求，分布式内核将异步删除实例。
        
        抛出：
           :cpp:class:`Exception` - 函数实例删除失败时抛出并携带错误信息。例如删除已提前退出的函数实例、收到删除函数实例响应超时等。

    .. cpp:function:: std::vector<NamedInstance<InstanceType>> GetInstances(int timeoutSec = NO_TIMEOUT)

        在超时时间内等待 ``Range`` 调度完成，返回 :cpp:class:`NamedInstance` 对象的列表。
        
        .. code-block:: cpp

           auto counter = YR::Instance(Counter::FactoryCreate).Invoke(1);
           auto c = counter.Function(&Counter::Add).Invoke(1);
           std::cout << "counter is " << *YR::Get(c) << std::endl;
           counter.Terminate();

        参数：
           - **timeoutSec** - 超时时间，单位为秒。可选参数，默认值为 ``-1``，表示无限等待。

        抛出：
           :cpp:class:`Exception` - 超时时抛出。例如因资源不足导致实例未被调度，获取超时。

        返回：
           :cpp:class:`NamedInstance` 对象列表，可通过迭代获取各个实例的句柄以进行调用。

    .. cpp:function:: FormattedMap Export() const

        导出 :cpp:class:`NamedInstance` 句柄信息，可被序列化存储在数据库或其他持久化工具中。

        .. code-block:: cpp

           auto counter = YR::Instance(Counter::FactoryCreate).Invoke(100);
           auto out = counter.Export();

        返回：
           以键值对形式存储的 :cpp:class:`NamedInstance` 句柄信息。

    .. cpp:function:: Import(FormattedMap &input)

        导入从数据库或其他持久化工具中读取并反序列化后的 :cpp:class:`NamedInstance` 句柄信息。

        .. code-block:: cpp

           NamedInstance<Counter> counter;
           counter.Import(in);

        参数：
           - **input** - 以键值对形式存储的 :cpp:class:`NamedInstance` 句柄信息。
