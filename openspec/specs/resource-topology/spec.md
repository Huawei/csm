# Resource Topology Capability

## Purpose

在 Kubernetes 中管理存储资源拓扑的生命周期与状态，通过集群级别的自定义资源定义（CRD）将 PersistentVolume 和 Pod 映射到其底层存储后端，实现拓扑感知的调度和监控。

## Requirements

### Requirement: ResourceTopology 自定义资源定义

系统 SHALL 定义集群级别的自定义资源定义（CRD），名称为 `ResourceTopology`，属于 API 组 `xuanwu.huawei.io/v1`，短名称为 `rt`。CRD MUST 包含 spec 和 status 子资源。

**ResourceTopology Spec：**

| 字段 | 类型 | 必填 | 描述 |
|---|---|---|---|
| `provisioner` | string | 是 | 卷 provisioner 名称 |
| `volumeHandle` | string | 是 | 后端名称和标识，格式：`<backend>.<identity>` |
| `tags` | []Tag | 是 | PV 及归属关系 |

**ResourceTopology Status：**

| 字段 | 类型 | 描述 |
|---|---|---|
| `status` | ResourceTopologyStatusPhase | 阶段：`Normal`、`Pending`、`Deleting` 或 `Crash` |
| `tags` | []Tag | 当前 PV 及归属关系 |

**Tag 结构：**

| 字段 | 类型 | 描述 |
|---|---|---|
| `namespace` | string | 资源命名空间 |
| `name` | string | 资源名称 |
| `kind` | string | 资源类型（通过 TypeMeta 内嵌） |
| `owner.namespace` | string | 属主资源命名空间 |
| `owner.name` | string | 属主资源名称 |
| `owner.kind` | string | 属主资源类型 |

CRD MUST 支持标准 Kubernetes 打印列：Provisioner、VolumeHandle、Status 和 Age。

#### Scenario: 用户通过 Kubernetes API 列出 ResourceTopology 资源
- **WHEN** 运维人员发送 GET 请求到 `/apis/xuanwu.huawei.io/v1/resourcetopologies`
- **THEN** Kubernetes API 服务器 SHALL 返回所有 ResourceTopology 对象列表

#### Scenario: 用户查询特定 ResourceTopology
- **WHEN** 运维人员发送 GET 请求到 `/apis/xuanwu.huawei.io/v1/resourcetopologies/{name}`
- **THEN** Kubernetes API 服务器 SHALL 返回指定名称的 ResourceTopology，包含其 spec 和 status

#### Scenario: 用户创建 ResourceTopology
- **WHEN** 运维人员发送 POST 请求到 `/apis/xuanwu.huawei.io/v1/resourcetopologies`，提供有效的 ResourceTopology spec
- **THEN** Kubernetes API 服务器 SHALL 创建 ResourceTopology 资源并返回创建的对象

#### Scenario: 用户更新 ResourceTopology status 子资源
- **WHEN** 控制器发送 PUT 请求到 `/apis/xuanwu.huawei.io/v1/resourcetopologies/{name}/status`，提供更新的 status
- **THEN** Kubernetes API 服务器 SHALL 仅更新指定 ResourceTopology 的 status 子资源

#### Scenario: 用户删除 ResourceTopology
- **WHEN** 运维人员发送 DELETE 请求到 `/apis/xuanwu.huawei.io/v1/resourcetopologies/{name}`
- **THEN** Kubernetes API 服务器 SHALL 删除该 ResourceTopology 资源

#### Scenario: 用户通过短名称查询
- **WHEN** 运维人员使用 `kubectl get rt` 查询 ResourceTopology
- **THEN** Kubernetes API 服务器 SHALL 返回所有 ResourceTopology 对象，等同于使用全名查询

---

### Requirement: ResourceTopology 状态阶段

ResourceTopology 的 status phase SHALL 为以下值之一：

| 阶段 | 描述 |
|---|---|
| `Normal` | 资源拓扑健康且一致 |
| `Pending` | 资源拓扑正在创建或更新中 |
| `Deleting` | 资源拓扑正在被删除 |
| `Crash` | 资源拓扑遇到致命错误 |

#### Scenario: 资源拓扑从 Pending 转为 Normal
- **WHEN** 拓扑控制器成功调和了处于 `Pending` 状态的 ResourceTopology
- **THEN** 状态阶段 SHALL 转变为 `Normal`

#### Scenario: 资源拓扑转为 Crash 状态
- **WHEN** 拓扑控制器在调和 ResourceTopology 时遇到不可恢复的错误
- **THEN** 状态阶段 SHALL 转变为 `Crash`

#### Scenario: 资源拓扑转为 Deleting 状态
- **WHEN** 拓扑控制器开始删除 ResourceTopology 的存储标签
- **THEN** 状态阶段 SHALL 先设置为 `Deleting`，待存储标签清理完成后再移除 finalizer

---

### Requirement: 拓扑控制器调和

拓扑控制器 SHALL 监视 ResourceTopology、PersistentVolume、PersistentVolumeClaim、Pod 和 StorageBackendClaim 资源，调和其状态以维护准确的资源拓扑映射。

控制器 MUST 处理三个工作队列：
1. **ResourceTopology 队列** - 处理 ResourceTopology CR 的增/改/删事件
2. **PersistentVolume 队列** - 处理支持驱动和后端的 CSI PV 的增/改/删事件
3. **Pod 队列** - 处理 Pod 的增/改/删事件

#### Scenario: PV 创建触发拓扑创建
- **WHEN** 具有 CSI 驱动和有效 StorageBackendClaim 后端的 PersistentVolume 被创建
- **THEN** 控制器 SHALL 创建对应的 ResourceTopology CR，包含 PV 的 provisioner、volumeHandle 和 tag 信息

#### Scenario: PV 删除触发拓扑移除
- **WHEN** PersistentVolume 被删除（从 informer 中检测到 NotFound）
- **THEN** 控制器 SHALL 删除对应的 ResourceTopology CR

#### Scenario: Pod 创建更新拓扑标签
- **WHEN** 使用了支持 CSI 驱动后端 PVC 的 Pod 被创建
- **THEN** 控制器 SHALL 更新对应 ResourceTopology CR 的 tags，将 Pod 添加为属主

#### Scenario: Pod 删除移除拓扑标签
- **WHEN** Pod 被删除
- **THEN** 控制器 SHALL 从对应 ResourceTopology CR 中移除该 Pod 的 tag

---

### Requirement: PV 事件验证

拓扑控制器 SHALL 在 PV 事件入队前进行验证。仅 CSI 驱动匹配配置的驱动名称、且具有有效的已初始化 StorageBackendClaim 后端（存储类型在支持列表中）的 PV SHALL 被处理。

#### Scenario: 非 CSI PV 被跳过
- **WHEN** 不具有 CSI driver spec 的 PersistentVolume 被创建或更新
- **THEN** 控制器 SHALL 跳过该 PV，不入队处理

#### Scenario: 不支持的 CSI 驱动 PV 被跳过
- **WHEN** CSI 驱动名称与配置的驱动名称不匹配的 PersistentVolume 被创建
- **THEN** 控制器 SHALL 跳过该 PV

#### Scenario: PV 后端未初始化时被跳过
- **WHEN** PV 引用的 StorageBackendClaim 没有 Status（CSI 驱动尚未初始化后端）
- **THEN** 控制器 SHALL 跳过该 PV

#### Scenario: PV 后端存储类型不支持时被跳过
- **WHEN** PV 引用的 StorageBackendClaim 的 StorageType 不在支持列表（oceanstor-nas、oceanstor-san）中
- **THEN** 控制器 SHALL 跳过该 PV，不创建 ResourceTopology CR

#### Scenario: PV 处于 Failed 阶段时被跳过
- **WHEN** PersistentVolume 的 Status.Phase 为 `Failed`
- **THEN** 控制器 SHALL 跳过该 PV，不触发 ResourceTopology 创建，且不会重试

---

### Requirement: ResourceTopology 名称过滤

拓扑控制器 SHALL 仅处理名称以 `rt-` 前缀开头的 ResourceTopology 资源。不具有该前缀的 ResourceTopology 对象 SHALL 在入队时被跳过。

#### Scenario: 不支持名称前缀的 ResourceTopology 被跳过
- **WHEN** 用户手动创建了一个名称不以 `rt-` 开头的 ResourceTopology
- **THEN** 控制器 SHALL 跳过该 ResourceTopology，不进行任何调和操作

#### Scenario: 控制器创建的 ResourceTopology 使用 rt- 前缀
- **WHEN** 控制器为 PV 创建 ResourceTopology CR
- **THEN** CR 名称 SHALL 为 `rt-{pvName}` 格式，确保可被控制器正常处理

---

### Requirement: ResourceTopology Finalizer 保护

控制器 SHALL 为每个同步的 ResourceTopology 添加 finalizer `resourcetopology.xuanwu.huawei.io/resourcetopology-protection`，防止在存储标签清理完成前被过早删除。删除时，控制器 MUST 先完成存储标签清理，再移除 finalizer。

#### Scenario: Finalizer 防止过早删除
- **WHEN** 用户删除 ResourceTopology CR
- **THEN** 由于 finalizer 存在，Kubernetes SHALL 保留该对象直到控制器完成存储标签清理并移除 finalizer

#### Scenario: Finalizer 更新失败触发重试
- **WHEN** 控制器在删除流程中更新 finalizer 失败
- **THEN** 控制器 SHALL 记录事件并重试该操作

---

### Requirement: ResourceTopology 删除流程

ResourceTopology 的删除 SHALL 按以下顺序执行：(1) 设置状态为 Deleting，(2) 删除 Status 中所有 tags 对应的存储标签，(3) 将 Spec 中的 PV/Pod tag 重新入队到各自的工作队列进行再同步，(4) 移除 finalizer。步骤 (2) 失败时 SHALL 触发整个操作重试。

#### Scenario: 删除流程清理存储标签后移除 finalizer
- **WHEN** ResourceTopology 被标记为删除
- **THEN** 控制器 SHALL 先清理所有存储标签，然后移除 finalizer，最后 Kubernetes 删除该 CR

#### Scenario: 存储标签清理失败触发重试
- **WHEN** 删除流程中某个存储标签删除失败
- **THEN** 控制器 SHALL 重试整个删除操作

#### Scenario: 删除完成后关联资源被重新入队
- **WHEN** ResourceTopology 删除流程完成
- **THEN** Spec 中引用的 PV 和 Pod SHALL 被重新入队到各自的工作队列，确保它们的状态被重新评估

---

### Requirement: PV 删除中的重试等待

当 PV 的 DeletionTimestamp 已设置但尚未完全移除时，控制器 SHALL 返回重试错误，使工作项被重新入队。直到 PV 完全删除后，控制器 SHALL 调用 `removeResourceTopology` 删除对应的 ResourceTopology CR。

#### Scenario: PV 正在删除中，控制器等待
- **WHEN** PersistentVolume 的 DeletionTimestamp 已设置但对象仍存在
- **THEN** 控制器 SHALL 返回重试错误，工作项以速率限制重新入队，等待 PV 完全删除

#### Scenario: PV 完全删除后清理 ResourceTopology
- **WHEN** PersistentVolume 从 informer 中已无法找到（NotFound）
- **THEN** 控制器 SHALL 调用 removeResourceTopology 删除名称为 `rt-{pvName}` 的 ResourceTopology CR

---

### Requirement: 孤儿 ResourceTopology 检测

在 ResourceTopology 同步过程中，控制器 SHALL 验证对应的 PV 是否仍然存在。如果 PV 已不存在（NotFound），控制器 SHALL 直接删除该 ResourceTopology CR。如果 Spec tags 中引用的 Pod 已不存在，控制器 SHALL 从 Spec 中移除该 Pod tag。

#### Scenario: 孤儿 ResourceTopology 被自动清理
- **WHEN** ResourceTopology 存在但其对应的 PV 已被删除
- **THEN** 控制器 SHALL 直接通过 API 调用删除该 ResourceTopology CR

#### Scenario: 已删除 Pod 的 tag 从 Spec 中清理
- **WHEN** ResourceTopology 的 Spec tags 中引用的 Pod 已不存在
- **THEN** 控制器 SHALL 从 Spec.Tags 中移除该 Pod 的条目，下次同步时将同步更新 Status

---

### Requirement: 状态更新冲突重试

控制器 SHALL 处理 Kubernetes 资源版本冲突，在更新 ResourceTopology status 时最多重试 10 次，每次间隔 100ms。每次重试前 SHALL 重新从 API 获取最新的 ResourceTopology。若 10 次均失败，SHALL 返回 "too many conflicts" 错误。

#### Scenario: 状态更新遇到版本冲突
- **WHEN** 控制器更新 ResourceTopology status 时遇到资源版本冲突
- **THEN** 控制器 SHALL 重新获取最新对象后重试，最多重试 10 次

#### Scenario: 状态更新重试耗尽
- **WHEN** 状态更新重试 10 次后仍然遇到版本冲突
- **THEN** 控制器 SHALL 返回错误，该工作项将被重新入队以速率限制重试

---

### Requirement: Pod 同步行为

Pod 同步 SHALL 仅处理容器处于 `Running` 状态的 PVC。尚未启动完成的容器的 PVC SHALL 被排除，Pod 将在 10 秒后重新入队。如果 Pod 对应的 ResourceTopology 不存在或正在删除中，Pod SHALL 在 10 秒后重新入队，等待 PV 控制器先创建 ResourceTopology。

#### Scenario: 仅运行中容器的 PVC 触发标签添加
- **WHEN** Pod 的部分容器已 Running、部分容器仍在启动中
- **THEN** 控制器 SHALL 仅为 Running 容器挂载的 PVC 添加标签，Pod 在 10 秒后重新入队以处理尚未就绪的容器

#### Scenario: ResourceTopology 不可用时 Pod 被延迟处理
- **WHEN** Pod 使用的 PV 对应的 ResourceTopology 尚未创建或正在删除中
- **THEN** 控制器 SHALL 不添加标签，而是在 10 秒后重新入队该 Pod

#### Scenario: Pod 标签删除时跳过正在删除的 ResourceTopology
- **WHEN** Pod 被删除，但其关联的 ResourceTopology 的 DeletionTimestamp 已设置
- **THEN** 控制器 SHALL 跳过该 ResourceTopology 的标签移除，因为删除流程会处理清理

---

### Requirement: 工作队列速率限制

拓扑控制器 SHALL 对所有三个工作队列（ResourceTopology、PersistentVolume、Pod）使用指数退避速率限制。失败的工作项 SHALL 以递增的延迟重新入队，直到配置的最大延迟。

#### Scenario: 失败项以速率限制重新入队
- **WHEN** 处理某个工作项失败
- **THEN** 该项 SHALL 使用指数退避速率限制重新入队

---

### Requirement: 缓存同步

拓扑控制器 SHALL 在启动 worker goroutine 之前等待 informer 缓存同步完成。如果缓存同步失败，控制器 SHALL 记录错误并终止。

#### Scenario: 缓存同步成功
- **WHEN** 所有 informer 缓存（ResourceTopology、PV、PVC、Pod、StorageBackendClaim）均已同步
- **THEN** 控制器 SHALL 为每个队列启动 worker goroutine

#### Scenario: 缓存同步失败
- **WHEN** 任何 informer 缓存同步失败
- **THEN** 控制器 SHALL 记录错误并关闭，不启动 worker

---

### Requirement: 拓扑控制器 Leader Election

拓扑控制器 SHALL 通过 `--enable-leader-election` 参数支持可选的 leader election。启用时，仅 leader 实例 SHALL 调和资源。Leader election 使用 ConfigMap 锁，租约时长 8 秒，续约超时 6 秒，重试间隔 2 秒。

#### Scenario: Leader election 启用
- **WHEN** leader election 启用且控制器实例获得 leadership
- **THEN** 该控制器 SHALL 开始调和 ResourceTopology、PV 和 Pod 资源

#### Scenario: Leader election 未启用（默认）
- **WHEN** leader election 未启用（默认）
- **THEN** 控制器 SHALL 立即开始调和，不等待 leadership

#### Scenario: 非 leader 实例待命
- **WHEN** leader election 启用但控制器实例未获得 leadership
- **THEN** 该控制器 SHALL 不执行任何调和操作，仅待命

---

### Requirement: 拓扑控制器配置参数

拓扑控制器 SHALL 支持以下命令行参数：

| 参数 | 默认值 | 描述 |
|---|---|---|
| `--controller-workers` | `4` | 每个 queue 的 worker 数量（必须 >= 1） |
| `--rt-retry-base-delay` | `5s` | ResourceTopology 重试基础延迟 |
| `--rt-retry-max-delay` | `5m` | ResourceTopology 重试最大延迟 |
| `--pv-retry-base-delay` | `5s` | PV 重试基础延迟 |
| `--pv-retry-max-delay` | `1m` | PV 重试最大延迟 |
| `--pod-retry-base-delay` | `5s` | Pod 重试基础延迟 |
| `--pod-retry-max-delay` | `1m` | Pod 重试最大延迟 |
| `--resync-period` | `15m` | 全量重新同步周期（必须 > 5m） |
| `--support-resources` | `Pod,PersistentVolume` | 支持的资源类型列表（必须 >= 2 种） |

#### Scenario: 用户自定义 worker 数量
- **WHEN** 部署人员指定 `--controller-workers=8`
- **THEN** 每个 queue SHALL 启动 8 个 worker goroutine

#### Scenario: 用户配置无效参数
- **WHEN** 部署人员指定 `--controller-workers=0`
- **THEN** 控制器 SHALL 校验失败并拒绝启动

---

### Requirement: StorageBackendClaim 生命周期联动

当 StorageBackendClaim 的 Spec 发生变更时，控制器 SHALL 注销旧客户端、从缓存移除、重新发现并登录新客户端。仅 Spec 差异（非 Status 差异）触发客户端重注册。当 StorageBackendClaim 被删除时，控制器 SHALL 注销客户端并从缓存中移除。

#### Scenario: StorageBackendClaim Spec 变更触发客户端重注册
- **WHEN** StorageBackendClaim 的 Spec 字段发生变更（如用户名、密码、地址变更）
- **THEN** 控制器 SHALL 注销旧客户端、从缓存移除、使用新 Spec 发现并登录新客户端

#### Scenario: StorageBackendClaim 仅 Status 变更不触发重注册
- **WHEN** StorageBackendClaim 仅 Status 字段发生变更（如 CSI 驱动更新后端状态）
- **THEN** 控制器 SHALL 不触发客户端重注册，继续使用缓存的客户端

#### Scenario: StorageBackendClaim 被删除
- **WHEN** StorageBackendClaim 被删除
- **THEN** 控制器 SHALL 注销对应客户端并从缓存中移除，后续对该后端的请求将因客户端不存在而失败
