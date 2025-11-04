# 单独使用数据系统

openYuanrong 数据系统提供了 Python/C++ 语言接口，封装 heterogeneous object/KV/object 多种语义，支撑业务实现数据快速读写。

- heterogeneous object：基于 NPU 卡的 HBM 内存抽象异构对象接口，实现昇腾 NPU 卡间数据高速直通传输。同时提供 H2D/D2H 高速迁移接口，实现数据快速在 DRAM/HBM 之间传输。
- KV：基于共享内存实现免拷贝的 KV 数据读写，实现高性能数据缓存，支持通过对接外部组件提供数据可靠性语义。
- object：基于 host 侧共享内存抽象数据对象，实现基于引用计数的生命周期管理，将共享内存封装为 buffer，提供指针直接读写。

更多信息和示例，请参阅 openYuanrong 数据系统[代码仓库](https://gitee.com/openeuler/yuanrong-datasystem){target="_blank"}和[文档](https://pages.openeuler.openatom.cn/openyuanrong-datasystem/docs/zh-cn/latest/index.html){target="_blank"}。

## 入门

首先安装 openYuanrong 数据系统（包含 Python、C++ SDK以及命令行工具）：

```bash
pip install https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/openyuanrong_datasystem-0.5.0-cp39-cp39-manylinux_2_34_x86_64.whl
```

以 H2D(Host to Device)/D2H(Device to Host) 数据传输为例，异构对象接口提供了 MGetH2D 和 MSetD2H 接口，实现数据在 HBM 和 DRAM 之间快速 swap。

::::{tab-set}
:::{tab-item} Python

```python
import acl
import random
from datasystem.ds_client import DsClient

def random_str(slen=10):
    seed = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#%^*()_+=-"
    sa = []
    for _ in range(slen):
        sa.append(random.choice(seed))
    return ''.join(sa)

def hetero_mset_d2h_mget_h2d():
    client = DsClient("127.0.0.1", 31501)
    client.init()

    acl.init()
    device_idx = 0
    acl.rt.set_device(device_idx)

    key_list = [ 'key1', 'key2', 'key3' ]
    data_size = 1024 * 1024
    test_value = random_str(data_size)

    in_data_blob_list = []
    for _ in key_list:
        tmp_batch_list = []
        for _ in range(4):
            dev_ptr, _ = acl.rt.malloc(data_size, 0)
            acl.rt.memcpy(dev_ptr, data_size, acl.util.bytes_to_ptr(test_value.encode()), data_size, 1)
            blob = Blob(dev_ptr, data_size)
            tmp_batch_list.append(blob)
        blob_list = DeviceBlobList(device_idx, tmp_batch_list)
        in_data_blob_list.append(blob_list)

    out_data_blob_list = []
    for _ in key_list:
        tmp_batch_list = []
        for _ in range(4):
            dev_ptr, _ = acl.rt.malloc(data_size, 0)
            blob = Blob(dev_ptr, data_size)
            tmp_batch_list.append(blob)
        blob_list = DeviceBlobList(device_idx, tmp_batch_list)
        out_data_blob_list.append(blob_list)

    client.hetero().mset_d2h(key_list, in_data_blob_list)

    client.hetero().mget_h2d(key_list, out_data_blob_list, 60000)
```

:::

:::{tab-item} C++

```cpp

#include "datasystem/hetero_cache/hetero_client.h"
#include <acl/acl.h>

ConnectOptions connectOptions = { .host = "127.0.0.1", .port = 31501 };
auto client = std::make_shared<DsClient>(connectOptions);
ASSERT_TRUE(client->Init().IsOk());

// Initialize the ACL interface.
int deviceId = 1;
aclInit(nullptr);
aclrtSetDevice(deviceId); // Bind the NPU card.

std::vector<std::string> keys = { "test-key1" };
std::vector<uint64_t> blobSize = { 10, 20 };
int blobNum = blobSize.size();
std::vector<DeviceBlobList> swapOutBlobList;
swapOutBlobList.resize(keys.size());
// Allocate the HBM memory and fill it in the swapOutBlobList.
for (size_t i = 0; i < swapOutBlobList.size(); i++) {
    swapOutBlobList[i].deviceIdx = deviceId;
    for (int j = 0; j < blobNum; j++) {
        void *devPtr = nullptr;
        int code = aclrtMalloc(&devPtr, blobSize[j], ACL_MEM_MALLOC_HUGE_FIRST);
        // Copying Data to the Device Memory.
        // aclrtMemcpy(devPtr, blobSize[j], value.data(), size, aclrtMemcpyKind::ACL_MEMCPY_HOST_TO_DEVICE)
        ASSERT_EQ(code, 0);
        Blob blob = { .pointer = devPtr, .size = blobSize[j] };
        swapOutBlobList[i].blobs.emplace_back(std::move(blob));
    }
}

Status status = client->Hetero()->MSetD2H(keys, swapOutBlobList);
ASSERT_TRUE(status.IsOk());

std::vector<DeviceBlobList> swapInBlobList;
swapInBlobList.resize(keys.size());
// Allocate the HBM memory and fill it in the swapInBlobList.
for (size_t i = 0; i < swapInBlobList.size(); i++) {
    swapInBlobList[i].deviceIdx = deviceId;
    for (int j = 0; j < blobNum; j++) {
        void *devPtr = nullptr;
        int code = aclrtMalloc(&devPtr, blobSize[j], ACL_MEM_MALLOC_HUGE_FIRST);
        ASSERT_EQ(code, 0);
        Blob blob = { .pointer = devPtr, .size = blobSize[j] };
        swapInBlobList[i].blobs.emplace_back(std::move(blob));
    }
}
std::vector<std::string> failedList;
status = client->Hetero()->MGetH2D(keys, swapInBlobList, failedList, 1);
ASSERT_TRUE(status.IsOk());
```

:::
::::


