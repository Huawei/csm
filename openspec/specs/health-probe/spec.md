# Health Probe Capability

## Purpose

为 CSM 服务容器提供 Kubernetes 存活探针端点，使 Kubernetes 容器运行时能够自动检测并重启不健康的 Pod。

## Requirements

### Requirement: 存活探针 HTTP 端点

CSM 存活探针 SHALL 在端口 9808（可通过 `--healthz-port` 配置）上暴露 `/healthz` HTTP 端点。该端点 MUST 返回 HTTP 200 表示进程存活。探针作为 sidecar 容器与 Prometheus collector 和拓扑服务 Pod 共同部署。

#### Scenario: Kubernetes 探针检查存活状态成功
- **WHEN** Kubernetes 向存活探针端口发送 GET 请求到 `/healthz`
- **THEN** 系统 SHALL 返回 HTTP 200，无响应体，表示进程存活

---

### Requirement: 存活探针请求超时

存活探针 SHALL 强制执行 10 秒的请求超时。如果健康检查处理程序未在超时时间内完成，请求上下文 SHALL 被取消。

#### Scenario: 健康检查在超时内完成
- **WHEN** GET 请求发送到 `/healthz` 且处理程序在 10 秒内完成
- **THEN** 系统 SHALL 正常返回 HTTP 200

#### Scenario: 健康检查超过超时
- **WHEN** GET 请求发送到 `/healthz` 且处理程序未在 10 秒内完成
- **THEN** 请求上下文 SHALL 被取消

---

### Requirement: 存活探针配置

存活探针 SHALL 支持以下命令行参数：

| 参数 | 默认值 | 描述 |
|---|---|---|
| `--ip-address` | `0.0.0.0` | 容器内监听 IP 地址 |
| `--healthz-port` | `9808` | healthz 请求监听端口 |
| `--log-file` | `liveness-probe` | 日志文件名 |
| `--namespace` | `huawei-csm` | CSM 部署命名空间 |

#### Scenario: 用户自定义端口配置
- **WHEN** 部署人员指定 `--healthz-port=8080` 启动存活探针
- **THEN** HTTP 服务器 SHALL 在端口 8080 上监听 healthz 请求

#### Scenario: 使用默认配置启动
- **WHEN** 部署人员不指定任何参数启动存活探针
- **THEN** HTTP 服务器 SHALL 在 `0.0.0.0:9808` 上监听 healthz 请求

---

### Requirement: 进程级健康指示

在进程内架构中，存活探针 SHALL 以进程本身作为健康指示器。只要进程在运行，探针始终返回成功。Kubernetes 通过容器运行时原生处理进程存活检测。

#### Scenario: 进程存活时探针返回成功
- **WHEN** 存活探针进程正在运行且可响应
- **THEN** `/healthz` 端点 SHALL 返回 HTTP 200，确认进程存活

#### Scenario: 存储后端不可达时探针仍返回成功
- **WHEN** 所有存储后端不可达，但存活探针进程仍在运行
- **THEN** `/healthz` 端点 SHALL 仍返回 HTTP 200；探针不检查后端连通性，仅反映进程存活

#### Scenario: 进程崩溃时 Kubernetes 自动重启
- **WHEN** 存活探针进程崩溃，无法响应 `/healthz` 请求
- **THEN** Kubernetes SHALL 通过容器运行时检测到进程退出，自动重启 Pod

---

### Requirement: 存活探针作为 Sidecar

存活探针 SHALL 作为 sidecar 容器部署在 `csm-prometheus-service` 和 `csm-storage-service` 两个 Kubernetes Deployment 中。每个 sidecar 实例 SHALL 独立监听自己的端口（9808）。

#### Scenario: Prometheus collector Pod 中的存活探针
- **WHEN** `csm-prometheus-service` Deployment 被创建
- **THEN** Pod SHALL 同时包含 `prometheus-collector` 容器（端口 8887）和 `liveness-probe` sidecar 容器（端口 9808）

#### Scenario: 拓扑服务 Pod 中的存活探针
- **WHEN** `csm-storage-service` Deployment 被创建
- **THEN** Pod SHALL 同时包含 `topo-service` 容器和 `liveness-probe` sidecar 容器（端口 9808）

#### Scenario: Prometheus Exporter 的 healthz 被自身 sidecar 探测
- **WHEN** liveness-probe sidecar 在 Prometheus collector Pod 中启动
- **THEN** Kubernetes SHALL 配置 liveness probe 指向 Prometheus Exporter 的 `/healthz` 端点（HTTPS 模式下为 `https://localhost:8887/healthz`），而非 sidecar 自身的 `/healthz`

---

### Requirement: 版本上报

存活探针 SHALL 在启动时初始化名为 `huawei-csm-version` 的 ConfigMap 到配置的命名空间中。版本信息 SHALL 包含组件名称 `liveness-probe` 及其版本字符串。

#### Scenario: 版本 ConfigMap 初始化
- **WHEN** 存活探针启动
- **THEN** 系统 SHALL 创建或更新 `huawei-csm-version` ConfigMap，写入版本信息

#### Scenario: 版本 ConfigMap 写入失败不影响服务
- **WHEN** 存活探针启动时版本 ConfigMap 初始化失败
- **THEN** 探针 SHALL 记录错误日志但仍继续运行，不因版本上报失败而终止
