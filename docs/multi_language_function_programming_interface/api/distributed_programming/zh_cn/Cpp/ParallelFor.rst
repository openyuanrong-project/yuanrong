ParallelFor
=============

.. cpp:function:: template<typename Index, typename Handler>\
                  void YR::Parallel::ParallelFor(Index start, Index end, const Handler &handler, size_t chunkSize = -1, int workThreadSize = -1)

    ParallelFor 是一个并行计算函数框架，可将任务在多线程间并行执行，从而提升计算效率。

    ParallelFor 通过任务分配与调度在内部实现并行计算，自动将任务分发到可用线程。

    .. code-block:: cpp

       #include "yr/parallel/parallel_for.h"

       int main(int argc, char **argv)
       {
           YR::Config conf;
           conf.mode = YR::Config::Mode::LOCAL_MODE;
           const int threadPoolSize = 8;
           conf.threadPoolSize = threadPoolSize;
           YR::Init(conf);
           std::vector<uint64_t> results;
           const uint32_t start = 0;
           const uint32_t end = 1000000;
           results.resize(end);
           const uint32_t chunkSize = 1000;
           const int workerNum = 4;

           auto handler = [&results](size_t start, size_t end) {
               for (size_t i = start; i < end; i++) {
                   results[i] += i;
               }
           };
           YR::Parallel::ParallelFor<uint32_t>(start, end, handler, chunkSize, workerNum);

           std::vector<std::vector<int>> v(workerNum);
           auto f = [&v](int start, int end, const YR::Parallel::Context &ctx) {
               std::cout << "start: " << start << " , end: " << end << " ctx: " << ctx.id << std::endl;
               for (int i = start; i < end; i++) {
                   int result = i;
                   if (result) {
                       v[ctx.id].emplace_back(result);
                   }
               }
           };
           YR::Parallel::ParallelFor<uint32_t>(start, end, f, chunkSize, workerNum);
       }

    模板参数：
        - **Index** - 迭代变量的类型。
        - **Handler** - 待调用函数的类型。

    参数：
        - **start** - 循环迭代区间的起始值（包含）。
        - **end** - 循环迭代区间的结束值（不包含）。
        - **handler** - 循环体内要执行的函数。用户自定义 handler 的参数列表必须是以下两种形式之一：

          - (Index, Index)
          - (Index, Index, const YR::Parallel::Context&)，如果 handler 使用了 YR::Parallel::Context 参数，则 context.id 的值域为 [0, parallelism)。
        
        - **chunkSize** - 任务粒度。
        - **workThreadSize** - 工作线程数。设为 -1（默认值）时，将取线程池线程数加 1。
    
    抛出：
        :cpp:class:`Exception` - 以下两种情形：

        - 如果在初始化前调用 ParallelFor，将抛出异常：Assertion IsInitialized() failed !!!
        - 如果用户自定义 handler 的参数列表不符合上述格式，编译时报错：
          error: static assertion failed: handler must have 2 or 3 arguments. And arguments should be (Index, Index) or (Index, Index, const YR::Parallel::Context&)。

参数结构补充说明如下：

.. cpp:struct:: Context

    线程上下文信息。

    **公共成员**

    .. cpp:member:: size_t id

       线程标识 ID
