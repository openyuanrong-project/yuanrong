## runtime go sdk 包使用说明

### 准备工作

本 sdk 包需要用到的库文件在`yuanrong`包中，请先下载`yuanrong.tar.gz`并解压，然后设置环境变量：

```bash
export CGO_LDFLAGS="-L${WORKING_DIR}/yuanrong/runtime/service/go/bin -lcpplibruntime"
export LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:${WORKING_DIR}/yuanrong/runtime/service/go/bin
```

即可执行`go build`。