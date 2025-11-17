# 配置函数实例亲和

本节通过 Python 示例向您介绍如何使用资源与实例亲和，以实现自定义的调度策略。

## 配置函数实例亲和资源

资源亲和适合解决函数实例调度到特定资源上的问题。您需要预先在资源上打标签，在主机上部署时，标签打在节点上。我们将在示例中展示强亲和资源的两种用法。

### 在主机部署环境上配置资源亲和

我们部署两个节点。

首先部署主节点，设置两个标签 `{"master":"dev"}` 和 `{"agent":"dev"}`。

```bash
yr start --master --labels="{\"master\":\"dev\",\"agent\":\"dev\"}"
```

再部署从节点，设置一个标签 `{"agent":"uat"}`。

```bash
# 替换 {Cluster master info} 为成功部署主节点时输出的主节点信息
yr start --labels="{\"agent\":\"uat\"}" --master_info "{Cluster master info}"
```

代码中有状态函数 Detector 的 show 方法将帮助打印函数实例所在节点信息。我们为两个 Detector 实例分别设置了资源亲和标签。主节点带有 key 为 master 的标签，因此第一个 Detector 实例不会部署在主节点上。第二个 Detector 实例不仅要求标签的 key 匹配，还要求标签的 value 也必须匹配。因此，第二个 Detector 实例只会部署在从节点上。

```python
# resource-affinity.py
import os
import yr

@yr.instance
class Detector:
    def __init__(self, number):
        self.number = number

    def show(self):
        print("Detector " + str(self.number) + ",NODE_ID:" + os.getenv('NODE_ID') + ",LABELS:" + os.getenv('LABELS'))

if __name__ == '__main__':
    yr.init()

    # 第一个Detector实例，亲和性设置在非master标签节点拉起
    affinity_not_master = yr.Affinity(yr.AffinityKind.RESOURCE, yr.AffinityType.REQUIRED, [yr.LabelOperator(yr.OperatorType.LABEL_NOT_EXISTS, "master")])
    opt_zero = yr.InvokeOptions()
    opt_zero.schedule_affinities = [affinity_not_master]

    try:
        detector = Detector.options(opt_zero).invoke(0)
        yr.get(detector.show.invoke())
        detector.terminate()
    except RuntimeError as e:
        print(e)


    # 第二个Detector实例，亲和性设置在标签key为agent，值为uat的节点拉起
    affinity_agent_with_uat = yr.Affinity(yr.AffinityKind.RESOURCE, yr.AffinityType.REQUIRED, [yr.LabelOperator(yr.OperatorType.LABEL_IN, "agent", ["uat"])])
    opt_one = yr.InvokeOptions()
    opt_one.schedule_affinities = [affinity_agent_with_uat]

    try:
        detector = Detector.options(opt_one).invoke(1)
        yr.get(detector.show.invoke())
        detector.terminate()
    except RuntimeError as e:
        print(e)
```

执行命令 `python resource-affinity.py` 运行程序。查看从节点上的函数日志文件 `{node_id}-user_func_std.log`，可见如下输出，表明两个实例都部署在标签为 `{"agent":"uat"}` 节点。

```bash
2025-07-18 17:12:33|56412d11-0000-4000-8000-005cef06b506|runtime-56412d11-0000-4000-8000-005cef06b506-c6d59c3a409e|INFO|Detector 0,NODE_ID:dggphis35945-2731346,LABELS:{"agent":"uat"}
2025-07-18 17:12:33|04d8cf02-727f-4714-8000-000000000071|runtime-04d8cf02-727f-4714-8000-000000000071-000000d8f917|INFO|Detector 1,NODE_ID:dggphis35945-2731346,LABELS:{"agent":"uat"}
```

## 配置函数实例间亲和

实例亲和适合解决不同函数实例组合部署在相同节点或者考虑容灾等情况需要分散部署的问题。我们将在示例中展示如何实现两种函数实例在节点上 1:1 部署。

### 在主机部署环境上配置实例亲和

我们部署两个节点。

首先部署主节点：

```bash
yr start --master
```

再部署从节点：

```bash
# 替换 {Cluster master info} 为成功部署主节点时输出的主节点信息
yr start --master_info "{Cluster master info}"
```

代码中有状态函数的 show 方法将帮助打印函数实例所在节点信息。我们配置 Detector 实例的标签 key 为 detector，不与有相同标签的实例部署在同一节点。Partner 实例配置标签 key 为 partner，不与有相同标签的实例部署在同一节点，但要与标签 key 为 detector 的实例部署在同一节点。

```python
# instance-affinity.py
import os
import yr

@yr.instance
class Detector:
    def __init__(self, number):
        self.number = number

    def show(self):
        print("Detector " + str(self.number) + ",NODE_ID:" + os.getenv('NODE_ID'))

@yr.instance
class Partner:
    def __init__(self, number):
        self.number = number

    def show(self):
        print("Partner " + str(self.number) + ",NODE_ID:" + os.getenv('NODE_ID'))


if __name__ == '__main__':
    yr.init()

    # 亲和性设置为detector实例彼此不部署在一起
    affinity_not_detector = yr.Affinity(yr.AffinityKind.INSTANCE, yr.AffinityType.REQUIRED, [yr.LabelOperator(yr.OperatorType.LABEL_NOT_EXISTS, "detector")])
    opt_d = yr.InvokeOptions()
    opt_d.labels = ["detector"]
    opt_d.schedule_affinities = [affinity_not_detector]

    # 亲和性设置为partner实例彼此不部署在一起，但必须和detector实例部署在一起
    operators = [yr.LabelOperator(yr.OperatorType.LABEL_NOT_EXISTS, "partner"), yr.LabelOperator(yr.OperatorType.LABEL_EXISTS, "detector")]
    affinity_with_detector = yr.Affinity(yr.AffinityKind.INSTANCE, yr.AffinityType.REQUIRED, operators)
    opt_p = yr.InvokeOptions()
    opt_p.labels = ["partner"]
    opt_p.schedule_affinities = [affinity_with_detector]


    try:
        instances_num = 2
        detectors = [Detector.options(opt_d).invoke(i) for i in range(instances_num)]
        for d in detectors:
            yr.get(d.show.invoke())

        partners = [Partner.options(opt_p).invoke(i) for i in range(instances_num)]
        for p in partners:
            yr.get(p.show.invoke())

        for i in range(instances_num):
            detectors[i].terminate()
            partners[i].terminate()
    except RuntimeError as e:
        print(e)
```

执行命令 `python instance-affinity.py` 运行程序。查看主从节点上的函数日志文件 `{node_id}-user_func_std.log`，可见如下输出，在每个节点上分别部署了一个 Detector 实例和一个 Partner 实例。

主节点：

```bash
2025-07-18 17:20:43|74f15c86-1100-4000-8000-00006cee4186|runtime-74f15c86-1100-4000-8000-00006cee4186-000000001165|INFO|Detector 0,NODE_ID:dggphis35893-1360565
2025-07-18 17:20:44|bb1d0000-0000-4000-80be-8d1a712b5b0e|runtime-bb1d0000-0000-4000-80be-8d1a712b5b0e-00000000008e|INFO|Partner 0,NODE_ID:dggphis35893-1360565
```

从节点：

```bash
2025-07-18 17:20:44|47120000-0000-4000-803c-5dff5d65a252|runtime-47120000-0000-4000-803c-5dff5d65a252-0083aa759f33|INFO|Detector 1,NODE_ID:dggphis35945-3030452
2025-07-18 17:20:44|ef042b74-7975-4e00-8000-00000000c91a|runtime-ef042b74-7975-4e00-8000-00000000c91a-00000000cd30|INFO|Partner 1,NODE_ID:dggphis35945-3030452
```
