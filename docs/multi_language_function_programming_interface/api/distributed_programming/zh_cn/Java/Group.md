# Group

包名：`org.yuanrong`。

## public class Group

类 Group 是 openYuanrong 成组实例调度类，用于完成成组实例的生命周期管理，包括成组实例的创建及销毁。

### 接口说明

#### public Group()

Group 的构造函数。

#### public Group(String name, GroupOptions opts)

Group 的构造函数。

- 参数：

   - **name** (String) - 实例成组调度类名称，名称不能重复。
   - **opts** (GroupOptions) - 实例成组调度生命周期参数。

#### public void invoke() throws YRException

成组调度实例 invoke 接口，成组实例创建。

```java

Group g = new Group();
g.setGroupName("groupName");
GroupOptions gOpts = new GroupOptions();
gOpts.setTimeout(60);
g.setGroupOpts(gOpts);
InvokeOptions opts = new InvokeOptions();
opts.setGroupName("groupName");
InstanceHandler res1 = YR.instance(MyClass::new).options(opts).invoke();
InstanceHandler res2 = YR.instance(MyClass::new).options(opts).invoke();
g.invoke();
```

:::{Note}

1、单个 group 最多成组创建 256 个实例。

2、并发创建最多支持 12 个 group ，每个 group 最多成组创建 256 个实例。

3、invoke() 接口在 NamedInstance::Export() 之后调用会导致当前线程卡住。

4、不调用 invoke() 接口，直接向有状态函数实例发起函数请求并 get 结果会导致当前线程卡住。

5、重复调用 invoke() 接口会导致异常抛出。

6、group 内实例不支持指定生命周期为 detached 。

:::

- 抛出：

   - **YRException** - 成组实例遵守 Fate-sharing 原则，当某个实例创建失败后，接口抛异常，全部实例创建失败。

#### public void terminate() throws YRException

成组调度实例 terminate 接口，删除成组实例创建。

```java

Group g = new Group();
g.setGroupName("groupName");
InvokeOptions opts = new InvokeOptions();
opts.setGroupName("groupName");
InstanceHandler res1 = YR.instance(MyClass::new).options(opts).invoke();
InstanceHandler res2 = YR.instance(MyClass::new).options(opts).invoke();
g.invoke();
g.terminate();
```

- 抛出：

   - **YRException** - 统一抛出的异常。

#### public void Wait() throws YRException

成组调度实例 Wait 接口，等待成组实例创建完成。

- 抛出：

   - **YRException** - 成组实例遵守 Fate-sharing 原则，当某个实例创建失败后，接口抛异常，全部实例创建失败。wait 超时后接口抛异常。

### 私有成员

``` java

private String groupName
```

实例成组调度类名称。

``` java

private GroupOptions groupOpts
```

实例成组调度生命周期参数。

## public class GroupOptions

类 GroupOptions 是 openYuanrong 成组实例调度类的一个参数，用于定义实例成组调度生命周期参数，包括内核资源不足时重调度超时时间等。

### 接口说明

#### public GroupOptions()

GroupOptions 的构造函数。

#### public GroupOptions(int timeout)

GroupOptions 的构造函数。

- 参数：

   - **timeout** (int) - openYuanrong 内核资源不足重调度超时时间。

#### public GroupOptions(int timeout, boolean sameLifecycle)

GroupOptions 的构造函数。

- 参数：

   - **timeout** (int) - openYuanrong 内核资源不足重调度超时时间。
   - **sameLifecycle** (boolean) - 是否为分组实例启用 fate-sharing 配置。

### 私有成员

``` java

private int timeout
```

openYuanrong 内核资源不足重调度超时时间，单位秒, ``-1`` 表示内核无限重试调度，其余情形赋值小于 ``0`` 抛异常。

``` java

private boolean sameLifecycle = true
```

是否为分组实例启用 fate-sharing 配置。``true`` 表示组中的实例将一起创建和销毁，``false`` 表示实例可以拥有独立的生命周期。默认为 ``true``。
