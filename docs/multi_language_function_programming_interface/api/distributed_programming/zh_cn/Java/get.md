# get

包名：`org.yuanrong.api`。

## 接口说明

### public static Object get(ObjectRef ref, int timeoutSec) throws YRException

获得存储在数据系统中的值。

此方法将阻塞 timeoutSec 秒，直至获取到结果。如果过了超时时间仍未获取到结果，则抛出异常。超时时间应大于 0，否则会立即抛出 YRException("timout is invalid, it needs greater than 0")。

```java

ObjectRef ref = YR.put(1);
System.out.println(YR.get(ref,3000));
```

- 参数：

   - **ref** – `invoke` 方法或者 `YR.put` 返回值，必填。
   - **timeoutSec** – 获取数据的超时时间，单位为秒。必填，取值要求大于 ``0``。

- 返回：

    Object 取得的对象。

- 抛出：

   - **YRException** - 取回数据出错，超时或数据系统服务不可用。
