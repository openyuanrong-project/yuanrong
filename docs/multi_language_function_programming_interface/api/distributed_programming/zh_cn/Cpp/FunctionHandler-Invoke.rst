FunctionHandler::Invoke
=========================

.. cpp:function:: template<typename ...Args>\
                  inline ObjectRef<ReturnType<F>>\
                  YR::FunctionHandler::Invoke(Args&&... args)

    指定参数调用已注册的函数。

    Invoke 方法发送请求到后端，后端使用指定的参数执行已注册的函数。函数必须使用 YR_INVOKE 宏进行注册。\
    传递给 Invoke 方法的参数必须与函数的参数类型和数量匹配，否则将发生编译时错误。\
    请确保参数类型完全匹配，避免因隐式类型转换引发问题。

    .. code-block:: cpp

       #include "yr/yr.h"

       int PlusOne(int x)
       {
           return x + 1;
       }

       YR_INVOKE(PlusOne)

       int main()
       {
           YR::Config conf;
           YR::Init(conf);

           std::vector<YR::ObjectRef<int>> n2;
           int k = 1;
           for (int i = 0; i < k; i++) {
               auto r2 = YR::Function(PlusOne).Invoke(2);
               n2.emplace_back(r2);
           }
           for (int i = 0; i < k; i++) {
               auto integer = *YR::Get(n2[i]);
               printf("%d :%d\n", i, integer);
           }

           YR::Finalize();
           return 0;
       }

    **链式调用用法：** 在单机模式下，Invoke 方法可以链式执行多个函数调用。但要注意，如果函数执行实例的总数超过了线程池的大小，程序可能会被阻塞。\
    建议根据预期的函数调用次数来配置线程池的大小。例如，如下示例链式函数调用涉及三层（AddThree、AddTwo 和 AddOne），\
    每次调用都会创建一个新线程，因此线程池的大小应至少配置为 3。

    .. code-block:: cpp
   
       int AddOne(int x)
       {
           return x + 1;
       }

       int AddTwo(int x)
       {
           auto obj = YR::Function(AddOne).Invoke(x);
           auto res = *YR::Get(obj);
           return res + 1;
       }

       int AddThree(int x)
       {
           auto obj = YR::Function(AddTwo).Invoke(x);
           auto res = *YR::Get(obj);
           return res + 1;
       }

       YR_INVOKE(AddOne, AddTwo, AddThree);

       int main(void) {
           YR::Config conf;
           conf.mode = YR::Config::Mode::LOCAL_MODE;
           conf.threadPoolSize = 3;
           YR::Init(conf);
           auto r = YR::Function(AddThree).Invoke(5);
           int ret = *(YR::Get(r)); // ret = 8
           return 0;
       }
       YR_INVOKE(HelloWorld)

    有关配置线程池的更多详细信息，请参阅 :doc:`结构体说明 <./struct-Config>`。

    模板参数：
        - **Args** - 传递给函数的参数类型。

    参数：
        - **args** - 传递给函数的参数。参数的类型和数量必须与函数的签名匹配。

    抛出：
        :cpp:class:`Exception` - 如果调用失败，例如由于创建函数实例时出错或资源不足会抛出异常。
  
    返回：
        ObjectRef<ReturnType<F>>：对结果对象的引用，本质上是一个键。要获取实际的值，请使用 `YR::Get <Get.html>`_ 方法。

