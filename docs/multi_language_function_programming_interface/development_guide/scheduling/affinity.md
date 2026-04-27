# 亲和

除了硬件资源的逻辑分配策略之外，openYuanrong 也提供了丰富的亲和接口。openYuanrong 基于标签实现亲和语义。标签是一个最小调度单元的别名或者标记，不同的节点或者实例可以有相同的标签，标签以键值对的形式表示。

## 资源标签

资源标签通常是 openYuanrong 最小调度单元在创建时采集或启动时分配的静态标签, 它在运行过程中不会因为节点故障恢复或者组件故障重启而发生变化。具体类型包含：

- 采集节点 ID 作为资源标签：主机上部署 openYuanrong 时，会自动获取 `NODE_ID` 环境变量的值，以键为 `NODE_ID`，值为 `${NODE_ID}` 作为该节点的标签。
- 自定义资源标签：

    主机上部署 openYuanrong 时，您可以通过命令行参数自定义该节点的资源标签，例如：

    ```shell
    yr start --master --labels="{\"Product\":\"Yuanrong\", \"NodeType\":\"worker\"}"
    ```

## 实例标签

实例标签在创建函数实例时通过配置项设置。例如：`opts.labels = ["instance:value1", "actor"]`, 表示创建的实例具有键为 `instance`，值为 `value1` 和 键为`actor` 两个标签。

## 亲和规则

配置的亲和规则仅在调度时生效。

一条亲和规则由三个属性定义：

- `AffinityKind`：表示亲和资源标签还是实例标签。
- `LabelOperator`：表示如何匹配标签，可以配置多个，这些条件间是且的关系，必须全部满足才匹配。
   - `LABEL_EXISTS`：标签的键必须存在。
   - `LABEL_NOT_EXISTS`：标签的键必须不存在。
   - `LABEL_IN`：标签的键必须存在且值在指定列表中。
   - `LABEL_NOT_IN`：标签的键必须存在且值不在指定的列表中。
- `AffinityType`：表示亲和类型，包括强亲和，强反亲和，弱亲和，弱反亲和。
   - `AffinityType.REQUIRED` 和 `AffinityType.REQUIRED_ANTI`：该实例必须在满足条件的资源上部署，否则不部署。
   - `AffinityType.PREFERRED` or `AffinityType.PREFERRED_ANTI`：该实例优先在满足条件的资源上部署，但不强制。当未找到满足条件的资源时，可能会选择其他有资源的节点或 Pod 上进行调度。您可以通过开启 `preferred_anti_other_labels` 配置，在弱亲和条件都不能满足时直接返回调度失败。

一个函数实例可以配置多条亲和规则，具有相同 `AffinityKind` 和 `AffinityType` 的亲和规则间是或的关系，满足任意一条即匹配。多条亲和规则也可以设置优先级，调度时会按顺序优先选择满足最早添加的亲和规则。

- `required_priority`：强亲和优先级调度开关。开启后，当传入多个强亲和条件，按顺序匹配打分，都不满足时，调度失败。
- `preferred_priority`：弱亲和优先级调度开关。开启后，当传入多个弱亲和条件，按顺序匹配打分，出现一个满足的即可调度成功。

您可参考以下示例配置亲和，更多完整示例参考[配置函数实例亲和](../../examples/affinity.md)。

亲和资源标签示例：

::::{tab-set}

:::{tab-item} Python

```python
import yr

operators = [yr.LabelOperator(yr.OperatorType.LABEL_EXISTS, "key1"), yr.LabelOperator(yr.OperatorType.LABEL_EXISTS, "key2")]

# 强资源亲和，匹配同时包含资源标签键为 "key1" 和 "key2" 的资源
affinity1 = yr.Affinity(yr.AffinityKind.RESOURCE, yr.AffinityType.REQUIRED, operators)
# 强资源亲和，匹配资源标签键为"key3"的资源
affinity2 = yr.Affinity(yr.AffinityKind.RESOURCE, yr.AffinityType.REQUIRED, yr.LabelOperator(yr.OperatorType.LABEL_EXISTS, "key3"))

invokeOpt = yr.InvokeOptions()
# 资源必须满足affinity1或者affinity2
invokeOpt.schedule_affinities = [affinity1, affinity2]
```

:::
:::{tab-item} C++

```c++
#include "yr/yr.h"

// 弱资源亲和，配置资源标签键为"key1"的资源
auto affinity1 = YR::ResourcePreferredAffinity(YR::LabelExistsOperator("key1"));

// 强资源亲和，配置资源标签键为"key2"，值为"value1"或"value2"的资源
auto affinity2 = YR::ResourceRequiredAffinity(YR::LabelInOperator("key2", {"value1, value2"}));

// 弱资源亲和，排除资源标签键为"key3"的资源
auto affinity3 = YR::ResourcePreferredAffinity({YR::LabelDoesNotExistOperator("key3")});

// 弱资源亲和，配置资源标签同时包含键为"key4"且键为"key5"，值为"value3"的资源
auto affinity4 = YR::ResourcePreferredAffinity({YR::LabelExistsOperator("key4"), YR::LabelNotInOperator("key5", {"value3"})});

YR::InvokeOptions opts;
// 资源必须满足 affinity2，尽量满足 affinity1 或 affinity3 或 affinity4
opts.AddAffinity(affinity1).AddAffinity(affinity2).AddAffinity({affinity3, affinity4});
```

:::
:::{tab-item} Java

```java
import java.util.ArrayList;
import org.yuanrong.InvokeOptions;
import org.yuanrong.affinity.*;

// 弱资源亲和，匹配资源标签键为"key1"的资源
LabelOperator labelOperator = new LabelOperator(OperatorType.LABEL_EXISTS, "key1");
ArrayList<LabelOperator> operatorsList = new ArrayList<>();
operatorsList.add(labelOperator);
Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, operatorsList);

ArrayList<Affinity> affinityList = new ArrayList<>();
affinityList.add(affinity);

// 弱资源亲和，匹配资源标签键为"key2"的资源
LabelOperator labelOperator1 = new LabelOperator(OperatorType.LABEL_EXISTS, "key2");
ArrayList<LabelOperator> operatorsList1 = new ArrayList<>();
operatorsList1.add(labelOperator1);

// 资源尽量满足资源标签键为"key1"或键为"key2"
InvokeOptions options = new InvokeOptions.Builder().scheduleAffinity(affinityList)
    .addResourceAffinity(AffinityType.PREFERRED, operatorsList1)
    .build();
```

:::
::::

亲和实例标签示例：

::::{tab-set}

:::{tab-item} Python

```python
import yr

operators = [yr.LabelOperator(yr.OperatorType.LABEL_EXISTS, "key1"), yr.LabelOperator(yr.OperatorType.LABEL_EXISTS, "key2")]

# 强实例亲和，匹配同时包含实例标签键为"key1"和"key2"的资源
affinity1 = yr.Affinity(yr.AffinityKind.INSTANCE, yr.AffinityType.REQUIRED, operators)
# 强实例亲和，匹配包含实例标签键为"key3"的资源
affinity2 = yr.Affinity(yr.AffinityKind.INSTANCE, yr.AffinityType.REQUIRED, yr.LabelOperator(yr.OperatorType.LABEL_EXISTS, "key3"))

invokeOpt = yr.InvokeOptions()
# 资源必须满足affinity1或者affinity2
invokeOpt.schedule_affinities = [affinity1, affinity2]
```

:::
:::{tab-item} C++

```c++
#include "yr/yr.h"

// 弱实例亲和，匹配包含实例标签键为"key1"的资源
auto affinity1 = YR::InstancePreferredAffinity(YR::LabelExistsOperator("key1"));

// 强实例亲和，配置包含实例标签键为"key2"，值为"value1"或"value2"的资源
auto affinity2 = YR::InstanceRequiredAffinity(YR::LabelInOperator("key2", {"value1, value2"}));

YR::InvokeOptions opts;
// 资源必须满足 affinity2，尽量满足 affinity1
opts.AddAffinity(affinity1).AddAffinity(affinity2);
```

:::
:::{tab-item} Java

```java
import java.util.ArrayList;
import org.yuanrong.InvokeOptions;
import org.yuanrong.affinity.*;

// 弱实例亲和，匹配包含实例标签键为"key1"的资源
LabelOperator labelOperator = new LabelOperator(OperatorType.LABEL_EXISTS, "key1");
ArrayList<LabelOperator> operatorsList = new ArrayList<>();
operatorsList.add(labelOperator);
Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, operatorsList);

ArrayList<Affinity> affinityList = new ArrayList<>();
affinityList.add(affinity);

// 弱实例亲和，匹配包含实例标签键为"key2"的资源
LabelOperator labelOperator1 = new LabelOperator(OperatorType.LABEL_EXISTS, "key2");
ArrayList<LabelOperator> operatorsList1 = new ArrayList<>();
operatorsList1.add(labelOperator1);

// 资源尽量满足实例标签键为"key1"或键为"key2"
InvokeOptions options = new InvokeOptions.Builder().scheduleAffinity(affinityList)
    .addInstanceAffinity(AffinityType.PREFERRED, operatorsList1)
    .build();
```

:::
::::

:::{note}

- 如果任何实例已经在某个节点上被调度和运行，那么它将不会因为运行时节点上的标签变化而被驱逐，只有新的实例才需要匹配新的亲和条件。
- 需要强亲和或强反亲和某标签时，请确保带有该实例标签的实例或资源标签已经创建完成。如果您希望实例两两亲和，建议分别为两个实例增加亲和配置。

:::
