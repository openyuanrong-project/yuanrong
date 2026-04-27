YR_RECOVER
==================

.. cpp:function:: YR_RECOVER(...)
    
    用户可以使用此接口执行数据恢复操作。

    在实例恢复期间要执行的函数必须使用 `YR_RECOVER` 宏进行装饰。此宏用于装饰在云中执行的参与者内的用户定义的状态成员函数。
    使用 `YR_RECOVER` 装饰的函数可以使用 openYuanrong 接口。这些函数将在以下场景中执行：在运行时恢复请求以恢复实例期间。

    .. code-block:: cpp

       class Counter {
       public:
           int count;
           int recoverFlag = 0;
    
           Counter() {}
    
           explicit Counter(int init) : count(init) {}
    
           static Counter *FactoryCreate(int init)
           {
               return new Counter(init);
           }
           int SaveState();
    
           int LoadState();
    
           void Recover();
       };
    
       int Counter::SaveState()
       {
           YR::SaveState();
           return count;
       }
    
       int Counter::LoadState()
       {
           YR::LoadState();
           return count;
       }
    
       YR_INVOKE(&Counter::SaveState, &Counter::LoadState);
    
       void Counter::Recover()
       {
           std::cout << "recover" << std::endl;
           recoverFlag++;
       }
    
       YR_RECOVER(&Counter::Recover);

    抛出：
        :cpp:class:`Exception` - 如果检测到一个函数被多次注册，程序将退出并显示错误消息。
