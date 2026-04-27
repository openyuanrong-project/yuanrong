# ObjectRef

包名：`org.yuanrong.runtime.client`。

## public class ObjectRef

ObjectRef 类。

:::{Note}

1.openYuanrong 鼓励用户将大对象通过 YR.put 存入数据系统并获取一个唯一 ObjectRef （对象引用），在 invoke 调用用户函数时使用该 ObjectRef 替代对象本身作为函数入参，减小大对象在 openYuanrong 及用户函数各组件间传输产生的开销，保证流转高效。

2.每次调用用户函数的返回值也将以 ObjectRef 的形式返回，用户可以将该 ObjectRef 用于下一步调用的入参或者通过 YR.get 取回对应对象。

3.当前不支持用户自行构造 objectRef。

:::

### 接口说明

#### public ObjectRef(String objectID)

ObjectRef 的构造函数。

- 参数：

   - **objectID** - object 在 openYuanrong 集群内的 id。

#### public ObjectRef(String objectID, Class<?> returnType)

ObjectRef 的构造函数。

- 参数：

   - **objectID** - object 在 openYuanrong 集群内的 id。
   - **returnType** - object 类型。

#### public String getObjId()

获取 objectID。

- 返回：

    objectID: object 在 openYuanrong 集群内的 id。

#### public Object get(Class<?> classType) throws YRException, LibRuntimeException

获取结果。

- 参数：

   - **classType** - 类的类型。

- 返回：

    ObjectRef 的结果。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
   - **LibRuntimeException** - 数据系统错误。

#### public Object get() throws YRException, LibRuntimeException

获取结果。

- 返回：

    ObjectRef 的结果。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
   - **LibRuntimeException** - 数据系统错误。

#### public Object get(int timeout) throws YRException, LibRuntimeException

获取结果。

- 参数：

   - **timeout** - 超时时间，单位为秒。

- 返回：

    ObjectRef 的结果。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
   - **LibRuntimeException** - 数据系统错误。

### 私有成员

``` java

private Class<?> type
```

object 类型。

``` java

private boolean isByteBuffer = false
```

object 是否为 ByteBuffer 类型。

``` java

private final String objectID
```

object 在 openYuanrong 集群内的 id。
