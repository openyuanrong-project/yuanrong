# 有状态函数中使用 NPU 资源

AI 应用如机器学习任务中，异构计算资源已成为必须依赖。openYuanrong 支持管理和调度不同类型的计算资源，本节以 NPU 为例，介绍如何在有状态函数中指定和使用。

## 准备工作

一台或多台包含 NPU 卡的主机，在部署时使用 `--npu_collection_mode` 参数开启 NPU 信息自动采集。

以单节点为例，参考部署命令如下：

```bash
yr start --master --npu_collection_mode=all
```

## 使用示例

```python
import yr
import torch
import torch_npu
import os

@yr.instance
class NPUInstance:
    def __init__(self):
        self.device = None

    def getEnv(self):
        ascend_device_id = os.getenv('ASCEND_RT_VISIBLE_DEVICES')
        return (f"ASCEND_RT_VISIBLE_DEVICES: {ascend_device_id}")

    def run(self):
        if torch.npu.is_available():
            device0 = torch.device("npu:0")
            print(f"user can see device: {device0}")

            tensor0 = torch.randn(3, 3).to(device0)

            # if request 1 NPU, the following code will raise error
            # device1 = torch.device("npu:1")
            # tensor1 = torch.randn(3, 3).to(device1)
            return f"Tensor device: {tensor0.device}, result: \n {tensor0}"
        else:
            return "NPU is not available"


def main():
    yr.init()

    npu_resource = {"NPU/.+/count": 1} # request 1 NPU of any type
    opt = yr.InvokeOptions()
    opt.custom_resources = npu_resource
    npu_instance = NPUInstance.options(opt).invoke()

    res = npu_instance.run.invoke()
    print(yr.get(res))

    yr.finalize()

if __name__ == "__main__":
    main()
```

运行程序，正常输出如下：

```bash
Tensor device: npu:0, result:
 tensor([[-0.9499, -0.2434,  0.8055],
        [-0.6371,  0.7378, -0.5710],
        [ 0.8289, -0.4673,  1.3490]], device='npu:0')
```

打开注释部分代码逻辑，再次运行程序。因为指定了 1 个 NPU 但却使用 2 个 NPU，将抛出如下异常：

```bash
  File "/workspace/caseTest/npuTest/demoNpu.py", line 24, in run
    tensor1 = torch.randn(3, 3).to(device1)
RuntimeError: exchangeDevice:build/CMakeFiles/torch_npu.dir/compiler_depend.ts:42 NPU function error: c10_npu::SetDevice(d.index()), error code is 107001
[ERROR] 2025-04-10-11:26:58 (PID:1828121, Device:0, RankID:-1) ERR00100 PTA call acl api failed
[Error]: Invalid device ID.
```

这是因为 openYuanrong 基于昇腾原生能力，使用环境变量 `ASCEND_RT_VISIBLE_DEVICES` 提供了 NPU 隔离。如果您希望自行控制 NPU 资源的使用，可在部署 openYuanrong 前在节点上写入环境变量 `export YR_NOSET_ASCEND_RT_VISIBLE_DEVICES=1`。

查看更多 [NPU 资源使用规则](development-scheduling-config-npu)。
