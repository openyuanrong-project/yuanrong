# Affinity

package: `org.yuanrong.affinity`.

:::{Note}

Constraints:

- Affinity configuration only takes effect during scheduling.
- If any instance has already been scheduled and is running on a node, it will not be evicted due to label changes on the node at runtime. Only new instances need to match the new affinity conditions.
- Specifically, when strong affinity or strong anti-affinity to a certain label is required, ensure that the instance is created or the resource label is in place. Alternatively, add affinity configurations to each instance to ensure mutual affinity between instances.

:::

## public class Affinity

The Affinity class is a configuration class for affinity scheduling, which can be set in the options of the function call. It supports both node-level and pod-level affinity and anti-affinity.

:::{Note}

1. The instance label only contains the key. For instance affinity, the label selection condition can only be set to LABEL_EXISTS or LABEL_NOT_EXISTS.

:::

### Interface description

#### public Affinity(AffinityKind affinityKind, AffinityType affinityType, List<LabelOperator> labelOperators)

The constructor of the Affinity class.

- Parameters:

   - **affinityKind** - Specifies whether the affinity is for resource labels or instance labels.
   - **affinityType** - Specifies the affinity type, including strong affinity, strong anti-affinity, weak affinity, and weak anti-affinity.
   - **labelOperators** - Specifies how to match labels, and multiple matches can be specified.

### Private Members

``` java

private AffinityKind affinityKind
```

Indicates the affinity object. The optional parameters are ``AffinityKind.RESOURCE`` (resource affinity) and ``AffinityKind.INSTANCE`` (instance affinity).

``` java

private AffinityType affinityType
```

Indicates the affinity type. The optional parameters are ``AffinityType.PREFERRED`` (weak affinity), ``AffinityType.PREFERRED_ANTI`` (weak anti-affinity), ``AffinityType.REQUIRED`` (strong affinity), and ``AffinityType.REQUIRED_ANTI`` (strong anti-affinity).

``` java

private List<LabelOperator> labelOperators
```

The label operator list specifies how to match labels and can match multiple labels. These conditions are in an AND relationship, meaning all must be satisfied for a match. For detailed information, please refer to the LabelOperator class.

## public class LabelOperator

The class LabelOperator specifies how to match labels.

### Interface description

#### public LabelOperator(OperatorType operatorType, String key)

The constructor of the LabelOperator class.

- Parameters:

   - **operatorType** - Specifies the type of label operation.
   - **key** - The key of the label.

#### public LabelOperator(OperatorType operatorType, String key, List<String> values)

The constructor of the LabelOperator class specifies the type of label operation and the key of the label.

```java

List<String> values = Arrays.asList(new String[]{"val1", "val2"});
LabelOperator labelOperator = new LabelOperator(OperatorType.LABEL_EXISTS, "key1", values);
```

- Parameters:

   - **operatorType** - Specifies the type of label operation.
   - **key** - Specifies the key of the label.
   - **values** - Specifies the value of the label.

### Private Members

``` java

private OperatorType operatorType
```

The label operation type.

``` java

private String key
```

The key of the label.

``` java

private List<String> values
```

The value of the label.

#### Introduction to the OperatorType enumeration type

| Enumerated constants             | Description                                |
|------------------|---------------------------------------|
| LABEL_IN         | The resource has a label with the specified key and value in a specific list.  |
| LABEL_NOT_IN     | The resource has a label with the specified key and value not in a specific list.|
| LABEL_EXISTS     | The resource has a label with the specified key.           |
| LABEL_NOT_EXISTS | The resource does not have a label with the specified key. |
