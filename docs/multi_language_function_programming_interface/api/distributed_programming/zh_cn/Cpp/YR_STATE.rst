YR_STATE
==================

.. cpp:function:: YR_STATE(...)
    
    将类的参数标记为状态，并且 openYuanrong 将自动保存和恢复类的状态成员变量。

    需要标记为状态的类成员变量应使用 `YR_STATE` 进行装饰。约束：`YR_STATE` 必须在公共部分声明。

    .. code-block:: cpp

       class MyClass {
       public:
           int a;
           std::string b;
           YR_STATE(a, b);
       };
