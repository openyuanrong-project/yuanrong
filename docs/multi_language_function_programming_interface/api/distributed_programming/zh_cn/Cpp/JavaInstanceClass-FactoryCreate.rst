JavaInstanceClass FactoryCreate
==================================

.. cpp:function:: static inline JavaInstanceClass\ 
                  YR::JavaInstanceClass::FactoryCreate(const std::string &className)

    为调用 Java 函数创建一个 `JavaInstanceClass` 对象。

    .. code-block:: cpp

       // Java code:
       package io.yuanrong.demo;
 
       // A regular Java class.
       public class Counter {
           private int value = 0;
 
           public int increment() {
                   this.value += 1;
                   return this.value;
               }
       }
 
       // C++ code:
       int main(void) {
           YR::Config conf;
           YR::Init(conf);
           auto javaInstance = YR::JavaInstanceClass::FactoryCreate("io.yuanrong.demo.Counter");
           auto r1 = YR::Instance(javaInstance).Invoke();
           auto r2 = r1.JavaFunction<int>("increment").Invoke(1);
           auto res = *YR::Get(r2);
 
           std::cout << "PlusOneWithJavaClass with result=" << res << std::endl;
           return res;
       }

    参数：
        - **className** - Java 函数的全限定类名，包括包名。如果该类是一个内部静态类，则使用 `$` 连接外部类和内部类。
  
    返回：
        一个携带创建 Java 函数类实例所需信息的 `JavaInstanceClass` 对象。该对象可以作为参数传递给 `YR::Instance <Instance.html>`_。
