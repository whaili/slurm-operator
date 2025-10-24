# 05-架构文档

## 系统整体架构综述

### 项目定位与价值主张

**Slinky** 是由 **SchedMD** 官方维护的 Kubernetes Slurm Operator，旨在统一 Kubernetes 平台和 Slurm 集群管理系统的优势：

- **统一管理**: 在 Kubernetes 平台上管理完整的 Slurm 集群
- **云原生**: 利用 Kubernetes 的容器化、调度、扩缩容能力
- **混合架构**: 支持部分组件在 Kubernetes 外部署的混合模式
- **企业级**: 提供高可用、安全、可扩展的 HPC 集群管理解决方案

### 架构特点
- **分层架构**: 明确的API层、控制层、构建层、工具层分离
- **控制器模式**: 使用Kubernetes控制器范式进行资源编排
- **事件驱动**: 基于Kubernetes Informer的事件驱动架构
- **构建器模式**: 统一的资源构建接口和策略
- **Webhook保护**: 通过Admission Webhook进行资源验证和默认设置

### 核心组件
1. **Manager**: 主控制器，负责协调所有CRD的reconcile循环
2. **Webhook**: 独立的验证服务，确保资源约束和引用完整性
3. **Controller层**: 各CRD的reconcile逻辑实现
4. **Builder层**: 资源构建和配置合并策略
5. **ClientMap**: Slurm客户端连接管理
6. **工具库**: 通用功能模块（对象同步、引用解析等）

### 核心功能特性

#### 高级功能
- **自动扩缩容**: 支持 KEDA 和 HPA，基于 Slurm ���标动态调整节点数量
- **混合部署**: 支持部分组件在 Kubernetes 外部署，兼容传统架构
- **工作负载隔离**: 使用 Taints、Tolerations 和 Pod Anti-Affinity 确保资源隔离
- **Scale to Zero**: 支持节点缩容到零，节省资源成本

#### 容器化支持
- **Pyxis 集成**: 支持 NVIDIA GPU 容器化作业
- **特权模式**: 支持 enroot 容器化作业运行
- **SPANK 插件**: 集成 slurm-spank-enroot 实现容器化

#### 企业级特性
- **认证集成**: 支持 SSSD 用户身份管理
- **计费系统**: 集成 MariaDB 进行作业计费
- **监控指标**: Prometheus 指标收集和 Grafana 可视化
- **高可用**: 多副本部署，故障自愈

## 顶层目录表

| 目录 | 作用 | 关键文件 |
|------|------|----------|
| **api/v1alpha1/** | CRD定义和类型定义 | `*_types.go`, `*_keys.go`, `well_known.go` |
| **cmd/** | 程序入口点 | `manager/main.go`, `webhook/main.go` |
| **config/** | 生成的Kubernetes配置 | `crds.yaml`, `rbac.yaml`, `manager.yaml` |
| **internal/** | 内部实现代码 | `controller/`, `builder/`, `webhook/`, `utils/` |
| **helm/** | Helm chart打包 | `slurm-operator-crds/`, `slurm-operator/`, `slurm/` |
| **hack/** | 开发工具脚本 | `build.sh`, `test.sh`, `generate.sh` |
| **docs/** | 项目文档 | API文档、架构文档、用户指南 |

### 关键入口文件

- **`cmd/manager/main.go`**: Manager服务入口，启动控制器管理器
- **`cmd/webhook/main.go`**: Webhook服务入口，启动Admission服务器
- **`api/v1alpha1/*_types.go`**: CRD类型定义，定义资源规范
- **`internal/controller/*_reconciler.go`**: 各CRD的控制器实现

## 启动流程图

```mermaid
graph TD
    A[启动命令] --> B{选择服务}
    B --> C[Manager服务]
    B --> D[Webhook服务]

    C --> E[解析配置]
    E --> F[创建Scheme]
    F --> G[注册CRD]
    G --> H[创建Manager]
    H --> I[启动Webhook]
    I --> J[启动控制器]
    J --> K[启动Informer]
    K --> L[启动Informers]
    L --> M[开始事件循环]
    M --> N[Ready信号]

    D --> O[创建Scheme]
    O --> P[注册Webhook]
    P --> Q[启动Server]
    Q --> R[配置健康检查]
    R --> S[Ready信号]

    style N fill:#90EE90
    style S fill:#90EE90
```

### Manager启动详细流程

```mermaid
graph LR
    A[main.go] --> B[配置管理器]
    B --> C[创建Scheme]
    C --> D[注册所有CRD]
    D --> E[创建Manager]
    E --> F[添加Webhook]
    F --> G[添加控制器]
    G --> H[设置指标]
    H --> I[设置健康检查]
    I --> J[启动管理器]
    J --> K[开始协调循环]
```

## 核心调���链时序图

### 1. 资源创建协调流程

```mermaid
sequenceDiagram
    participant K8sAPI as Kubernetes API
    participant Informer as Informer Cache
    participant Controller as ControllerReconciler
    participant Builder as Builder
    participant ClientMap as ClientMap
    participant Slurm as Slurm Service

    K8sAPI->>Informer: 创建CRD资源
    Informer->>Controller: 触发Reconcile事件
    Controller->>Controller: Reconcile()
    Controller->>Builder: BuildNodeSetApp()
    Builder->>ClientMap: 获取客户端连接
    ClientMap->>Slurm: 查询集群状态
    Slurm->>ClientMap: 返回状态信息
    ClientMap->>Builder: 构建StatefulSet
    Builder->>Controller: 返回构建的资源
    Controller->>K8sAPI: 创建/更新资源
    K8sAPI->>Informer: 更新资源状态
    Informer->>Controller: 触发状态更新
    Controller->>Controller: 更新CRD状态
```

### 2. 节点生命周期管理流程

```mermaid
sequenceDiagram
    participant NS as NodeSetReconciler
    participant PC as PodControl
    participant CS as ClientMap
    participant Slurm as Slurm Service
    participant K8s as Kubernetes API

    NS->>NS: syncNodeSetStatus()
    NS->>CS: 获取Slurm客户端
    CS->>Slurm: 查询节点状态
    Slurm->>CS: 节点状态列表
    CS->>NS: 返回状态信息
    NS->>NS: 更新状态条件
    NS->>PC: 驱动Pod变更
    PC->>K8s: 创建/删除Pod
    K8s->>PC: 返回结果
    PC->>NS: 同步Pod预期
    NS->>NS: 继续协调
```

### 3. Webhook验证流程

```mermaid
sequenceDiagram
    participant K8sAPI as Kubernetes API Server
    participant Webhook as Webhook Server
    participant ControllerWebhook as Controller Webhook
    participant Validation as Validation Logic

    K8sAPI->>Webhook: 转发Admission请求
    Webhook->>ControllerWebhook: 调用ValidateCreate/Update
    ControllerWebhook->>Validation: 检查字段约束
    Validation->>ControllerWebhook: 返回验证结果
    ControllerWebhook->>Webhook: 返回AdmissionResponse
    Webhook->>K8sAPI: 返回验证结果
    K8sAPI->>K8sAPI: 执行或拒绝操作
```

## 模块依赖关系图

```mermaid
graph TD
    subgraph "上层入口"
        A[cmd/manager/main.go]
        B[cmd/webhook/main.go]
    end

    subgraph "核心服务"
        C[Manager]
        D[Webhook]
    end

    subgraph "控制层"
        E[ControllerReconciler]
        F[NodeSetReconciler]
        G[LoginSetReconciler]
        H[AccountingReconciler]
        I[RestapiReconciler]
        J[TokenReconciler]
    end

    subgraph "构建层"
        K[ControllerBuilder]
        L[NodeSetBuilder]
        M[LoginSetBuilder]
        N[AccountingBuilder]
        O[RestApiBuilder]
    end

    subgraph "客户端层"
        P[ClientMap]
        Q[SlurmClient]
    end

    subgraph "工具层"
        R[objectutils]
        S[refresolver]
        T[podcontrol]
        U[historycontrol]
        V[durationstore]
    end

    subgraph "数据层"
        W[api/v1alpha1]
        X[Kubernetes API]
    end

    A --> C
    B --> D
    C --> E
    C --> F
    C --> G
    C --> H
    C --> I
    C --> J
    E --> K
    F --> L
    G --> M
    H --> N
    I --> O
    E --> P
    F --> P
    G --> P
    H --> P
    I --> P
    J --> P
    K --> R
    L --> R
    M --> R
    N --> R
    O --> R
    K --> S
    L --> S
    M --> S
    N --> S
    O --> S
    E --> T
    F --> T
    E --> U
    F --> U
    E --> V
    F --> V
    G --> V
    W --> E
    W --> F
    W --> G
    W --> H
    W --> I
    W --> J
    K --> X
    L --> X
    M --> X
    N --> X
    O --> X
    P --> Q
    Q --> X

    style C fill:#FFB6C1
    style D fill:#87CEEB
    style E fill:#98FB98
    style K fill:#DDA0DD
    style P fill:#F0E68C
```

## 外部依赖

### 1. Kubernetes API
- **作用**: 集群状态管理、资源创建/更新/删除
- **接口**: Client-go库提供Kubernetes客户端
- **资源**: StatefulSet, Deployment, Service, ConfigMap, Secret, Pod等
- **版本**: Kubernetes v1.29+

### 2. Slurm服务
- **作用**: Slurm集群状态查询、作业提交、节点管理
- **协议**: gRPC/REST API
- **版本**: Slurm 25.05+
- **组件**:
  - slurmctld (控制器守护进程)
  - slurmd (节点守护进程)
  - slurmdbd (数据库守护进程)
  - slurmrestd (REST API服务)

### 3. 数据库
- **作用**: Slurm持久化数据和记账
- **类型**: PostgreSQL/MySQL (通过slurmdbd), MariaDB (推荐)
- **功能**: 用户信息、作业历史、集群统计、计费数据
- **依赖**: 需要部署外部数据库服务或使用数据库 Helm chart

### 4. 认证系统
- **作用**: Slurm集群安全认证
- **组件**:
  - Munge key (共享密钥认证)
  - JWT (JSON Web Token认证)
  - SSSD (System Security Services Daemon)
  - LDAP/Active Directory (可选)

### 5. 自动扩缩容组件
- **KEDA**: Kubernetes Event-driven Autoscaling
- **HPA**: Horizontal Pod Autoscaler
- **Prometheus**: 指标收集和存储
- **Metrics Server**: Kubernetes资源指标API

### 6. 配置管理
- **作用**: 系统配置和密钥管理
- **来源**: Kubernetes ConfigMap/Secret
- **内容**: Slurm配置文件、数据库凭证、JWT密钥
- **支持**: 运行时配置覆盖

### 7. 监控和日志
- **作用**: 系统健康监控和调试
- **协议**: Prometheus Metrics, Structured Logs
- **端点**: /metrics, /healthz
- **组件**: slurm-exporter, Prometheus, Grafana

### 8. 开发工具链
- **KIND**: Kubernetes in Docker，本地测试集群
- **Docker**: 容器构建和运行
- **Helm**: 应用打包和部署
- **Skaffold**: 本地开发自动化
- **kubectl**: Kubernetes命令行工具
- **Pre-commit**: 代码质量检查钩子

## 配置项

### 1. Manager配置 (manager.yaml)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: slurm-operator-config
data:
  # 控制器配置
  controller-workers: "8"
  controller-requeue-after: "30s"

  # 客户端配置
  client-timeout: "10s"
  client-retry-interval: "5s"
  client-max-retries: "3"

  # 构建器配置
  builder-image-pull-policy: "IfNotPresent"
  builder-default-annotations: |
    slurm.slinky.net/operator-version: v1.0.0

  # Webhook配置
  webhook-port: 9443
  webhook-tls-file: "/etc/webhook/tls.crt"
```

### 2. Helm Values (values.yaml)

```yaml
# 全局配置
global:
  image:
    repository: "registry.example.com/slurm-operator"
    tag: "latest"
    pullPolicy: "IfNotPresent"

  # 监控配置
  metrics:
    enabled: true
    port: 8080
    serviceMonitor:
      enabled: true

  # 资源限制
  resources:
    limits:
      cpu: "500m"
      memory: "512Mi"
    requests:
      cpu: "100m"
      memory: "128Mi"

# CRD配置
crds:
  # 控制器资源限制
  controller:
    resources:
      limits:
        cpu: "1000m"
        memory: "2Gi"
      requests:
        cpu: "200m"
        memory: "256Mi"

# Webhook配置
webhook:
  resources:
    limits:
      cpu: "200m"
      memory: "256Mi"
    requests:
      cpu: "50m"
      memory: "64Mi"

# RBAC配置
rbac:
  create: true
  serviceAccount:
    create: true
    name: "slurm-operator"
```

### 3. 环境变量

| 环境变量 | 类型 | 默认值 | 描述 |
|----------|------|--------|------|
| `WATCH_NAMESPACE` | string | "" | 监控的命名空间 (空表示所有) |
| `LEADER_ELECTION_ENABLED` | bool | true | 启用领导者选举 |
| `LEADER_ELECTION_ID` | string | "slurm-operator.slurm.slinky.net" | 领导者选举ID |
| `METRICS_BIND_ADDRESS` | string | ":8080" | 监控指标绑定地址 |
| `HEALTH_BIND_ADDRESS` | string | ":9440" | 健康检查绑定地址 |
| `WEBHOOK_PORT` | int | 9443 | Webhook服务端口 |
| `WEBHOOK_TLS_FILE` | string | "" | Webhook TLS证书文件路径 |

### 4. Slurm配置模板

```yaml
# slurm.conf模板
slurm.conf: |
  # Slurm全局配置
  ClusterName={{ .ClusterName }}
  ControlMachine={{ .ControllerName }}.{{ .Namespace }}.svc.cluster.local
  SlurmUser=slurm
  SlurmctldPort=6817
  SlurmdPort=6818

  # Cgroup配置 (Slurm 25.05+)
  CgroupPlugin=cgroup/v2
  SlurmctldParameters=enable_configless

  # 网络配置
  ProctrackType=cgroup
  TaskPlugin=task/cgroup
  ReturnToService=1
  MaxJobCount=10000

  # 记账配置
  JobAcctGatherType=jobacct/task
  JobAcctGatherFrequency=30
```

## 架构设计原则

### 1. 单一职责原则
- 每个控制器专注于特定CRD的管理
- 构建器专门负责资源构建逻辑
- 工具库提供通用功能

### 2. 依赖注入原则
- 通过构造函数注入依赖
- 接口抽象便于测试和替换实现

### 3. 事件驱动原则
- 基于Kubernetes Informer的事件响应
- 异步处理避免阻塞主线程

### 4. 可观测性原则
- 内置Prometheus指标
- 结构化日志记录
- 健康检查端点

### 5. 容错性原则
- 重试机制处理临时故障
- 优雅关闭处理资源释放
- 错误状态记录和恢复

这种架构设计确保了系统的可扩展性、可维护性和可靠性，能够高效管理大规模Slurm HPC集群的部署和运维。