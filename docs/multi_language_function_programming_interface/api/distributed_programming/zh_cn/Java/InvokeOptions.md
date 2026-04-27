# InvokeOptions

包名：`org.yuanrong`。

## public class InvokeOptions

函数调用选项，用于指定调用资源等功能。

### 接口说明

#### public InvokeOptions()

构造函数。

#### public void setCustomExtensions(Map<String, String> extension)

配置 InvokeOptions 中的 customExtensions。

```java

HashMap<String, String> testCustomExtensions = new HashMap<>();
testCustomExtensions.put("test-extension", "true");
testCustomExtensions.put(Constants.POST_START_EXEC, "false");
InvokeOptions options = new InvokeOptions();
options.setCustomExtensions(testCustomExtensions);
```

- 参数：

   - **extension** (Map<String, String>) – 用户自定义的配置。

#### public void addCustomExtensions(String key, String value)

向 InvokeOptions 中 customExtensions 增加新的字段。

```java

InvokeOptions options = new InvokeOptions();
options.addCustomExtensions("test-extension", "true");
```

- 参数：

   - **key** (String) – 用户自定义配置的键。
   - **value** (String) – 用户自定义的配置的值。

#### public static Builder builder()

创建 InvokeOptions 的 Builder。

- 返回：

    用于构建 InvokeOptions 实例的构建器。

### Builder 接口

以下是用来构建 InvokeOptions 实例的 Builder 接口。

```java
String expectedMapKey = "gpu-time";
Float expectedMapValue = 1.5f;
int expectedSize = 2;
HashMap<String, Float> customResources = new HashMap<>();
customResources.put("accelerator", 0.5f);
String expectedMapKey = "test-extension";
String expectedMapValue = "true";
int priority = 22;
HashMap<String, String> testCustomExtensions = new HashMap<>();
testCustomExtensions.put(expectedMapKey, expectedMapValue);
testCustomExtensions.put(Constants.POST_START_EXEC, "false");
LabelOperator labelOperator = new LabelOperator(OperatorType.LABEL_IN, "test-label");
ArrayList<LabelOperator> testOperatorsList = new ArrayList<>();
testOperatorsList.add(labelOperator);
Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, testOperatorsList);
ArrayList<Affinity> testAffinityList = new ArrayList<>();
testAffinityList.add(affinity);
boolean preemptedAllowed = true;
int instancePriority = 22;
long scheduleTimeoutMs = 20000L;

InvokeOptions options = new InvokeOptions.builder().customExtensions(testCustomExtensions)
            .addCustomExtensions(expectedMapKey, expectedMapValue)
            .addCustomExtensions(Constants.POST_START_EXEC, "true")
            .customResources(customResources)
            .addCustomResource(expectedMapKey,expectedMapValue)
            .priority(priority)
            .scheduleAffinity(testAffinityList)
            .addInstanceAffinity(AffinityType.PREFERRED, testOperatorsList)
            .addResourceAffinity(AffinityType.PREFERRED, testOperatorsList)
            .preemptedAllowed(preemptedAllowed)
            .instancePriority(instancePriority)
            .scheduleTimeoutMs(scheduleTimeoutMs)
            .build();
```

| 接口                  | 参数                                              | 含义                                          | 样例                                                              |
|---------------------|-------------------------------------------------|---------------------------------------------|-----------------------------------------------------------------|
| cpu                 | int cpu                                         | 配置 InvokeOptions 中 cpu 的参数大小。                | InvokeOptions.builder().cpu(400).build()                        |
| memory              | int memory                                      | 配置 InvokeOptions 中 memory 的参数大小。             | InvokeOptions.builder().memory(400).build()                     |
| customResources     | Map<String, Float> customResources              | 配置 InvokeOptions 中 customResources。          | InvokeOptions.builder().customResources(customResources).build() |
| addCustomResource   | String key, Float val                           | 向 InvokeOptions 中 customResources 增加新的资源配置。  | 参考样例代码                                                          |
| customExtensions    | Map<String, String> customExtensions            | 配置 InvokeOptions 中 customExtensions。         | InvokeOptions.builder().customExtensions(customExtensions).build() |
| addCustomExtensions  | String key, String val                          | 向 InvokeOptions 中 customExtensions 增加新的字段。   | 参考样例代码                                                          |
| recoverRetryTimes   | int recoverRetryTimes                           | 配置 InvokeOptions 中 recoverRetryTimes 的参数大小。  | InvokeOptions.builder().recoverRetryTimes(3).build()            |
| podLabels           | Map<String, String> podLabels                   | 配置 InvokeOptions 中 podLabels 参数。**仅在 Kubernetes 上部署时配置有效**。             | InvokeOptions.builder().podLabels(podLabels).build()            |
| addPodLabel         | String key, String val                          | 向 InvokeOptions 中 podLabels 增加新的标签。**仅在 Kubernetes 上部署时配置有效**。          | InvokeOptions.builder().addPodLabel(key, val).build()           |
| addLabel            | String val                                      | 向 InvokeOptions 中 labels 增加新的标签。             | InvokeOptions.builder().addLabel(val).build()                   |
| scheduleAffinity    | List<Affinity> affinities                       | 配置 InvokeOptions 中 scheduleAffinity 参数。      | InvokeOptions.builder().scheduleAffinity(affinities).build()    |
| addScheduleAffinity | Affinity affinity                               | 向 InvokeOptions 中 scheduleAffinity 增加新的亲和配置。 | 参考样例代码                                                          |
| addInstanceAffinity | AffinityType type, List<LabelOperator> operators| 向 InvokeOptions 中 scheduleAffinity 增加实例亲和配置。 | 参考样例代码                                                          |
| addResourceAffinity | AffinityType type, List<LabelOperator> operators| 向 InvokeOptions 中 scheduleAffinity 增加资源亲和配置。 | 参考样例代码                                                          |
| preferredPriority   | boolean val                                     | 配置 InvokeOptions 中 preferredPriority。        | InvokeOptions.builder().preferredPriority(false).build()        |
| requiredPriority    | boolean val                                     | 配置 InvokeOptions 中 requiredPriority。         | InvokeOptions.builder().requiredPriority(false).build()         |
| preferredAntiOtherLabels | boolean val                                | 配置 InvokeOptions 中 preferredAntiOtherLabels。 | InvokeOptions.builder().preferredAntiOtherLabels(false).build() |
| priority            | int val                                         | 配置 InvokeOptions 中 priority。                 | InvokeOptions.builder().priority(22).build()                    |
| needOrder           | boolean val                                     | 配置 InvokeOptions 中 needOrder。                | InvokeOptions.builder().needOrder(false).build()                |
| groupName           | String val                                      | 配置 InvokeOptions 中 groupName。                | InvokeOptions.builder().groupName(group).build()                |
| preemptedAllowed    | boolean val                                     | 配置 InvokeOptions 中 preemptedAllowed。         | InvokeOptions.builder().preemptedAllowed(true).build()          |
| instancePriority    | int val                                         | 配置 InvokeOptions 中 instancePriority。         | InvokeOptions.builder().instancePriority(22).build()            |
| scheduleTimeoutMs   | long val                                        | 配置 InvokeOptions 中 scheduleTimeoutMs。        | InvokeOptions.builder().scheduleTimeoutMs(20000L).build()       |

### 私有成员

``` java

private int cpu = 500
```

指定 cpu core 资源，默认与 service.yaml 配置相同，单位为 ``1/1000 cpu core`` , 取值 ``[300, 16000]``。

``` java

private int memory = 500
```

指定内存资源，默认与 service.yaml 配置相同，单位是 ``MB``，取值 ``[128, 65536]``，默认 ``500``。

``` java

private Map<String, Float> customResources
```

指定用户自定义资源，如 gpu 等。

``` java

private Map<String, String> customExtensions
```

指定用户自定义的配置，如函数并发度。同时也可以作为 metrics 的用户自定义 tag，用于采集用户信息。

``` java

private int recoverRetryTimes = 0
```

实例最大恢复次数（当实例异常退出时，自动用当前最新状态恢复实例）。默认 ``0``，当实例异常退出时不进行自动恢复（并不保证所有指定的重试次数都会完成，存在不可恢复的异常场景，openYuanrong 会提前停止重试并抛出异常）。

``` java

private Map<String, String> podLabels
```

pod 需要打上的标签，低可靠实例不支持配置。仅在 Kubernetes 上部署时配置有效。

``` java

private List<String> labels
```

实例需要打上的标签，用于实例亲和。

``` java

private List<Affinity> scheduleAffinities
```

亲和配置，具体使用参考[亲和](Affinity.md)章节。

``` java

private boolean preferredPriority = true
```

是否开启弱亲和优先级调度，开启后，当传入多个弱亲和条件，按顺序匹配打分，出现一个满足的即可调度成功，默认 ``true``。

``` java

private boolean requiredPriority = false
```

是否开启强亲和优先级调度，开启后，当传入多个强亲和条件，按顺序匹配打分，当传入多个强亲和条件都不满足时，调度失败， 默认 ``false``。

``` java

private boolean preferredAntiOtherLabels = true
```

是否开启反亲和不可选资源，开启后，当传入多个弱亲和条件都不满足时，调度失败，默认 ``true``。

``` java

private int priority = 0
```

定义无状态函数优先级，默认值为 ``0``。

``` java

private boolean needOrder = false
```

设置实例请求是否有序，默认 ``false``。仅在并发度为 1 时生效。

``` java

private String groupName
```

指定成组实例调度器 name ，默认为空。

``` java

private boolean preemptedAllowed = false
```

实例是否可以被抢占，仅在优先级场景下（openYuanrong 部署的 maxPriority 配置项大于 0 的场景）生效。默认 ``false``。

``` java

private int instancePriority = 0
```

实例的优先级，数值越大优先级越高，高优先级的实例可以抢占低优先级且被配置为（preemptedAllowed = true）的实例。仅在优先级场景下（openYuanrong 部署的 maxPriority 配置项大于 0 的场景）生效。instancePriority 的最小值为 0，最大值为 openYuanrong 部署的 maxPriority 配置。默认值为 ``0``。

``` java

private long scheduleTimeoutMs = 30000L
```

实例的调度超时时间，单位毫秒，取值 ``[-1, Long.MAX_VALUE]``。默认值为 ``30000L``。

``` java

private int retryTimes = 0
```

无状态函数或有状态函数的方法调用重试次数。

``` java

private Map<String, String> envVars
```

设置函数启动时的环境变量。

``` java

private String traceId
```

函数调用的 traceId，用于链路追踪。

#### 关于 customExtensions

customExtensions 表示用户自定义的配置，接收用户自定义的 key / value 键值对，常见的配置如下：

| **key，String 类型**        | **value，String 类型** | **描述**                                                                                                                                 |
|--------------------------|---------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| Concurrency              | 并发度                 | 实例并发度。限制：``[1,1000]``。                                                                                                                     |
| lifecycle                | detached            | 支持 detached 模式。                                                                                                                        |
| init_call_timeout   | init 调用超时时间                |  init 调用超时时间，如果未配置，将以 openYuanrong 部署时配置时间为准                                                       |
| ReliabilityType   | 实例可靠性配置                  | 配置 low 时，将减少实例数据持久化次数，不支持 Recover，不支持配置 pod Labels                                                                                                         |
| DELEGATE_DIRECTORY_INFO  | 自定义目录               | 支持创建、删除子目录能力。当实例创建时，若用户自定义目录存在且有读写权限，则在该目录下创建子目录作为工作目录；否则，在/tmp 目录下创建子目录作为工作目录。当实例销毁时，工作目录销毁。其中，用户函数可以通过 INSTANCE_WORK_DIR 环境变量获取工作目录。 |
| DELEGATE_DIRECTORY_QUOTA | 子目录配额大小             | 取值范围大于 0 M 小于 1 TB, 如果没有该配置，默认 ``512 M``, 配置 -1 则不监控, 单位：``MB``。                                                                                                 |
| GRACEFUL_SHUTDOWN_TIME   | 自定义优雅退出时间           | 实例优雅退出超时时间，单位秒。 限制：>=0 , 0 表示立即退出，不保证能够完成用户的优雅退出函数；若配置 <0 ，则使用部署时的系统配置作为超时时间                                                           |
| RECOVER_RETRY_TIMEOUT    | 自定义 recover 超时时间               | 实例 recover 超时时间，单位毫秒。限制：>0, 默认值 ``10 * 60 * 1000``。                                                           |
| AFFINITY_POOL_ID   | 自定义亲和 pool 池 ID                | 对于实例创建，当遇到资源不足（或者亲和条件不满足）时，内核将创建指定 poolID 的 POD，用于实例调度。**仅在 Kubernetes 上部署时配置有效**。                                                           |

作为 metrics 的用户自定义 label 时：

```java
InvokeOptions opt = new InvokeOptions();
opt.addCustomExtensions("YR_Metrics", "{\"endpoint\":\"127.0.0.1\", \"project_id\":\"my_project_id\"}")
```

键为 YR_Metrics，值为自定义 label 的 json 字符串

Prometheus 中选择 `metrics name` 为 `yr_app_instance_billing_invoke_latency`，可在采集到的 invoke 信息中找到自定义 tag 信息：

```shell
yr_app_instance_billing_invoke_latency{
...
endpoint="127.0.0.1",
...}
```

#### 关于 podLabels

podLabels 为 InvokeOptions 中的一个 map 类型字段。当 create 一个函数实例时，podLabels 可以接收来自用户的 key、value 键值对，并将其传给内核。ActorPattern 函数实例特化完成（Running）后，Scaler 给 POD 打上传入的标签；ActorPattern 函数实例故障/删除时，Scaler 将 POD 对应标签置为空（移除）。

:::{Note}

- 仅在 Kubernetes 上部署时配置有效。
- podLabels 中可以存放的标签不能超过 5 个。
- podLabels 中 key 和 value 的约束如下表:

| 约束对象 | 类型   | 说明                                           |
| -------- | ------ |----------------------------------------------|
| key      | String | 支持大小写字母、数字以及中划线，允许长度 1-63，不以中划线开头或结尾；不允许空字符串 |
| value    | String | 支持大小写字母、数字以及中划线，允许长度 1-63，不以中划线开头或结尾；允许空字符串  |

- 当用户传入的 PodLabels 不满足约束条件时，会抛出对应异常及报错信息。

:::

```java
public static class Counter {
    private int value = 0;

    public int increment() {
        this.value += 1;
        return this.value;
    }

    public static int functionWithAnArgument(int value) {
        return value + 1;
    }
}

// Instance example
InvokeOptions options= new InvokeOptions();
options.setPodLabels(new HashMap<String, String>() {
    {
        put("k1", "v1");
    }
});
InstanceHandler counter = YR.instance(Counter::new).options(options).invoke();
ObjectRef objectRef = counter.function(Counter::increment).invoke();
System.out.println(YR.get(objectRef, 3000));

// Function example
public static class MyYRApp {
    public static int myFunction() {
        return 1;
    }

    public static int functionWithAnArgument(int value) {
        return value + 1;
    }
}

ObjectRef res = YR.function(MyYRApp::myFunction).options(options).invoke();
System.out.println(YR.get(res, 3000));
ObjectRef objRef2 = YR.function(MyYRApp::functionWithAnArgument).invoke(1);
System.out.println(YR.get(objRef2, 3000));
YR.Finalize();
```
