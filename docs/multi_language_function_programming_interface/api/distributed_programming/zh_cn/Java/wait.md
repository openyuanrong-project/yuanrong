# wait

包名：`org.yuanrong.api`。

## 接口说明

### public static WaitResult wait(List<ObjectRef> refs, int waitNum, int timeoutSec) throws YRException

等待结果返回。

当返回的结果数量达到 `waitNum` 或等待时间超过 `timeoutSec` 时，此方法将返回。

```java

int y = 1;
// 并行获取多个对象引用的值。
List<ObjectRef> objectRefs = new ArrayList<>();
for (int i = 0; i < 3; i++) {
    objectRefs.add(YR.put(i));
}
WaitResult waitResult = YR.wait(objectRefs, /*num_returns=*/ 1, /*timeoutMs=*/ 1000);
System.out.println(waitResult.getReady()); // ready objects 的列表。
System.out.println(waitResult.getUnready()); // unready objects 的列表。
```

- 参数：

   - **refs** – ObjectRef 列表。
   - **waitNum** - 应返回的结果个数，取值要求大于 0。
   - **timeoutSec** – 等待的超时时间, 单位秒，取值要求大于等于 0 或者 等于 -1。

- 返回：

    WaitResult：用于存储返回结果。使用 getReady() 获取可检索的 ObjectRef 列表；使用 getUnready() 获取尚未可检索的 ObjectRef 列表。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
