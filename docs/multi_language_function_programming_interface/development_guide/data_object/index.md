# 数据对象

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   KV
```

数据对象是可以在多个 openYuanrong 函数间跨节点分布式共享访问的内存数据，支持基于共享内存的高性能 put/get 访问和修改。openYuanrong 将远程对象缓存在分布式共享内存对象存储中，对象存储在集群的每个节点上都会创建一个。

对象引用本质上是一个指针或唯一的 ID，用于引用远程对象而无需获取它实际的值，它和 Future 在概念上有类似之处。openYuanrong 基于引用计数进行数据对象自动生命周期管理，当引用计数为 0 时自动删除数据对象。

## 使用限制

- 数据对象不能被修改，如果您需要修改，请创建新的数据对象。
- 单个数据对象的大小除 Java 语言外没有限制，但是不能超出 openYuanrong 数据系统配置的共享内存大小。Java 语言的单个数据对象大小不能超过 `Integer.MAX_VALUE - 8`。
- 数据对象不保证数据可靠性，当发生故障数据可能会丢失。
- 使用 `get` 接口批量获取数据时，一次获取的总数据大小不能超出 openYuanrong 数据系统的共享内存大小。

## 创建数据对象

数据对象可通过以下两种方式创建。创建完成后返回对象引用并自动增加引用计数，当对象引用实例消亡时，自动减小引用计数。

- 函数调用返回对象引用。
- 直接调用 `put` 接口返回对象引用。

::::{tab-set}
:::{tab-item} Python

```python
import yr

yr.init()
object_ref = yr.put(123)
yr.finalize()
```

:::
:::{tab-item} Java

```java
import com.yuanrong.Config;
import com.yuanrong.api.YR;
import com.yuanrong.runtime.client.ObjectRef;
import com.yuanrong.exception.YRException;

public class Main {
    public static void main(String[] args) throws YRException {
        YR.init(new Config());
        ObjectRef objectRef = YR.put(123);
        YR.Finalize();
    }
}
```

:::

:::{tab-item} C++

```c++
#include <yr/yr.h>

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);
    YR::ObjectRef<int> objectRef = YR::Put<int>(123);
    YR::Finalize();
    return 0;
}
```

:::
::::

## 获取数据对象

使用 `get` 接口从对象引用中获取远程对象的结果。如果当前节点的 openYuanrong 数据系统中没有该对象，则自动从远端节点下载。`get` 接口支持并行获取多个对象结果，支持设置获取超时时间，超时时间内如果数据对象没有产生则抛出异常。

::::{tab-set}
:::{tab-item} Python

```python
import time
import yr

yr.init()

# Get the value of one object ref.
object_ref = yr.put(123)
assert yr.get(object_ref) == 123

# Get the values of multiple object refs in parallel.
assert yr.get([yr.put(i) for i in range(3)]) == [0, 1, 2]

# Get timeout example.
@yr.invoke
def long_running_function():
    time.sleep(10)

object_ref = long_running_function.invoke()
try:
    yr.get(object_ref, 3)
except TimeoutError as e:
    print(e)

yr.finalize()
```

:::
:::{tab-item} Java

```java
import com.yuanrong.Config;
import com.yuanrong.api.YR;
import com.yuanrong.runtime.client.ObjectRef;
import com.yuanrong.exception.YRException;
import com.yuanrong.storage.WaitResult;

import java.util.concurrent.TimeUnit;
import java.util.HashMap;
import java.util.ArrayList;
import java.util.List;

public class Main {
    // Get timeout example.
    public static class MyApp {
        public static int longRunningFunction() throws InterruptedException {
            TimeUnit.SECONDS.sleep(10);
            return 1;
        }
    }

    public static void main(String[] args) throws YRException {
        YR.init(new Config());

        // Get the value of one object ref.
        ObjectRef objectRef = YR.put(123);
        System.out.println(YR.get(objectRef, 30));

        // Get the values of multiple object refs in parallel.
        List<ObjectRef> objectRefs = new ArrayList<>();
        for (int i = 0; i < 3; i++) {
            objectRefs.add(YR.put(i));
        }

        WaitResult waitResult = YR.wait(objectRefs, objectRefs.size(), -1);
        for (ObjectRef ref : waitResult.getReady()) {
            System.out.println(YR.get(ref, 30));
        }

        try {
            ObjectRef resultRef = YR.function(MyApp::longRunningFunction).invoke();
            YR.get(resultRef, 3);
        } catch (YRException e) {
            System.out.println(e);
        }

        YR.Finalize();
    }
}
```

:::
:::{tab-item} C++

```c++
#include <iostream>
#include <yr/yr.h>

// Get timeout example.
int LongRunningFunction()
{
    std::this_thread::sleep_for(std::chrono::seconds(10));
    return 1;
}
YR_INVOKE(LongRunningFunction);


int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);

    // Get the value of one object ref.
    YR::ObjectRef<int> objectRef = YR::Put<int>(123);
    std::cout << *YR::Get(objectRef) << std::endl;

    // Get the values of multiple object refs in parallel.
    std::vector<YR::ObjectRef<int>> object_refs;
    for (int i = 0; i < 3; i++) {
        object_refs.emplace_back(YR::Put(i));
    }

    auto waitResults = YR::Wait(object_refs, object_refs.size());
    for (auto ref : waitResults.first) {
        std::cout << *YR::Get(ref) << std::endl;
    }

    YR::ObjectRef<int> object_ref = YR::Function(LongRunningFunction).Invoke();
    try {
        int result = *YR::Get(object_ref, 3);
    } catch (YR::Exception &e) {
        std::cout << e.what() << std::endl;
    }

    YR::Finalize();
    return 0;
}
```

:::
::::

## 传递数据对象

对象引用可以在应用程序之间自由传递，对象通过分布式引用计数自动进行生命周期管理。

- **将对象引用作为函数的顶层参数传递：** 将对象引用作为顶层参数传递时，openYuanrong 会自动调用 `get` 接口将数据对象的值获取出来传入被调函数，这意味着当对象数据完全可用时才会执行任务。

    ```python
    import yr

    yr.init()

    @yr.invoke
    def get_nums():
        return [1, 2, 3]

    @yr.invoke
    def dis_sum(args): # dis_sum 被调用时，传入的是值为 [1, 2, 3]
        return sum(args)

    ref = dis_sum.invoke(get_nums.invoke()) # 传递给 dis_sum 的参数是一个对象引用，在运行时 openYuanrong 将自动获取对象的值并传递
    num = yr.get(ref, 30)
    assert num == 6
    ```  

- **将对象引用作为嵌套参数传递：** 当数据对象嵌套在其他数据对象中传递时，在任务中需要调用 `get` 接口来获取具体的值。数据对象嵌套传递时，外层的对象引用会自动增加内层对象引用的引用计数，从而保证当外层对象引用存在时，内层对象引用一定不会被释放。

    ```python
    import yr

    yr.init()

    @yr.invoke
    def get_num(x):
        return x

    @yr.invoke
    def dis_sum(args): # dis_sum 被调用时，传入的值为 [objref1， objref2， objref3]
        return sum(yr.get(args)) # 调用 get 时，objref1/objref2/objref3 的值才会传输到当前机器。

    objref1 = get_num.invoke(1) # 此处返回 object ref，发出调用请求后即可返回，无需等待 get_num 执行完成
    objref2 = get_num.invoke(2)
    objref3 = get_num.invoke(3)
    objref = yr.put([objref1, objref2, objref3]) # 嵌套传递 object ref
    ref = dis_sum.invoke(objref)
    num = yr.get(ref, 30)
    assert num == 6
    ```

## 数据对象溢出到磁盘

数据对象数据存储在 openYuanrong 数据系统的共享内存中，当内存不足时，支持自动将数据溢出到磁盘并从内存中删除数据。当数据需要读取时，自动从磁盘中加载到共享内存。需要注意，当磁盘空间也不足时，数据对象会写入失败。

使用数据对象溢出功能，需要在部署时指定相关参数，默认为关闭。

```yaml
# 指定 object 溢出到磁盘的路径，如果配置为空，表示禁止使用 object 溢出功能。
spillDirectory: ""
```

数据对象溢出有以下参数，可用于设置磁盘空间上限、溢出的并发线程、文件大小等参数，用于性能调优。<br>

```yaml
# 指定溢出到磁盘时占用的最大磁盘空间，单位是字节。当配置为 0 时，openYuanrong 会采用openYuanrong 集群拉起时的磁盘剩余空间的 95% 作为最大可占用空间。
# 当 spillDirectory 配置目录是独占磁盘时，建议将该值配置为 0，由系统自动计算可用值。否则，建议按实际的磁盘空间规划配置该值。
spillSizeLimit: "0"
# 指定向磁盘写数据时的最大并行度，当磁盘性能非常高时，可尝试将该值调高提升性能。
spillThreadNum: 8
# 指定溢出到磁盘上的单个文件大小。单位为 MB，取值范围：200-10240。
# 当对象较小时，多个对象合并写入到一个文件中。当对象很大且超出该值时，每个文件一个对象。
spillFileMaxSizeMb: 200
# 用于指定对象溢出最大打开的文件句柄数。当该值被调小时，可能会降低性能。
spillFileOpenLimit: 512
# 是否允许文件系统的预读功能。当设置为 false 时，关闭文件系统的预读功能，可减少读放大，但可能会影响性能。
spillEnableReadahead: true
```

## 更多信息

openYuanrong 也提供了近计算 KV 缓存能力，基于共享内存实现免拷贝的 KV 数据读写。查看[KV 缓存开发指南](./KV.md)。
