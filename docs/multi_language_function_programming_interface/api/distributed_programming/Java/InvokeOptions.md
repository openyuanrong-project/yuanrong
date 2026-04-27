# InvokeOptions

package: `org.yuanrong`。

## public class InvokeOptions

Function call options, used to specify call resources and other functions.

### Interface description

#### public InvokeOptions()

Constructor.

#### public void setCustomExtensions(Map<String, String> extension)

Configure `customExtensions` in `InvokeOptions`.

```java

HashMap<String, String> testCustomExtensions = new HashMap<>();
testCustomExtensions.put("test-extension", "true");
testCustomExtensions.put(Constants.POST_START_EXEC, "false");
InvokeOptions options = new InvokeOptions();
options.setCustomExtensions(testCustomExtensions);
```

- Parameters:

   - **extension** (Map<String, String>) – User-defined configurations.

#### public void addCustomExtensions(String key, String value)

Add a new field to `customExtensions` in `InvokeOptions`.

```java

InvokeOptions options = new InvokeOptions();
options.addCustomExtensions("test-extension", "true");
```

- Parameters:

   - **key** (String) – The key for user-defined configuration.
   - **value** (String) – The value for user-defined configuration.

#### public static Builder builder()

Create a Builder for `InvokeOptions`.

- Returns:

    A builder used to construct instances of `InvokeOptions`.

### Builder Interface

Below is the Builder interface used to construct instances of `InvokeOptions`.

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

| Interface                | Parameters                                       | Meaning                                                                  | Example                                                            |
| ------------------------ | ------------------------------------------------ | ------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| cpu                      | int cpu                                          | Configure the cpu parameter size in InvokeOptions.                        | InvokeOptions.builder().cpu(400).build()                           |
| memory                   | int memory                                       | Configure the memory parameter size in InvokeOptions.                     | InvokeOptions.builder().memory(400).build()                        |
| customResources          | Map\<String, Float> customResources              | Configure customResources in InvokeOptions.                               | InvokeOptions.builder().customResources(customResources).build()   |
| addCustomResource        | String key, Float val                            | Add a new resource configuration to customResources in InvokeOptions.     | See example code                                                   |
| customExtensions         | Map\<String, String> customExtensions            | Configure customExtensions in InvokeOptions.                              | InvokeOptions.builder().customExtensions(customExtensions).build() |
| addCustomExtensions       | String key, String val                           | Add a new field to customExtensions in InvokeOptions.                     | See example code                                                   |
| recoverRetryTimes        | int recoverRetryTimes                            | Configure the recoverRetryTimes parameter size in InvokeOptions.          | InvokeOptions.builder().recoverRetryTimes(3).build()               |
| podLabels                | Map\<String, String> podLabels                   | Configure podLabels parameter in InvokeOptions. **effective only when deployed on Kubernetes**.              | InvokeOptions.builder().podLabels(podLabels).build()               |
| addPodLabel              | String key, String val                           | Add a new label to podLabels in InvokeOptions. **effective only when deployed on Kubernetes**.              | InvokeOptions.builder().addPodLabel(key, val).build()              |
| addLabel                 | String val                                       | Add a new label to labels in InvokeOptions.                               | InvokeOptions.builder().addLabel(val).build()                      |
| scheduleAffinity         | List<Affinity> affinities                        | Configure scheduleAffinity parameter in InvokeOptions.                    | InvokeOptions.builder().scheduleAffinity(affinities).build()       |
| addScheduleAffinity      | Affinity affinity                                | Add a new affinity configuration to scheduleAffinity in InvokeOptions.    | See example code                                                   |
| addInstanceAffinity      | AffinityType type, List<LabelOperator> operators | Add instance affinity configuration to scheduleAffinity in InvokeOptions. | See example code                                                   |
| addResourceAffinity      | AffinityType type, List<LabelOperator> operators | Add resource affinity configuration to scheduleAffinity in InvokeOptions. | See example code                                                   |
| preferredPriority        | boolean val                                      | Configure preferredPriority in InvokeOptions.                             | InvokeOptions.builder().preferredPriority(false).build()           |
| requiredPriority         | boolean val                                      | Configure requiredPriority in InvokeOptions.                              | InvokeOptions.builder().requiredPriority(false).build()            |
| preferredAntiOtherLabels | boolean val                                      | Configure preferredAntiOtherLabels in InvokeOptions.                      | InvokeOptions.builder().preferredAntiOtherLabels(false).build()    |
| priority                 | int val                                          | Configure priority in InvokeOptions.                                      | InvokeOptions.builder().priority(22).build()                       |
| needOrder                | boolean val                                      | Configure needOrder in InvokeOptions.                                     | InvokeOptions.builder().needOrder(false).build()                   |
| groupName                | String val                                       | Configure groupName in InvokeOptions.                                     | InvokeOptions.builder().groupName(group).build()                   |
| preemptedAllowed         | boolean val                                      | Configure preemptedAllowed in InvokeOptions.                              | InvokeOptions.builder().preemptedAllowed(true).build()             |
| instancePriority         | int val                                          | Configure instancePriority in InvokeOptions.                              | InvokeOptions.builder().instancePriority(22).build()               |
| scheduleTimeoutMs        | long val                                         | Configure scheduleTimeoutMs in InvokeOptions.                             | InvokeOptions.builder().scheduleTimeoutMs(20000L).build()          |

### Private Members

``` java

private int cpu = 500
```

Specify the CPU core resources, which default to the same configuration as in `service.yaml`, in units of ``1/1000 CPU core``, with a value range of ``[300, 16000]``.

``` java

private int memory = 500
```

Specify the memory resources, which default to the same configuration as in `service.yaml`, in units of ``MB``, with a value range of ``[128, 65536]``, and the default value is ``500``.

``` java

private Map<String, Float> customResources
```

Specify user-defined resources, such as GPUs.

``` java

private Map<String, String> customExtensions
```

Specify user-defined configurations, such as function concurrency. It can also be used as a user-defined tag for metrics to collect user-related information.

``` java

private int recoverRetryTimes = 0
```

The maximum number of instance recoveries (when an instance exits abnormally, it is automatically restored with the current latest state). The default value is ``0``, meaning no automatic recovery is performed when the instance exits abnormally (it is not guaranteed that all specified retries will be completed; there are scenarios with irrecoverable exceptions where openYuanrong will stop retrying early and throw an exception).

``` java

private Map<String, String> podLabels
```

Labels that need to be applied to the pod, not supported for low-reliability instances. effective only when deployed on Kubernetes.

``` java

private List<String> labels
```

Labels that need to be applied to the instance, used for instance affinity.

``` java

private List<Affinity> scheduleAffinities
```

Specify the affinity configuration. For specific usage, refer to the [Affinity](Affinity.md) section.

``` java

private boolean preferredPriority = true
```

Whether to enable weak affinity priority scheduling. When enabled, if multiple weak affinity conditions are provided, they will be matched and scored in order. Scheduling will succeed if any one condition is met. The default is ``true``.

``` java

private boolean requiredPriority = false
```

Whether to enable strong affinity priority scheduling. When enabled, if multiple strong affinity conditions are provided, they will be matched and scored in order. Scheduling will fail if none of the conditions are met. The default is ``false``.

``` java

private boolean preferredAntiOtherLabels = true
```

Whether to enable anti-affinity for non-selectable resources. When enabled, scheduling will fail if none of the weak affinity conditions are met. The default is ``true``.

``` java

private int priority = 0
```

Define the priority of stateless functions, with a default value of ``0``.

``` java

private boolean needOrder = false
```

Set whether the instance request is ordered, defaulting to ``false``. It only takes effect when the concurrency is 1.

``` java

private String groupName
```

Specify the name of the group instance scheduler, defaulting to empty.

``` java

private boolean preemptedAllowed = false
```

Whether the instance can be preempted, only effective in priority scenarios (where the maxPriority configuration item of openYuanrong deployment is greater than 0). The default is ``false``.

``` java

private int instancePriority = 0
```

The priority of the instance, where a higher number indicates a higher priority. Instances with higher priority can preempt instances with lower priority that are configured with (`preemptedAllowed = true`). This is only effective in priority scenarios (where the maxPriority configuration item of openYuanrong deployment is greater than 0). The minimum value for `instancePriority` is 0, and the maximum value is the maxPriority configuration of openYuanrong deployment. The default value is ``0``.

``` java

private long scheduleTimeoutMs = 30000L
```

The scheduling timeout for the instance, in milliseconds, with a value range of ``[-1, Long.MAX_VALUE]``. The default value is ``30000L``.

``` java

private int retryTimes = 0
```

The number of retry attempts for method calls of stateless functions or stateful functions.

``` java

private Map<String, String> envVars
```

Set environment variables when the instance starts.

``` java

private String traceId
```

Set the traceId for function calls for link tracing.

#### About customExtensions

`customExtensions` represents user-defined configurations, accepting user-defined key/value pairs. Common configurations include the following:

| **key, String type**       | **value, String type**             | **Description**                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| -------------------------- | ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Concurrency                | Concurrency level                  | The concurrency level of the instance. Limit: `[1,1000]`.                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| lifecycle                  | detached                           | Supports detached mode.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| init\_call\_timeout        | Init call timeout                  | The timeout for the init call. If not configured, the time configured during openYuanrong deployment will be used.                                                                                                                                                                                                                                                                                                                                                                                                 |
| ReliabilityType            | Instance reliability configuration | Configures the reliability type. When set to `low`, it reduces the number of instance data persistence operations, does not support `Recover`, and does not support configuring pod labels.                                                                                                                                                                                                                                                                                                                    |
| DELEGATE\_DIRECTORY\_INFO  | Custom directory                   | Supports the ability to create and delete subdirectories. When an instance is created, if the user-defined directory exists and has read/write permissions, a subdirectory will be created under it as the working directory; otherwise, a subdirectory will be created under the `/tmp` directory as the working directory. When the instance is destroyed, the working directory is also destroyed. The user function can obtain the working directory through the `INSTANCE_WORK_DIR` environment variable. |
| DELEGATE\_DIRECTORY\_QUOTA | Subdirectory quota size            | The value range is greater than 0 M and less than 1 TB. If not configured, the default is `512 M`. A value of `-1` means no monitoring. Unit: `MB`.                                                                                                                                                                                                                                                                                                                                                            |
| GRACEFUL\_SHUTDOWN\_TIME   | Custom graceful shutdown time      | The graceful shutdown timeout for the instance, in seconds. Limit: >=0. A value of `0` indicates immediate shutdown without guaranteeing the completion of the user's graceful shutdown function. If configured to be less than `0`, the system configuration at deployment time will be used as the timeout.                                                                                                                                                                                                  |
| RECOVER\_RETRY\_TIMEOUT    | Custom recover timeout             | The recover timeout for the instance, in milliseconds. Limit: >0. The default value is `10 * 60 * 1000`.                                                                                                                                                                                                                                                                                                                                                                                                       |
| AFFINITY\_POOL\_ID         | Custom affinity pool ID            | For instance creation, when encountering insufficient resources (or unsatisfied affinity conditions), the kernel will create a POD with the specified pool ID for instance scheduling. **effective only when deployed on Kubernetes**.   |

When used as a user-defined label for metrics:

```java
InvokeOptions opt = new InvokeOptions();
opt.addCustomExtensions("YR_Metrics", "{\"endpoint\":\"127.0.0.1\", \"project_id\":\"my_project_id\"}")
```

The key is `YR_Metrics`, and the value is the JSON string of the custom label.

In Prometheus, select `metrics name` as `yr_app_instance_billing_invoke_latency`, and you can find the custom tag information in the collected invoke information.

```shell
yr_app_instance_billing_invoke_latency{
...
endpoint="127.0.0.1",
...}
```

#### About podLabels

`podLabels` is a map - type field in `InvokeOptions`. When creating a function instance, `podLabels` can receive key - value pairs from the user and pass them to the kernel. After an `ActorPattern` function instance is specialized (Running), the Scaler applies the input labels to the POD. When an `ActorPattern` function instance fails or is deleted, the Scaler sets the corresponding POD labels to empty (removes them).

:::{Note}

- Effective only when deployed on Kubernetes.
- No more than 5 labels can be stored in podLabels.
- The constraints for keys and values in podLabels are as follows:

| Constraint Object | Type   | Description                                                                                                                                                           |
| ----------------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| key               | String | Supports uppercase and lowercase letters, digits, and hyphens. Allowed length is 1-63 characters. Must not start or end with a hyphen. Empty strings are not allowed. |
| value             | String | Supports uppercase and lowercase letters, digits, and hyphens. Allowed length is 1-63 characters          |

- When the PodLabels provided by the user do not meet the constraints, corresponding exceptions and error messages will be thrown.

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
