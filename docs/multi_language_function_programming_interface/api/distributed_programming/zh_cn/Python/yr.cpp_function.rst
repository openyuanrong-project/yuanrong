.. _cpp_function:

yr.cpp_function
=====================

.. py:function:: yr.cpp_function(function_name: str, function_urn: str) -> FunctionProxy

    用于构造 cpp 函数的代理，远程调用 cpp 函数。

    参数:
        - **function_name** (str) - cpp 函数的名称。
        - **function_urn** (str) - cpp 函数的 URN（统一资源名称）。

    返回:
        返回远程 c++ 函数的代理对象。
        数据类型：FunctionProxy。

    样例：

    .. code-block:: cpp

       #include "yr/yr.h"

           int Square(int x){
               return x * x;
           }

           // 定义无状态函数 Square
           YR_INVOKE(Square)

    .. code-block:: python

       >>> import yr
       >>> yr.init()
       >>> cpp_function_urn = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mycpp:$latest"
       >>> square_func = yr.cpp_function("Square", cpp_function_urn)
       >>> result = square_func.invoke(5)
       >>> print(yr.get(result))
       >>>
       >>> yr.finalize()

