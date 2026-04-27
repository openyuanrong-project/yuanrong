YR_INVOKE
===========

.. cpp:function:: YR_INVOKE(...)

    openYuanrong 分布式调用注册函数。

    在本地模式下，注册的函数在当前进程中执行；在集群模式下，函数在远程执行。
    所有用于远程执行的函数都需通过此接口进行注册。如果一个函数被多次注册，程序将在运行时抛出异常并终止。

    .. code-block:: cpp

       int AddOne(int x)
       {
           return x + 1;
       }
    
       YR_INVOKE(AddOne);
    
       class Counter {
       public:
           Counter() {}
           Counter(int init)
           {
               count = init;
           }
    
           static Counter *FactoryCreate(int init)
           {
               return new Counter(init);
           }
    
           int Add(int x)
           {
               count += x;
               return count;
           }
    
           int Get(void)
           {
               return count;
           }
    
           YR_STATE(count);
    
           public:
               int count;
           };
    
           YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Get);

    .. note::
        当通过 `YR_INVOKE` 在远程注册的函数中使用 `printf` 时，需要注意 openYuanrong 的运行时内核会重定向标准输出，
        并将其切换到完全缓冲模式。这意味着只有当缓冲区满或显式调用 `fflush` 时，输出才会写入磁盘，换行符也不会立即打印。
        例如：

        .. code-block:: cpp

           int HelloWorld()
           {
               printf("helloworld\n"); // Not recommended; output may not immediately appear due to remote runtime buffering
               return 0;
            }
            YR_INVOKE(HelloWorld)
        
        因此，建议使用 `std::cout` 或显式调用 `fflush`：

        .. code-block:: cpp

           int HelloWorld()
           {
               std::cout << "helloworld" << std::endl; // Recommended
               // 或者显式调用 `fflush`。
               printf("helloworld\n");
               fflush(stdout);
               return 0;
           }
           YR_INVOKE(HelloWorld)

    抛出：
        :cpp:class:`Exception` - 如果检测到一个函数被多次注册，程序将退出并显示错误消息。