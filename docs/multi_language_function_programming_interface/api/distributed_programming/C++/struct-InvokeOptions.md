# struct InvokeOptions

## InvokeOptions Definition

```cpp
struct InvokeOptions {
    // in 1/1000 cpu core
    int cpu;

    // in 1MB
    int memory;

    // custom resources that other than cpu and memory
    std::unordered_map<std::string, float> customResources;

    // custom extentions
    std::unordered_map<std::string, std::string> customExtensions;

    // pod labels, effective only when deployed on Kubernetes
    std::unordered_map<std::string, std::string> podLabels;

    std::vector<std::string> labels;

    std::unordered_map<std::string, std::string> affinity;

    size_t retryTimes = 0;

    bool (*retryChecker)(const Exception &e) noexcept = nullptr;

    size_t priority = 0;

    bool alwaysLocalMode = false;

    bool preferredPriority = true;

    bool preferredAntiOtherLabels = true;

    bool requiredPriority = false;

    std::string groupName = "";

    InstanceRange instanceRange;

    int recoverRetryTimes = 0;
};
```

## InvokeOptions Specification

Function invocation options, used to specify calling resources and related configurations.

## InvokeOptions Parameters

| Name                     | Type                                      | Description                                                                                                                                                     | Default             |
|--------------------------|-------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------|
| cpu                      | int                                       | Specifies CPU core resources in `1/1000` CPU core units.                                                                                                        | ``500``                 |
| memory                   | int                                       | Specifies memory resources in `1MB` units.                                                                                                                      | ``500``                 |
| customResources          | std::unordered_map<std::string, float>    | Custom resources:<br>• Disk resources: "disk" (unit: GB).<br>• Heterogeneous resources: "GPU/XX/YY" and "NPU/XX/YY".<br>  - XX: card model (e.g. Ascend910B4, supports regex).<br>  - YY: count/HBM/latency/stream. | -                   |
| customExtensions         | std::unordered_map<std::string, std::string> | User-defined configurations (e.g. function concurrency).<br>Also used as custom tags for metrics collection.                                                     | -                   |
| podLabels                | std::unordered_map<std::string, std::string> | Specifies pod labels for instance placement(not supported for low-reliability instances). **Effective only when deployed on Kubernetes**.                         | -                   |
| labels                   | std::vector&lt;std::string&gt;                  | Labels for function instance affinity/anti-affinity(format: `"key"` or `"key:value"`).                                                                     | -                   |
| affinity                 | std::unordered_map<std::string, std::string> | Defines scheduling affinity/anti-affinity(deprecated).                                                                                                      | -                   |
| retryTimes               | size_t                                    | Retry count for stateless functions.                                                                                                                            | ``0``                 |
| retryChecker             | bool (*)(const YR::Exception &e) noexcept | Retry decision hook for stateless functions(empty by default, invalid when `retryTimes = 0`).                                                               | -                   |
| priority                 | size_t                                    | Defines priority for stateless functions.                                                                                                                       | ``0``                 |
| alwaysLocalMode          | bool                                      | Forces cluster mode to run locally with multi-threading(no effect in `local mode`).                                                                            | ``false``             |
| preferredPriority        | bool                                      | Priority scheduling. Matches weak affinity conditions sequentially, Succeeds when any condition is met.                                                | ``true``              |
| requiredPriority         | bool                                      | Priority scheduling. Requires all strong affinity conditions, Fails if any condition not met.                                                         | ``false``             |
| preferredAntiOtherLabels | bool                                      | Anti-affinity for non-selectable resources, Fails if no weak affinity condition met.                                                                       | ``true``              |
| groupName                | string                                    | Group instance scheduler name(empty by default).                                                                                                             | ``""``                |
| instanceRange            | InstanceRange                            | Function instance count range configuration, See [struct-InstanceRange](./struct-InstanceRange.md).                                                          | -                   |
| recoverRetryTimes        | int                                       | Maximum instance recovery attempts:<br>• Default ``0`` means no auto-recovery.<br>• May stop early for unrecoverable errors.                                          | ``0``                 |

### About retryTimes and retryChecker

For stateless functions, the following framework-retried error codes do not consume retry attempts:

1. ERR_RESOURCE_NOT_ENOUGH
2. ERR_INSTANCE_NOT_FOUND
3. ERR_INSTANCE_EXITED

Recommended user-decided retry for these error codes:

1. ERR_USER_FUNCTION_EXCEPTION
2. ERR_REQUEST_BETWEEN_RUNTIME_BUS
3. ERR_INNER_COMMUNICATION
4. ERR_SHARED_MEMORY_LIMIT
5. EDERR_OPERATE_DISK_FAILED
6. ERR_INSUFFICIENT_DISK_SPACE

`retryTimes` and `retryChecker` currently do not support Stateful functions, attempting to use them will throw exceptions. See example usage in [FunctionHandler::Options](./FunctionHandler-Options.md).

### About alwaysLocalMode

- When calling stateless functions in cluster mode with `alwaysLocalMode` set to ``true``, the generated ObjectRef will also be local, which can be checked using the `bool ObjectRef::IsLocal() const` method.

- When creating stateful functions in cluster mode with `alwaysLocalMode` set to ``true``, it is already determined whether the stateful function will be created locally with multithreading, and subsequent calls do not need to specify whether to execute locally.

Currently, mixing local and cluster ObjectRef is not supported for `Wait`, `Get`, `Cancel` interfaces or when passing `Invoke` parameters - doing so will throw an exception. An exception will also be thrown if the thread pool is empty during cluster mode initialization but still attempts to call `alwaysLocalMode` functions (whether stateless or stateful).

### About customExtensions

customExtensions represents user-defined configurations, accepting custom key/value pairs. Common configurations include:

| Key (std::string)       | Value (std::string)        | Description                                                                 | Default           |
|-------------------------|----------------------------|-----------------------------------------------------------------------------|-------------------|
| Concurrency             | Concurrency level          | Instance concurrency level<br>• Range: `[1,1000]`                          | ``1``              |
| lifecycle               | detached                   | Supports detached mode                                                      | -                 |
| init_call_timeout       | Init call timeout          | Timeout for init calls<br>• Uses deployment defaults if unset<br>See [runtime config](process-deploy-params-runtime) (`--runtime_init_call_timeout_seconds`) | -                 |
| ReliabilityType         | Instance reliability       | When configured as low, it will reduce the frequency of instance data persistence, does not support Recover, and does not support configuring pod Labels. | -                 |
| DELEGATE_DIRECTORY_INFO | Custom directory           | Supports creating and deleting subdirectories. When an instance is created:<br>• If the user-defined directory exists and has read/write permissions, a subdirectory is created under it as the working directory.<br>• Otherwise, a subdirectory is created under ``/tmp`` as the working directory.<br><br>The user function can access the working directory via the `INSTANCE_WORK_DIR` environment variable. | ``'/tmp'``         |
| DELEGATE_DIRECTORY_QUOTA| Subdirectory quota (MB)    | Range: 0 MB < value < 1 TB. Default: 512 MB. ``-1`` disables monitoring. Unit: MB. | ``512``            |
| GRACEFUL_SHUTDOWN_TIME  | Shutdown timeout (seconds) | Constraints: >=0<br>• 0: Immediate exit (no graceful guarantee)<br>• <0: Use system default (`--kill_process_timeout_seconds` in [runtime config](process-deploy-params-runtime) | -                 |
| RECOVER_RETRY_TIMEOUT   | Recover timeout (ms)       | Instance recover timeout, unit: milliseconds (ms). Constraint: Must be greater than 0 (>0).       | ``10 * 60 * 1000`` |
| AFFINITY_POOL_ID        | Affinity pool ID           | For instance creation, when resources are insufficient (or affinity conditions are not met),<br>the kernel will create a POD with the specified poolID for instance scheduling. **Effective only when deployed on Kubernetes**. | -                 |

Specify user-defined `metrics tag` in customExtensions：

```cpp
YR::InvokeOptions opt;
opt.customExtensions.insert({"YR_Metrics", "{\"endpoint\":\"127.0.0.1\", \"project_id\":\"my_project_id\"}"});
```

When the key is `YR_Metrics` and the value is a JSON string of custom labels:

In Prometheus, if you select `yr_app_instance_billing_invoke_latency` as the `metrics name`, the custom tag information can be found in the collected invoke metrics:

```shell
yr_app_instance_billing_invoke_latency{
...
endpoint="127.0.0.1",
...}
```

### About customResources

#### Configuring Disk Resources

Disk resources configured in `customResources` use GB as the unified unit.

Example usage:

```cpp
YR::InvokeOptions opt;
opt.customResources.insert({"disk", 10})
```

After a scheduling task is successfully executed, openYuanrong supports user processes in obtaining the mount path information of allocated disks by reading the environment variable `YR_DISK_MOUNT_POINT`.
