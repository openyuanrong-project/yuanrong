yr.java_instance_class
---------------------------------

.. py:method:: java_instance_class(class_name: str, function_urn: str) -> InstanceCreator

    用于构造 Java 类，并远程调用 Java 类的代理。

    参数：
        - **class_name** (str) – java 类名。
        - **function_urn** (str，可选) – 函数 urn，默认 ``sn:cn:yrk:default:function:0-defaultservice-java:$latest``。

    返回：
        对应的实例创建器。

    样例：

    .. code-block:: java

       package org.yuanrong.demo;

       public class Counter {
           private int count;
        
           public Counter(int init) {
               System.out.println("Counter constructor with init=" + init);
               this.count = init;
           }
        
           public int Add(int value) {
               this.count += value;
               System.out.println("Add called, new count=" + this.count);
               return this.count;
           }
        
           public int Get() {
               System.out.println("Get called, count=" + this.count);
               return this.count;
           }
       }

    .. code-block:: python

       >>> import yr
       >>> yr.init()
       >>> java_function_urn = (
       ...     "sn:cn:yrk:default:"
       ...     "function:0-yr-defaultservice-java:$latest"
       ... )
       >>>
       >>> java_instance = yr.java_instance_class("org.yuanrong.demo.Counter", java_function_urn).invoke(1)
       >>> res = java_instance.Add.invoke(5)
       >>> print(yr.get(res))
       >>>
       >>> res = java_instance.Get.invoke()
       >>> print(yr.get(res))
       >>>
       >>> java_instance.terminate()
       >>> yr.finalize()
