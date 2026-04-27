CppInstanceClass FactoryCreate
==================================

.. cpp:function:: static inline CppInstanceClass\ 
                  YR::CppInstanceClass::FactoryCreate(const std::string &creatorName)

    为调用 C++ 函数创建一个 `CppInstanceClass` 对象。

    .. code-block:: cpp

       // C++ 类定义
       class Counter {
       public:
           int count;
    
           Counter()
           {}
    
           Counter(int init)
           {
               count = init;
           }
    
           static Counter* FactoryCreate()
           {
               return new Counter(10);
           }
    
           std::string RemoteVersion()
           {
               return "RemoteActor v0";
           }
       };
    
       YR_INVOKE(Counter::FactoryCreate, &Counter::RemoteVersion);
    
       // 用于创建和调用实例的 C++ 代码。
       int main(void) {
           YR::Config conf;
           YR::Init(conf);
    
           auto cppCls = YR::CppInstanceClass::FactoryCreate("Counter::FactoryCreate");
           auto cppIns = YR::Instance(cppCls).Invoke();
           auto obj = cppIns.CppFunction<std::string>("&Counter::RemoteVersion").Invoke();
           auto res = *YR::Get(obj);
           std::cout << "add one result is " << res << std::endl;
       }

    参数：
        - **creatorName** - C++ 类的构造函数或工厂函数的名称。
 
    返回：
        一个携带创建 C++ 函数类实例所需信息的 CppInstanceClass 对象。该对象可以作为参数传递给 `YR::Instance <Instance.html>`_。
