GetInstance
=============

.. cpp:function:: template<typename F>\
                  NamedInstance<F> YR::GetInstance(const std::string &name, const std::string &nameSpace = "", int timeoutSec = 60)

    在超时时间内获取与指定名称和命名空间关联的实例。

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

           int Counter::Add(int x)
           {
               count += x;
               return count;
           }
       };

       YR_INVOKE(Counter::FactoryCreate, &Counter::Add);

       int main(void)
       {
           YR::Config conf;
           YR::Init(conf);

           std::string name = "test-get-instance";
           auto ins = YR::Instance(Counter::FactoryCreate, name).Invoke(1);
           auto res = ins.Function(&Counter::Add).Invoke(1);

           std::string ns = "";
           auto insGet = YR::GetInstance<Counter>(name, ns, 60);
           auto resGet = insGet.Function(&Counter::Add).Invoke(1);
           return 0;
       }

    模板参数：
        - **F** - 实例类型。

    参数：
        - **name** - 具名实例的实例名称。
        - **nameSpace** - 具名实例的命名空间，默认值为 ``""``。
        - **timeoutSec** - 超时时间（以秒为单位）。默认值为 60。

    抛出：
        :cpp:class:`Exception` - 
        
        - name 不能为空，否则会抛出“name should not be empty”的异常。。
        - timeoutSec 必须大于 0，否则会抛出异常“timeout should be greater than 0”。
    
    返回：
        具名实例，可通过 ``Function`` 方法调用其成员函数。


