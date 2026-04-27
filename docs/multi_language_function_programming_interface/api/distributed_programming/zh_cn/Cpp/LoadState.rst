LoadState
==================

.. cpp:function:: void YR::LoadState(const int &timeout = DEFAULT_SAVE_LOAD_STATE_TIMEOUT)
    
    保存实例状态。

    .. code-block:: cpp

       #include "yr/yr.h"

       class Counter {
       public:
           Counter() {}
           ~Counter() {}
           explicit Counter(int init) : count(init) {}

           static Counter *FactoryCreate(int init)
           {
               return new Counter(init);
           }

           int Save()
           {
               YR::SaveState();
               return count;
           }

           int Load()
           {
               YR::LoadState();
               return count;
           }

           int Add(int x)
           {
               count += x;
               return count;
           }

           YR_STATE(count)

       private:
           int count;
       };

       YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Save, &Counter::Load)

       int main()
       {
           YR::Config conf;
           conf.mode = YR::Config::Mode::CLUSTER_MODE;
           YR::Init(conf);
           auto creator1 = YR::Instance(Counter::FactoryCreate).Invoke(1);
           auto member1 = creator1.Function(&Counter::Add).Invoke(3);
           auto res1 = *YR::Get(member1);
           printf("res1 is %d\n", res1);  // 4
           auto member2 = creator1.Function(&Counter::Save).Invoke();
           auto res2 = *YR::Get(member2);
           printf("res2 is %d\n", res2);  // 4
           auto member3 = creator1.Function(&Counter::Add).Invoke(3);
           auto res3 = *YR::Get(member3);
           printf("ref3 is %d\n", res3);  // 7
           auto member4 = creator1.Function(&Counter::Load).Invoke();
           auto res4 = *YR::Get(member4);
           printf("res4 is %d\n", res4);  // 7
           auto member5 = creator1.Function(&Counter::Add).Invoke(3);
           auto res5 = *YR::Get(member5);
           printf("res5 is %d\n", res5);  // 7
           YR::Finalize();
           return 0;
       }

    参数：
        - **timeout** - 超时时间，以秒为单位，默认 ``30``。
    
    抛出：
        :cpp:class:`Exception` -

        - 本地模式不支持 SaveState，将抛出异常：“LoadState is not supported in local mode”。
        - 集群模式仅在远程代码中支持此操作；如果在本地代码中调用，将抛出异常：“LoadState is only supported on cloud with posix api”。
