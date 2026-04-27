# Group

package: `org.yuanrong`.

## public class Group

The Group class is a scheduling class for grouped instances in openYuanrong, used to manage the lifecycle of grouped instances, including the creation and destruction of grouped instances.

### Interface description

#### public Group()

The constructor of the Group class.

#### public Group(String name, GroupOptions opts)

The constructor of the Group class.

- Parameters:

   - **name** (String) - The name of the instance group scheduling class, which must be unique.
   - **opts** (GroupOptions) - The lifecycle parameters for instance group scheduling.

#### public void invoke() throws YRException

The invoke interface for group scheduling instances, used for creating group instances.

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

1. A single group can create up to 256 instances in a group.

2. Up to 12 groups can be created concurrently, with each group creating up to 256 instances.

3. Calling the invoke() interface after NamedInstance::Export() will cause the current thread to hang.

4. Not calling the invoke() interface and directly sending a function request to a stateful function instance and getting the result will cause the current thread to hang.

5. Repeatedly calling the invoke() interface will result in an exception being thrown.

6. Instances within a group do not support specifying a detached lifecycle.

:::

- Throws:

   - **YRException** - Grouped instances follow the fate-sharing principle. If the creation of any instance fails, the interface throws an exception, and the creation of all instances fails.

#### public void terminate() throws YRException

The terminate interface for group scheduling instances, used for deleting group instances.

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

- Throws:

   - **YRException** - Exceptions that are uniformly thrown.

#### public void Wait() throws YRException

The Wait interface for group scheduling instances, used to wait for the completion of group instance creation.

- Throws:

   - **YRException** - Grouped instances adhere to the fate-sharing principle. If the creation of any instance fails, the interface throws an exception, resulting in the failure of all instance creations. The interface also throws an exception if the wait operation times out.

### Private Members

``` java

private String groupName
```

The name of the instance group scheduling class.

``` java

private GroupOptions groupOpts
```

Lifecycle parameters for instance group scheduling.

## public class GroupOptions

The class GroupOptions is a parameter of the openYuanrong group instance scheduling class, used to define lifecycle parameters for instance group scheduling, including the rescheduling timeout when kernel resources are insufficient.

### Interface description

#### public GroupOptions()

The constructor of the GroupOptions class.

#### public GroupOptions(int timeout)

The constructor of the GroupOptions class.

- Parameters:

   - **timeout** (int) - The rescheduling timeout for insufficient kernel resources in openYuanrong.

#### public GroupOptions(int timeout, boolean sameLifecycle)

The constructor of the GroupOptions class.

- Parameters:

   - **timeout** (int) - Rescheduling timeout for insufficient kernel resources in openYuanrong.
   - **sameLifecycle** (boolean) - Whether to enable fate-sharing configuration for grouped instances.

### Private Members

``` java

private int timeout
```

The timeout for rescheduling due to insufficient kernel resources in openYuanrong, in seconds. ``-1`` indicates that the kernel will retry scheduling indefinitely. Any other value less than ``0`` will throw an exception.

``` java

private boolean sameLifecycle = true
```

Whether to enable fate-sharing configuration for the group instances. ``true`` means that instances in the group will be created and destroyed together, while ``false`` means that instances can have independent lifecycles. The default is ``true``.
