yr.java_function
=====================

.. py:function:: yr.java_function(class_name: str, function_name: str, function_urn: str) -> FunctionProxy

    用于构造 java 函数的代理，远程调用 java 函数。

    参数:
        - **class_name** (str) -  Java 类名。
        - **function_name** (str) - Java 函数的名称。
        - **function_urn** (str) - Java 函数的 URN（统一资源名称）。

    返回:
        被装饰函数的代理对象。
        数据类型：FunctionProxy。

    样例:

    .. code-block:: java

       package com.yuanrong.demo;

           public class PlusOne{
               public static int PlusOne(int x) {
                   System.out.println("PlusOne with x=" + x);
                   return x + 1;
               }
           }

    .. code-block:: python

       >>> import yr
       >>> yr.init()
       >>> java_function_urn = (
       ...     "sn:cn:yrk:12345678901234561234567890123456:"
       ...     "function:0-yr-defaultservice-java:$latest"
       ... )
       >>> java_add = yr.java_function("com.yuanrong.demo.PlusOne", "PlusOne", java_function_urn)
       >>> result = java_add.invoke(1)
       >>> print(yr.get(result))
