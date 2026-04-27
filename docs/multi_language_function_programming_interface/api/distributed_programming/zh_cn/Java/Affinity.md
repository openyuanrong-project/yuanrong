# Affinity

包名：`org.yuanrong.affinity`。

:::{Note}

约束：

- 亲和配置条件仅在调度时生效。
- 如果任何实例已经在某个节点上被调度和运行，那么它将不会因为运行时节点上的标签变化而被驱逐，只有新的实例才需要匹配新的亲和条件。
- 特别地：需要强亲和或强反亲和某标签时，请确保实例创建完成或资源标签已经创建完成。另一种做法，分别为两个实例增加亲和配置，保证实例两两亲和。

:::

## public class Affinity

类 Affinity 是亲和调度配置类，可在 openYuanrong 函数调用选项中的进行设置。支持 NODE 和 Pod 两个层级的亲和与反亲和。

:::{Note}

1. 实例的标签只包含 key, 对于实例亲和，标签选择条件只能设置 LABEL_EXISTS, LABEL_NOT_EXISTS;

:::

### 接口说明

#### public Affinity(AffinityKind affinityKind, AffinityType affinityType, List<LabelOperator> labelOperators)

Affinity 的构造函数。

- 参数：

   - **affinityKind** - 指定亲和资源标签还是实例标签。
   - **affinityType** - 指定亲和类型，包括强亲和，强反亲和，弱亲和，弱反亲和。
   - **labelOperators** - 指定如何匹配标签，可以匹配多个。

### 私有成员

``` java

private AffinityKind affinityKind
```

表示亲和对象。可选参数为 ``AffinityKind.RESOURCE``（资源亲和） 和 ``AffinityKind.INSTANCE``（实例亲和）。

``` java

private AffinityType affinityType
```

表示亲和类型。可选参数为 ``AffinityType.PREFERRED``（弱亲和）、``AffinityType.PREFERRED_ANTI``（弱反亲和）、``AffinityType.PREFERRED``（强亲和）、``AffinityType.PREFERRED``（强反亲和）。

``` java

private List<LabelOperator> labelOperators
```

标签操作列表，指定如何匹配标签，可以匹配多个，这些条件间是且的关系，必须全部满足才匹配，详细信息请参见 LabelOperator 类。

## public class LabelOperator

类 LabelOperator 指定如何匹配标签。

### 接口说明

#### public LabelOperator(OperatorType operatorType, String key)

LabelOperator 的构造函数。

- 参数：

   - **operatorType** - 指定标签操作类型。
   - **key** - 标签的键。

#### public LabelOperator(OperatorType operatorType, String key, List<String> values)

LabelOperator 的构造函数。

```java

List<String> values = Arrays.asList(new String[]{"val1", "val2"});
LabelOperator labelOperator = new LabelOperator(OperatorType.LABEL_EXISTS, "key1", values);
```

- 参数：

   - **operatorType** - 指定标签操作类型。
   - **key** - 标签的键。
   - **values** - 标签的值。

### 私有成员

``` java

private OperatorType operatorType
```

标签操作类型。

``` java

private String key
```

标签的键。

``` java

private List<String> values
```

标签的值。

#### OperatorType 枚举类型介绍

| 枚举常量             | 说明                                |
|------------------|---------------------------------------|
| LABEL_IN         | 资源存在指定 key, value 在特定列表的标签。  |
| LABEL_NOT_IN     | 资源存在指定 key, value 不在特定列表的标签。|
| LABEL_EXISTS     | 资源存在指定 key 的标签。                 |
| LABEL_NOT_EXISTS | 资源不存在指定 key 的标签。               |
