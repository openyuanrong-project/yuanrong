# 配置并获取指标

通过部署配置可以设置多种指标上报模式及导出方式，方便您按需对接。

## 配置指标导出

监控在 openYuanrong 中默认是关闭状态。在 k8s 部署时，开启需要配置参数 `observer.metrics.enable` 为 `true`，并且在 `observer.metrics.metricsConfig` 中设置相应的导出器。在主机上部署时，开启需要在启动参数中增加 `--enable_metrics=true --metrics_config_file={file_name}.json`，配置文件的格式参考[配置示例](metrics-config-example)。

### 导出模式

openYuanrong 支持立即导出（immediatelyExport）和批量导出（batchExport）两种导出模式。

- `immediatelyExport`：采集数据后立即由导出器导出至接收后端。优势是上报即时性好，但占用系统资源较多。
- `batchExport`：采集数据达到一定时间或指标数据累计到一定数量后向由导出器导出至接收后端。优势是占用系统资源较少，减少了后端接收压力，但即时性较差。

不同导出模式下可配置具体的导出器，相关参数如下：

- `name`：用户自定义的导出模式名称。
- `enable`：导出模式是否开启。
- `exporters`：该导出模式下，使用的导出器，可配置多个。当前支持：fileExporter（文件导出器）、prometheusPushExporter（ prometheus 导出器）。

### 导出器

导出器包含以下配置：

- `enable`: 导出器是否开启。
- `enabledInstruments`: 导出器允许导出的指标。可填写多个指标名，指标名之间通过逗号 `,` 分隔。指标名参考[系统指标](./system-metrics.md)及[告警指标](./alarm-metrics.md)。
- `failureQueueMaxSize`：导出失败时，指标内存存储队列最大容量。当超过设置值时，数据将写入磁盘，默认值为 1000 条。
- `failureDataDir`：导出失败时，指标存储文件路径，默认为 `/home/sn/metrics/failure`。在 k8s 上部署时，该路径默认挂载到主机 `/var/paas/sys/metrics/cff/default/failureMetrics/` 目录下。
- `failureDataFileMaxCapacity`：导出失败时，指标存储文件最大容量。当文件大小超过该容量时，文件将被覆盖，默认值为 20MB。
- `batchSize`：批量导出条数。当存储的指标数超过设置值时，触发导出，默认值为 512 条。仅当导出模式为 `batchExport` 时该配置生效。
- `batchIntervalSec`：批量导出间隔。每间隔一定时间导出一次指标，默认值为 15 秒。仅当导出模式为 `batchExport` 时该配置生效。
- `initConfig`: 导出器需要设置的初始化参数，openYuanrong 支持文件导出器（fileExporter）和 prometheus导出器（prometheusPushExporter）两种导出模式。详见以下表格说明。
  - fileExporter 导出器初始化参数：

    | 初始化参数 | 说明 | 约束 |
    | ---------- | -------------------- | -------------------- |
    | fileDir     | 文件存放路径。 | 可选，为空时取日志存储路径。                                 |
    | fileName    | 文件名称。     | 可选，为空时默认 `{nodeID}-{componentName}-metrics.data`。 |
    | rolling     | 文件滚动配置，包含以下配置项： <br>`enable`：文件滚动开关，默认 `false`。<br>`maxFiles`：保留文件最大数量，默认值 `3`。<br>`maxSize`：文件最大容量，默认值 `100MB`。<br>`compress`：是否开启文件压缩，默认 `false`, 仅在开启滚动时生效。 | 可选，为空时不开启文件滚动。                                 |
    | contentType | 指标内容形式，可选以下两种之一。 <br>`STANDARD`: 标准格式，保留所有信息。<br>`LABELS`: 标签模式，仅保留指标标签。| 可选，默认为标准模式。 |

  - prometheusPushExporter 导出器初始化参数：
  
    | 初始化参数          | 说明                                  | 约束               |
    |-------------------|------------------------------------ |------------------|
    | ip                | prometheus push gateway 地址。         | 必填 例: `127.0.0.1` |
    | port              | prometheus push gateway 端口。         | 必填 例: `9091`     |
    | heartbeatUrl      | prometheus push gateway 心跳地址。       | 必填 例: `/healthy` |
    | heartbeatInterval | prometheus push gateway 心跳间隔 单位（ms）。 | 必填  例:  `5000`     |

(metrics-config-example)=

### 配置示例

openYuanrong 支持文件（fileExporter）和 [prometheus](https://prometheus.io/){target="_blank"} （prometheusPushExporter）两种导出器。
 
- 以下是一个文件导出器 `metrics config` 配置示例：

```json
{
  "backends": [
    {
      "immediatelyExport": {
        "name": "your name",
        "enable": true,
        "exporters": [{
          "fileExporter": {
            "enable": true,
            "enabledInstruments": ["yr_alarm"],
            "failureQueueMaxSize": 1000,
            "failureDataDir": "/home/sn/metrics/failure",
            "failureDataFileMaxCapacity": 20,
            "initConfig": {
              "fileDir": "/home/sn/metrics/file",
              "rolling": {
                "enable": true,
                "maxFiles": 3,
                "maxSize": 100,
                "compress": false
              },
              "contentType": "STANDARD"
            }
          }
        }]
      }
    },
    {
      "batchExport": {
        "name": "your name",
        "enable": true,
        "exporters": [
          {
            "fileExporter": {
              "enable": true,
              "enabledInstruments": ["yr_app_instance_billing_invoke_latency"],
              "batchSize": 2,
              "batchIntervalSec": 10,
              "failureQueueMaxSize": 3,
              "failureDataDir": "/home/sn/metrics/failure",
              "failureDataFileMaxCapacity": 20,
              "initConfig": {
                "fileDir": "/home/sn/metrics/file",
                "rolling": {
                  "enable": true,
                  "maxFiles": 3,
                  "maxSize": 100,
                  "compress": false
                },
                "contentType": "STANDARD"
              }
            }
          }
        ]
      }
    }
  ]
}
```

示例中同时使用了立即上报和批量上报两种导出模式。

- 以下是一个 prometheusPush 导出器 `metrics config` 配置示例：

```json
{
  "backends": [
    {
      "batchExport": {
        "name": "your name",
        "enable": true,
        "exporters": [
          {
            "prometheusPushExporter": {
              "enable": true,
              "enabledInstruments": ["yr_node_cpu_usage","yr_node_memory_usage","yr_etcd_alarm"],
              "batchSize": 10,
              "batchIntervalSec": 5,
              "failureQueueMaxSize": 100,
              "failureDataDir": "/home/sn/metrics/failure",
              "failureDataFileMaxCapacity": 20,
              "initConfig": {
                "ip": "your prometheus pushgateway ip",
                "port": 9091,
                "heartbeatUrl": "/healthy",
                "heartbeatInterval": 5000
              }
            }
          }
        ]
      }
    }
  ]
}
```

示例中使用了批量上报导出模式。

## 获取导出的指标数据

您可通过配置不同类型的导出器获取指标数据。

### 文件导出器

文件导出器（fileExporter）的导出目录可通过配置传入，默认使用日志目录。文件名称格式如下:

- 应用指标文件名形式：`yr_metrics_xxx.data` 。
- 系统指标文件名形式：`{nodeName}-{moduleName}-metrics.data`，例如：`pekphis355665-3445437-function_master-metrics.data`。

导出文件内容格式有 `STANDARD`、`LABELS` 两种。

- `STANDARD` 模式：以采集函数实例进程内存占用情况和最近一次的调用请求为例，文件内容如下：

    ```text
    {"name":"runtime_memory_usage_vm_size","description":"","type":"Gauge","unit":"KB","value":"11000000","timestamp_ms":1691056024621,"labels":{"job_id":"job01","instance_id":"ins01"}}

    {"name":"runtime_memory_usage_vm_rss","description":"","type":"Gauge","unit":"KB","value":"11000000","timestamp_ms":1691056024621,"labels":{"job_id":"job01","instance_id":"ins01"}}

    {"name":"runtime_memory_usage_rss_anon","description":"","type":"Gauge","unit":"KB","value":"11000000","timestamp_ms":1691056024621,"labels":{"job_id":"job01","instance_id":"ins01"}}
    ```

- `LABEL` 模式：以采集函数实例进程内存占用情况和最近一次的调用请求为例，文件内容如下：

    ```text
    {"job_id":"job01","instance_id": "ins01"}
    ```

    其中，文件内容字段含义如下：

    | 字段 | 说明 |
    | ---------- | ---- |
    | name         | 指标名称 |
    | description  | 指标描述 |
    | type         | 指标类型 |
    | unit         | 指标单位 |
    | value        | 采集数值 |
    | timestamp_ms | 采集时刻 |
    | labels       | 标签属性 |
