# Prometheus Metrics Export Capability

## Purpose

为 Kubernetes 环境中的华为企业存储系统提供 Prometheus 兼容的指标导出服务，将存储后端的对象状态和性能指标以 Prometheus 文本格式暴露给外部监控平台（如 Prometheus Server），实现对存储后端的可观测性。

## Requirements

### Requirement: Prometheus 指标端点

CSM Prometheus Exporter SHALL 通过 `/{monitorType}/{monitorBackendName}` 路径暴露 HTTP/HTTPS 端点，返回 Prometheus 文本展示格式的存储监控指标。端点 MUST 支持两种监控类型：`object`（对象指标）和 `performance`（性能指标）。端点 MUST 对监控类型进行校验，对非法值返回 HTTP 400。

#### Scenario: 用户请求存储后端的对象指标
- **WHEN** 运维人员通过 Prometheus 或 curl 发送 GET 请求到 `/object/{monitorBackendName}`，并在查询参数中指定采集器名称（如 `array`、`controller`、`storagepool`、`lun`、`filesystem`、`pv`、`vstore`）
- **THEN** 系统 SHALL 返回 HTTP 200，以 Prometheus 文本展示格式输出 `huawei_storage` 命名空间下的请求对象指标

#### Scenario: 用户请求存储后端的性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{monitorBackendName}`，查询参数指定采集器名称及对应指标列表
- **THEN** 系统 SHALL 返回 HTTP 200，包含指定后端和采集器类型的性能指标数据

#### Scenario: 用户在单次请求中查询多个采集器类型
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?array&controller&storagepool`
- **THEN** 系统 SHALL 返回 HTTP 200，在一个 Prometheus 抓取响应中包含 array、controller 和 storagepool 三种采集器的对象指标

#### Scenario: 用户请求的监控类型无效
- **WHEN** 运维人员发送 GET 请求到 `/{invalidType}/{monitorBackendName}`，其中 `invalidType` 不是 `object` 或 `performance`
- **THEN** 系统 SHALL 返回 HTTP 400，错误信息为 "MonitorType is invalid."

#### Scenario: 用户请求的采集器名称不在白名单中
- **WHEN** 运维人员发送 GET 请求的查询参数中包含不在支持列表（`array`、`controller`、`storagepool`、`lun`、`filesystem`、`pv`、`vstore`）中的采集器名称
- **THEN** 系统 SHALL 返回 HTTP 400，错误信息为 "MetricsObjectType is invalid."

#### Scenario: 用户请求性能指标但未指定指标名称
- **WHEN** 运维人员发送 GET 请求到 `/performance/{monitorBackendName}`，某个采集器名称后没有提供指标值（如 `?controller=`）
- **THEN** 系统 SHALL 返回 HTTP 400，错误信息为 "MetricsObjectType is invalid."

#### Scenario: 用户请求的 URL 路径格式错误
- **WHEN** 运维人员发送 GET 请求的 URL 路径不包含恰好两个路径段（monitorType/monitorBackendName）
- **THEN** 系统 SHALL 返回 HTTP 400，错误信息为 "URL is invalid."

#### Scenario: 用户查询的后端不存在或不可达
- **WHEN** 运维人员发送 GET 请求到 `/object/{不存在的backendName}?array`，该后端在缓存中无数据
- **THEN** 系统 SHALL 返回 HTTP 200，但响应体中不包含任何指标数据（静默返回空结果），Prometheus 抓取将记录该目标无数据

#### Scenario: 批量采集时单个采集器失败不影响整体
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?array&controller`，其中 controller 采集器因后端不可达而失败
- **THEN** 系统 SHALL 返回 HTTP 200，仍包含 array 采集器的指标数据；失败的采集器指标被静默跳过，不导致整个请求失败

---

### Requirement: HTTPS/TLS 支持

Prometheus Exporter SHALL 同时支持 HTTP 和 HTTPS 传输。启用 HTTPS（默认行为）时，服务器 MUST 使用 TLS 1.2 或更高版本。TLS 证书和密钥 MUST 从 Kubernetes Secret 挂载到 `/etc/secret-volume/tls.crt` 和 `/etc/secret-volume/tls.key`。

#### Scenario: 默认启用 HTTPS 模式
- **WHEN** 部署人员未配置 `--use-https` 参数（默认为 true）
- **THEN** 服务器 SHALL 以 HTTPS 模式监听，使用 TLS 1.2 最低版本提供加密连接

#### Scenario: 用户禁用 HTTPS 使用纯 HTTP
- **WHEN** 部署人员设置 `--use-https=false`
- **THEN** 服务器 SHALL 以纯 HTTP 模式监听，不使用 TLS 加密

#### Scenario: HTTPS 模式下证书文件缺失
- **WHEN** 部署人员启用 HTTPS 但 TLS 证书或密钥文件不存在于挂载路径
- **THEN** 服务 SHALL 启动失败并记录错误日志，进程退出

---

### Requirement: Prometheus 健康检查端点

Prometheus Exporter SHALL 暴露 `/healthz` HTTP 端点，返回 HTTP 200 表示服务存活。该端点 MUST 与指标端点使用同一端口。

#### Scenario: Kubernetes 探针检查 Prometheus Exporter 健康
- **WHEN** Kubernetes 向 `/healthz` 发送 GET 请求
- **THEN** 系统 SHALL 返回 HTTP 200，无响应体

#### Scenario: 健康检查端点不反映后端存储连通性
- **WHEN** 存储后端全部不可达，Kubernetes 向 `/healthz` 发送 GET 请求
- **THEN** 系统 SHALL 仍返回 HTTP 200；健康检查仅反映进程存活状态，不检查后端连通性

---

### Requirement: 存储阵列指标

系统 SHALL 收集并导出 `array` 采集器类型的对象指标，前缀为 `huawei_storage_array_`：

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `basic_info` | 存储阵列基本信息 | endpoint, sn, model, version, object |
| `health_status` | 存储阵列健康状态 | endpoint, sn, status, object |
| `running_status` | 存储阵列运行状态 | endpoint, sn, status, object |

`array` 采集器 SHALL NOT 支持性能指标。

#### Scenario: 用户采集存储阵列对象指标
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?array`
- **THEN** 系统 SHALL 返回 `huawei_storage_array_basic_info`、`huawei_storage_array_health_status` 和 `huawei_storage_array_running_status` 指标及对应标签

#### Scenario: 用户尝试获取存储阵列性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{backendName}?array=total_iops`
- **THEN** 系统 SHALL 返回 HTTP 400，因为 array 不支持性能指标

---

### Requirement: 控制器指标

系统 SHALL 收集并导出 `controller` 采集器类型的对象和性能指标，前缀为 `huawei_storage_controller_`：

**对象指标：**

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `cpu_usage` | CPU 利用率 (%) | endpoint, id, name, object |
| `memory_usage` | 内存利用率 (%) | endpoint, id, name, object |
| `health_status` | 健康状态 | endpoint, id, status, name, object |
| `running_status` | 运行状态 | endpoint, id, status, name, object |

**性能指标：**

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `total_iops` | 总 IOPS | endpoint, id, name, object |
| `read_iops` | 读 IOPS | endpoint, id, name, object |
| `write_iops` | 写 IOPS | endpoint, id, name, object |
| `total_bandwidth` | 总带宽 | endpoint, id, name, object |
| `read_bandwidth` | 读带宽 | endpoint, id, name, object |
| `write_bandwidth` | 写带宽 | endpoint, id, name, object |
| `avg_io_response_time` | 平均 IO 响应时间 | endpoint, id, name, object |
| `ops` | 每秒操作数 | endpoint, id, name, object |
| `avg_read_ops_response_time` | 平均读 OPS 响应时间 | endpoint, id, name, object |
| `avg_write_ops_response_time` | 平均写 OPS 响应时间 | endpoint, id, name, object |

#### Scenario: 用户采集控制器对象指标
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?controller`
- **THEN** 系统 SHALL 返回 `huawei_storage_controller_cpu_usage`、`huawei_storage_controller_memory_usage`、`huawei_storage_controller_health_status` 和 `huawei_storage_controller_running_status` 指标

#### Scenario: 用户采集控制器性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{backendName}?controller=total_iops,read_iops,write_iops`
- **THEN** 系统 SHALL 返回后端中每个控制器的指定性能指标

---

### Requirement: 存储池指标

系统 SHALL 收集并导出 `storagepool` 采集器类型的对象和性能指标，前缀为 `huawei_storage_storagepool_`：

**对象指标：**

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `total_capacity` | 总容量 (GB) | endpoint, id, name, object |
| `free_capacity` | 可用容量 (GB) | endpoint, id, name, object |
| `capacity_usage` | 容量使用率 (%) | endpoint, id, name, object |
| `used_capacity` | 已用容量 (GB) | endpoint, id, name, object |

**性能指标：** 与控制器相同的指标集（total_iops, read_iops, write_iops, total_bandwidth, read_bandwidth, write_bandwidth, avg_io_response_time, ops, avg_read_ops_response_time, avg_write_ops_response_time）。

#### Scenario: 用户采集存储池对象指标
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?storagepool`
- **THEN** 系统 SHALL 返回每个存储池的容量和使用率指标

#### Scenario: 用户采集存储池性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{backendName}?storagepool=total_iops,total_bandwidth`
- **THEN** 系统 SHALL 返回每个存储池的指定性能指标

---

### Requirement: LUN 指标

系统 SHALL 收集并导出 `lun` 采集器类型的对象和性能指标，前缀为 `huawei_storage_lun_`：

**对象指标：**

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `capacity` | LUN 容量 (GB) | endpoint, id, name, object |
| `capacity_usage` | LUN 容量使用率 (%) | endpoint, id, name, object |

**性能指标：** 与控制器相同的指标集。

LUN 采集 MUST 使用分页模式进行大规模 LUN 查询。

#### Scenario: 用户采集 LUN 对象指标
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?lun`
- **THEN** 系统 SHALL 返回每个 LUN 的 `huawei_storage_lun_capacity` 和 `huawei_storage_lun_capacity_usage` 指标

#### Scenario: 用户采集 LUN 性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{backendName}?lun=total_iops,read_iops`
- **THEN** 系统 SHALL 返回每个 LUN 的指定性能指标

---

### Requirement: 文件系统指标

系统 SHALL 收集并导出 `filesystem` 采集器类型的对象和性能指标，前缀为 `huawei_storage_filesystem_`：

**对象指标：**

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `capacity` | 文件系统容量 (GB) | endpoint, id, name, object |
| `capacity_usage` | 文件系统容量使用率 (%) | endpoint, id, name, object |
| `snapshot_used_capacity` | 文件系统快照已用容量 (GB) | endpoint, id, name, object |

**性能指标：** 与控制器相同的指标集。

文件系统采集 MUST 使用分页模式进行大规模文件系统查询。

#### Scenario: 用户采集文件系统对象指标
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?filesystem`
- **THEN** 系统 SHALL 返回每个文件系统的 capacity、capacity_usage 和 snapshot_used_capacity 指标

#### Scenario: 用户采集文件系统性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{backendName}?filesystem=total_iops,total_bandwidth`
- **THEN** 系统 SHALL 返回每个文件系统的指定性能指标

---

### Requirement: VStore 指标

系统 SHALL 收集并导出 `vstore` 采集器类型的对象指标，前缀为 `huawei_storage_vstore_`：

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `total_capacity` | VStore 总容量 (GB) | endpoint, id, name, object, pool |
| `free_capacity` | VStore 可用容量 (GB) | endpoint, id, name, object, pool |
| `used_capacity` | VStore 已用容量 (GB) | endpoint, id, name, object, pool |
| `capacity_usage` | VStore 容量使用率 (%) | endpoint, id, name, object, pool |

`vstore` 采集器 SHALL NOT 支持性能指标。VStore 数据来源于 Kubernetes 自定义资源，而非存储 REST API。

#### Scenario: 用户采集 VStore 对象指标
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?vstore`
- **THEN** 系统 SHALL 返回每个 vstore 的容量和使用率指标，包含 `pool` 标签

#### Scenario: VStore 缺少 VStoreID/VStoreName 时使用默认值
- **WHEN** StorageBackendContent 的 Status 中缺少 VStoreID 或 VStoreName
- **THEN** 系统 SHALL 使用默认值 "0" 作为 VStoreID，"System_vStore" 作为 VStoreName 继续导出指标

#### Scenario: VStore 仅收集支持的存储类型后端
- **WHEN** StorageBackendClaim 的 StorageType 不在支持列表（oceanstor-nas、oceanstor-san）中
- **THEN** 系统 SHALL 跳过该后端，不在 VStore 指标中输出其数据

---

### Requirement: PV 指标

系统 SHALL 收集并导出 `pv` 采集器类型的对象和性能指标，前缀为 `huawei_storage_pv_`：

**对象指标：**

| 指标名称 | 描述 | 标签 |
|---|---|---|
| `capacity` | PV 容量 (GB) | backend, pv_name, pvc_name, object, storage_volume_type, storage_volume_id, storage_volume_name |
| `capacity_usage` | PV 容量使用率 (%) | backend, pv_name, pvc_name, object, storage_volume_type, storage_volume_id, storage_volume_name |

**性能指标（LUN 类型 PV）：**

| 指标名称 | 描述 |
|---|---|
| `lun_total_bandwidth` | 总带宽 (MB/s) |
| `lun_pv_lun_total_iops` | 总 IOPS (IO/s) |
| `lun_avg_io_response_time` | 平均 IO 响应时间 (us) |

**性能指标（文件系统类型 PV）：**

| 指标名称 | 描述 |
|---|---|
| `filesystem_ops` | 每秒操作数 |
| `filesystem_avg_read_ops_response_time` | 平均读 OPS 响应时间 (us) |
| `filesystem_avg_write_ops_response_time` | 平均写 OPS 响应时间 (us) |

PV 指标通过关联 Kubernetes PersistentVolume 信息与底层存储 LUN 或文件系统数据生成。

#### Scenario: 用户采集 PV 对象指标
- **WHEN** 运维人员发送 GET 请求到 `/object/{backendName}?pv`
- **THEN** 系统 SHALL 返回 `huawei_storage_pv_capacity` 和 `huawei_storage_pv_capacity_usage` 指标及 PV 相关标签

#### Scenario: 用户采集 LUN 类型 PV 的性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{backendName}?pv=lun_total_bandwidth,lun_pv_lun_total_iops,lun_avg_io_response_time`
- **THEN** 系统 SHALL 返回 LUN 存储后端 PV 的性能指标

#### Scenario: 用户采集文件系统类型 PV 的性能指标
- **WHEN** 运维人员发送 GET 请求到 `/performance/{backendName}?pv=filesystem_ops,filesystem_avg_read_ops_response_time,filesystem_avg_write_ops_response_time`
- **THEN** 系统 SHALL 返回文件系统存储后端 PV 的性能指标

#### Scenario: PV 无底层存储映射时静默跳过
- **WHEN** 运维人员采集 PV 指标时，某个 PV 的 VolumeHandle 无法在 StorageBackendClaim 列表中找到对应后端
- **THEN** 系统 SHALL 静默跳过该 PV，不输出其指标，其余 PV 指标正常返回

#### Scenario: PV 解析失败不影响其他 PV
- **WHEN** 运维人员采集 PV 指标时，某个 PV 不是 CSI 卷、驱动不匹配或 VolumeHandle 格式错误
- **THEN** 系统 SHALL 在 Debug 级别记录日志并跳过该 PV，其余 PV 指标正常返回

#### Scenario: PV 性能指标简写映射
- **WHEN** 运维人员使用简写指标名查询 PV 性能指标（如 `lun` 展开为指标 21/22/370，`filesystem` 展开为指标 182/524/525）
- **THEN** 系统 SHALL 将简写名映射为对应的存储性能指标编号进行查询

---

### Requirement: 动态采集器注册

系统 SHALL 支持在启动时动态注册采集器类型。通过 `collectdef.MustRegister` 注册的采集器定义 SHALL 自动添加到 HTTP 查询白名单，允许新的存储对象类型在无需修改 HTTP handler 代码的情况下被查询。

#### Scenario: 新注册的采集器类型可被查询
- **WHEN** 开发人员通过 `collectdef.MustRegister` 注册了新的采集器类型（启用 `SupportObject` 或 `SupportPerformance`）
- **THEN** 该采集器名称 SHALL 出现在 HTTP 白名单中，可作为有效的查询参数在指标请求中使用

---

### Requirement: 指标值解析容错

系统 SHALL 对指标值解析失败进行容错处理。当单个指标的值无法解析（如非数字值）或标签解析失败时，系统 SHALL 跳过该指标，继续导出同一条记录中的其余指标，不导致整个请求失败。

#### Scenario: 单个指标值解析失败
- **WHEN** 某个采集器返回的数据中某条记录的 CAPACITY 字段为非数字值
- **THEN** 系统 SHALL 跳过该条 CAPACITY 指标，同一条记录中其余指标（如 health_status）仍正常导出

#### Scenario: 标签数量不匹配
- **WHEN** 某个指标的标签解析结果与定义的标签数量不一致
- **THEN** 系统 SHALL 跳过该指标，不影响其他指标导出

---

### Requirement: 空数据响应处理

当采集器返回的数据详情为空（Details 为 nil 或长度为 0）时，系统 SHALL 不输出该采集器的任何指标，并在 Debug 级别记录日志。Prometheus 将看到该指标不存在而非收到错误。

#### Scenario: 采集器返回空数据
- **WHEN** 用户请求的某个采集器在后端采集到零条数据（如 LUN 列表为空）
- **THEN** 系统 SHALL 不输出该采集器的任何 Prometheus 指标，请求中其他采集器指标正常返回

---

### Requirement: Kubernetes 服务暴露

Prometheus Exporter SHALL 以 Kubernetes NodePort 类型 Service 暴露。默认容器端口 SHALL 为 8887，默认 NodePort SHALL 为 30074。服务 MUST 部署 liveness probe sidecar 容器在端口 9808。

#### Scenario: Prometheus 通过 NodePort 抓取指标
- **WHEN** 外部 Prometheus 服务器抓取 `http://<nodeIP>:30074/object/<backendName>?array`
- **THEN** 系统 SHALL 响应指定后端的 array 对象指标

---

### Requirement: Prometheus Exporter 优雅关闭

Prometheus Exporter SHALL 处理 SIGINT 和 SIGTERM 信号实现优雅关闭。收到关闭信号后，系统 MUST 在终止前清理 exporter client set。

#### Scenario: 收到信号后优雅关闭
- **WHEN** 进程收到 SIGINT 或 SIGTERM
- **THEN** 系统 SHALL 关闭 shutdown channel、删除 exporter client set（触发 cmicore.Core.Stop 注销所有后端客户端），然后干净退出

---

### Requirement: Prometheus Exporter 配置参数

Prometheus Exporter SHALL 支持以下命令行参数：

| 参数 | 默认值 | 描述 |
|---|---|---|
| `--ip-address` | `0.0.0.0` | 监听 IP 地址 |
| `--exporter-port` | `8887` | HTTP(S) 监听端口 |
| `--backend-namespace` | `huawei-csi` | StorageBackendClaim 所在命名空间 |
| `--csi-driver-name` | `csi.huawei.com` | CSI 驱动名称过滤 |
| `--use-https` | `true` | 是否启用 HTTPS |

#### Scenario: 用户自定义监听端口
- **WHEN** 部署人员启动时指定 `--exporter-port=9999`
- **THEN** 服务器 SHALL 在端口 9999 上监听

#### Scenario: 使用默认配置启动
- **WHEN** 部署人员不指定任何参数启动
- **THEN** 服务器 SHALL 在 `0.0.0.0:8887` 上以 HTTPS 模式监听

---

### Requirement: ClientSet 初始化失败处理

当 Prometheus Exporter 的 Kubernetes 客户端初始化失败时，后续的指标数据获取 SHALL 失败，系统 SHALL 在日志中记录错误。单个采集器的数据获取失败不影响其他采集器的正常工作。

#### Scenario: ClientSet 初始化失败导致指标不可用
- **WHEN** Exporter 启动时 KubeClient 或 cmicore 初始化失败
- **THEN** 后续指标请求中涉及的存储数据获取 SHALL 失败，但 HTTP 端点仍可响应，请求中不依赖 ClientSet 的部分（如空数据）仍正常返回
