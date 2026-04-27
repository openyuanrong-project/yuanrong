JavaFunction
=============

.. cpp:function:: template<typename R>\
                  FunctionHandler<JavaFunctionHandler<R>>\ 
                  YR::JavaFunction(const std::string &className, const std::string &functionName)

    为从 C++ 调用 Java 函数创建一个 `FunctionHandler`。

    .. code-block:: cpp

       // Java code:
       package com.example;
    
       public class YrlibHandler {
       public static class MyYRApp {
           public static int add_one(int x) {
                   return x + 1;
               }
           }
       }
    
       // C++ code:
       int main(void) {
           YR::Config conf;
           YR::Init(conf);
           auto ref = YR::JavaFunction<int>("com.example.YrlibHandler$MyYRApp", "add_one").Invoke(1);
           int res = *YR::Get(ref);   // get 2
           return 0;
       }

    模板参数：
        - **R** - 函数的返回类型。

    参数：
        - **className** - Java 函数的全限定类名，包括包名。如果该类是一个内部静态类，则使用 `$` 连接外部类和内部类。
        - **functionName** - 要调用的 Java 函数的名称。
  
    返回：
        一个可以用来执行 Java 函数的 `FunctionHandler` 对象。可以通过 `JavaFunctionHandler` 模板类来访问返回类型。