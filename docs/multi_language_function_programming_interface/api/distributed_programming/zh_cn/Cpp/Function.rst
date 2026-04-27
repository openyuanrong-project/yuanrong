Function
=========

.. cpp:function:: template<typename F> FunctionHandler<F> YR::Function(F f)

    为指定函数构造一个函数处理器。

    这个函数模板为指定的静态函数创建一个FunctionHandler对象。然后可以使用FunctionHandler对象来执行函数，并为函数调用设置各种选项，例如资源分配。

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

    模板参数：
        - **F** - 要执行的函数的类型。

    参数：
        - **f** - 要调用的静态函数。
    
    抛出：
        :cpp:class:`Exception` - 如果该函数未使用 `YR_INVOKE` 宏进行注册，则会抛出异常。
  
    返回：
        FunctionHandler<F>：一个提供方法来执行函数并为函数调用设置选项的 FunctionHandler 对象。