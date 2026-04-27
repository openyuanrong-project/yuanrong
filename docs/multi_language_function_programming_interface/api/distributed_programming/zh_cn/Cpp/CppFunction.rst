CppFunction
=============

.. cpp:function:: template<typename R>\
                  FunctionHandler<CppFunctionHandler<R>> YR::CppFunction(const std::string &funcName)

    为通过名称调用 C++ 函数创建一个 `FunctionHandler`。

    .. code-block:: cpp

       // 调用的 C++ 函数
       int PlusOne(int x)
       {
           return x + 1;
       }
    
       YR_INVOKE(PlusOne);
    
       // 用于调用函数的 C++ 代码。
       int main(void)
       {
           YR::Config conf;
           YR::Init(conf);
           auto ref = YR::CppFunction<int>("PlusOne").Invoke(1);
           int res = *YR::Get(ref);   // get 2
           return 0;
       }

    模板参数：
        - **R** - C++ 函数的返回类型。

    参数：
        - **funcName** - 要调用的 C++ 函数的名称。

    返回：
        一个可以用来执行 C++ 函数的 `FunctionHandler` 对象。可以通过 `CppFunctionHandler` 模板类来访问返回类型。