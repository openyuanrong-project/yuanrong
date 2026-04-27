# 基于 openYuanrong 实现 MapReduce

MapReduce 将计算任务分解为 Map 和 Reduce 两个主要阶段，从而简化大规模数据集的并行处理。openYuanrong 提供的分布式编程接口可以轻松实现 MapReduce 编程模型，用于大数据处理。

本示例通过一个词频统计的简单例子，向您展示如何基于 openYuanrong 实现 MapReduce，它包含以下内容：

- 如何使用无状态函数并行多个 Map 任务。
- 如何使用数据对象传递多个 Map 任务的结果用于 Reduce 流程。

## 方案介绍

词频统计的主要阶段

1. Map 阶段
   输入的数据被分割成多个块（Split）, 每个 Map 任务处理其中的一个数据块， 分别统计单词出现的频率

2. Shuffle 阶段
   对 Map 输出的中间结果进行排序和分组，将相同键的值聚集在一起，以便 Reduce 任务处理

3. Reduce 阶段
   Reduce 函数对相同键的值进行归约操作，例如在词频统计中， 将相同单词出现的频率进行合并

## 准备工作

部署 openYuanrong 集群：[在主机上部署](../../deploy/deploy_processes/index.md)

## 实现流程

### 文本加载以及分割

我们通过 "import this" 来加载 Zen of Python 的文本内容，并将其分割为指定数量

```python
import sys
import subprocess

zen_of_python = subprocess.check_output([sys.executable, "-c", "import this"])
corpus = zen_of_python.decode('utf-8').strip().split()

num_partitions = 3
chunk = len(corpus) // num_partitions
partitions = [
    corpus[i * chunk: (i + 1) * chunk] for i in range(num_partitions)
]
```

### Mapping

我们通过 `@yr.invoke` 装饰普通函数 apply_map 来定义一个无状态函数，用来分布式并行处理数据块

```python
@yr.invoke
def apply_map(corpus, num_partitions=3):
    map_results = [list() for _ in range(num_partitions)]
    for document in corpus:
        for result in map_function(document):
            first_letter = result[0].decode("utf-8")[0]
            word_index = ord(first_letter) % num_partitions
            map_results[word_index].append(result)
    return map_results
```

### Shuffling and Reducing

reduce 阶段的目标是将所有对从返回值到相同的节点。我们创建一个字典，将每个分区上出现的所有单词相加

```python
@yr.invoke
def apply_reduce(results):
    reduce_results = dict()
    for res in results:
        for _, value_list in enumerate(yr.get(res)):
           for _, value_tuple in enumerate(value_list):
                key = value_tuple[0]
                value = value_tuple[1]
                if key not in reduce_results:
                    reduce_results[key] = 0
                reduce_results[key] += value

    return reduce_results
```

### 定义主流程

主流程中通过 `yr.init()` 初始化 openYuanrong 运行上下文，您可以根据自己的集群规模调整任务的数量。

最后，汇总结果进行输出，并调用 yr.finalize() 来清理上下文。

```python
import sys
import subprocess
import yr

if __name__ == '__main__':
    yr.init()

    # 加载 zen of python 数据
    zen_of_python = subprocess.check_output([sys.executable, "-c", "import this"])
    corpus = zen_of_python.split()

    # 将数据分为3块
    num_partitions = 3
    chunk = len(corpus) // num_partitions
    partitions = [
        corpus[i * chunk: (i + 1) * chunk] for i in range(num_partitions)
    ]

    # 调用远程函数 apply_map 并行处理数据
    map_results = [
        apply_map.invoke(data, num_partitions)
        for data in partitions
    ]

    # 调用远程函数 apply_reduce 处理 mapping 的结果
    reduce_result = apply_reduce.invoke(map_results)
    counts = yr.get(reduce_result)

    # 将结果进行排序输出
    sorted_counts = sorted(counts.items(), key=lambda item: item[1], reverse=True)
    for count in sorted_counts:
        print(f"{count[0].decode('utf-8')}: {count[1]}")

    yr.finalize()
```

### 运行程序

:::
:::{dropdown} 完整代码
:chevron: down-up
:icon: chevron-down

```python
import sys
import subprocess
import yr

@yr.invoke
def apply_map(corpus, num_partitions=3):
    map_results = [list() for _ in range(num_partitions)]
    for document in corpus:
        for result in map_function(document):
            first_letter = result[0].decode("utf-8")[0]
            word_index = ord(first_letter) % num_partitions
            map_results[word_index].append(result)
    return map_results

@yr.invoke
def apply_reduce(results):
    reduce_results = dict()
    for res in results:
        for _, value_list in enumerate(yr.get(res)):
           for _, value_tuple in enumerate(value_list):
                key = value_tuple[0]
                value = value_tuple[1]
                if key not in reduce_results:
                    reduce_results[key] = 0
                reduce_results[key] += value

    return reduce_results

def map_function(document):
    for word in document.lower().split():
        yield word, 1

if __name__ == '__main__':
    yr.init()

    # 加载 zen of python 数据
    zen_of_python = subprocess.check_output([sys.executable, "-c", "import this"])
    corpus = zen_of_python.split()

    # 将数据分为3块
    num_partitions = 3
    chunk = len(corpus) // num_partitions
    partitions = [
        corpus[i * chunk: (i + 1) * chunk] for i in range(num_partitions)
    ]

    # 调用远程函数 apply_map 并行处理数据
    map_results = [
        apply_map.invoke(data, num_partitions)
        for data in partitions
    ]

    # 调用远程函数 apply_reduce 处理 mapping 的结果
    reduce_result = apply_reduce.invoke(map_results)
    counts = yr.get(reduce_result)

    # 将结果进行排序输出
    sorted_counts = sorted(counts.items(), key=lambda item: item[1], reverse=True)
    for count in sorted_counts:
        print(f"{count[0].decode('utf-8')}: {count[1]}")

    yr.finalize()
```

:::

在一个已经部署 openYuanrong 的环境上运行以上 Python 程序，输出如下：

```bash
is: 10
better: 8
than: 8
the: 6
to: 5
of: 3
although: 3
be: 3
special: 2
unless: 2
one: 2
should: 2
do: 2
may: 2
never: 2
way: 2
if: 2
implementation: 2
idea.: 2
a: 2
explain,: 2
ugly.: 1
implicit.: 1
complex.: 1
complex: 1
complicated.: 1
flat: 1
readability: 1
counts.: 1
cases: 1
rules.: 1
python,: 1
peters: 1
simple: 1
sparse: 1
dense.: 1
aren't: 1
zen: 1
by: 1
tim: 1
beautiful: 1
explicit: 1
nested.: 1
enough: 1
break: 1
in: 1
face: 1
refuse: 1
one--: 1
only: 1
--obvious: 1
it.: 1
obvious: 1
first: 1
practicality: 1
purity.: 1
pass: 1
silently.: 1
silenced.: 1
ambiguity,: 1
guess.: 1
and: 1
preferably: 1
at: 1
you're: 1
dutch.: 1
beats: 1
errors: 1
explicitly: 1
temptation: 1
there: 1
that: 1
not: 1
now: 1
often: 1
*right*: 1
it's: 1
it: 1
idea: 1
--: 1
let's: 1
good: 1
are: 1
great: 1
more: 1
never.: 1
now.: 1
hard: 1
bad: 1
easy: 1
namespaces: 1
honking: 1
those!: 1

```
