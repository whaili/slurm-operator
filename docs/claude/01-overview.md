# 01 - 项目概览

## 项目基本信息

**项目名称**: Slurm Operator for Kubernetes
**项目类型**: Kubernetes Operator (Custom Resource Definitions + Controllers)
**开发框架**: Kubebuilder v4 with controller-runtime
**编程语言**: Go 1.24
**最低 Kubernetes 版本**: v1.29
**最低 Slurm 版本**: 25.05
**CRD Domain**: `slinky.slurm.net`

## 目录结构与职责

| 目录 | 主要职责 | 关键文件 |
|------|---------|---------|
| `api/v1alpha1/` | CRD 定义、类型定义和辅助方法 | `controller_types.go`, `nodeset_types.go`, `loginset_types.go`, `accounting_types.go`, `restapi_types.go`, `token_types.go`, `*_keys.go`, `well_known.go`, `base_types.go` |
| `cmd/manager/` | 控制器管理器入口点 | `main.go` |
| `cmd/webhook/` | Admission Webhook 入口点 | `main.go` |
| `config/` | 生成的 CRD 和 RBAC 清单 | CRD YAML 文件、RBAC 配置 |
| `internal/builder/` | Kubernetes 资源构造器(StatefulSet、Service 等) | `builder.go`, `Build*()` 方法, `labels/`, `metadata/` 子包 |
| `internal/clientmap/` | Slurm 客户端连接管理 | `ClientMap` 实现 |
| `internal/controller/` | 各 CRD 的协调逻辑 | `controller/`, `nodeset/`, `loginset/`, `accounting/`, `restapi/`, `token/`, `slurmclient/` |
| `internal/utils/` | 共享工具函数 | `objectutils/` (SyncObject), `refresolver/` (RefResolver), `podcontrol/`, `historycontrol/`, `durationstore/`, `testutils/` |
| `internal/webhook/v1alpha1/` | Validation 和 Defaulting Webhooks | 每个 CRD 的 webhook 实现 |
| `helm/` | Helm Charts | `slurm-operator-crds/`, `slurm-operator/`, `slurm/` |
| `hack/` | 开发脚本和资源 | 辅助脚本 |
| `docs/` | 项目文档 | 各类文档 |

## 构建与运行方式

### 常用构建命令

```bash
# 运行测试 (需要 67% 代码覆盖率)
make test

# 运行单个测试
KUBEBUILDER_ASSETS="$(./bin/setup-envtest-* use 1.29 --bin-dir ./bin -p path)" \
  go test -v ./internal/controller/nodeset -run TestNodeSetReconciler

# 构建容器镜像和 Helm charts
make build

# 代码格式化、检查和 lint
make fmt
make vet
make golangci-lint

# 生成 CRDs 和 deep copy 方法
make manifests
make generate
```

### 开发工具安装

```bash
# 安装开发二进制文件 (dlv, kind, cloud-provider-kind)
make install-dev

# 验证 Helm charts
make helm-validate

# 更新 Helm 依赖
make helm-dependency-update

# 生成 Helm 文档
make helm-docs

# 为本地开发创建 values-dev.yaml 文件
make values-dev
```

### 本地运行

Operator 包含两个独立的二进制文件:

| 二进制文件 | 入口文件 | 职责 |
|-----------|---------|------|
| Manager | `cmd/manager/main.go` | 控制器协调循环 |
| Webhook | `cmd/webhook/main.go` | Admission 验证 |

可以通过 `go run` 直接运行:
```bash
go run cmd/manager/main.go
go run cmd/webhook/main.go
```

## 外部依赖

### Kubernetes 生态系统
- **controller-runtime**: Kubernetes controller 框架
- **Kubebuilder v4**: Operator 脚手架和代码生成
- **envtest**: 测试框架 (不含 kubelet 的 Kubernetes API server)

### Slurm 集成
- **Slurm Daemons**: slurmctld, slurmd, slurmdbd, slurmrestd (运行在容器中)
- **Slurm Client**: gRPC/REST 连接到 slurmctld 查询集群状态
- **Authentication**:
  - Munge key (`SlurmKeyRef`)
  - JWT HS256 key (`JwtHs256KeyRef`)

### 数据库 (可选)
- **slurmdbd**: 通过 `Accounting` CRD 的 `StorageConfig` 配置数据库凭据

### 其他依赖
- **SSSD**: 用于 LoginSet 的用户身份管理
- **Cgroup v2**: Slurm 25.05+ 要求 cgroup v2

## 核心概念

### 六大 CRD 资源

| CRD | 文件 | 管理对象 | 引用关系 |
|-----|------|---------|---------|
| **Controller** | `controller_types.go` | slurmctld (控制器守护进程) | SlurmKeyRef, JwtHs256KeyRef, AccountingRef (可选) |
| **NodeSet** | `nodeset_types.go` | slurmd (计算节点) | ControllerRef (必需) |
| **LoginSet** | `loginset_types.go` | 登录/提交节点 | ControllerRef (必需), SssdConfRef (可选) |
| **Accounting** | `accounting_types.go` | slurmdbd (计费守护进程) | StorageConfig (数据库配置) |
| **RestApi** | `restapi_types.go` | slurmrestd (REST API 服务器) | ControllerRef (必需) |
| **Token** | `token_types.go` | JWT token 生成 | JwtHs256KeyRef (必需) |

### 资源关系图

```
Controller (1) ──→ (N) NodeSet
           (1) ──→ (N) LoginSet
           (1) ──→ (N) RestApi
           (N) ──→ (1) Accounting (可选)

Token (1) ──→ (1) JWT Secret
```

## 建议的新手阅读顺序

### 第一阶段：理解 API 定义 (api/v1alpha1/)

1. **基础类型**: `base_types.go` - 了解共享类型如 `ObjectReference`, `PodTemplate`
2. **通用标签**: `well_known.go` - 理解标准注解和标签
3. **Controller CRD**: `controller_types.go` + `controller_keys.go` - 最核心的 CRD
4. **NodeSet CRD**: `nodeset_types.go` + `nodeset_keys.go` - 最复杂的 CRD
5. **其他 CRDs**: 按需阅读 `loginset_types.go`, `accounting_types.go`, `restapi_types.go`, `token_types.go`

### 第二阶段：理解控制器模式 (internal/controller/)

6. **简单控制器**: `controller/` - 从最直接的 Controller 协调器开始
7. **复杂控制器**: `nodeset/` - 理解 pod 生命周期管理、Slurm 状态感知
8. **工具控制器**: `token/` - 理解 JWT token 生成逻辑
9. **客户端管理**: `slurmclient/` - 理解如何管理 Slurm 客户端连接

### 第三阶段：理解资源构建 (internal/builder/)

10. **Builder 模式**: `builder.go` - 理解如何从 CRD 构造 Kubernetes 资源
11. **子包**: `labels/`, `metadata/` - 理解一致性标签和元数据管理

### 第四阶段：理解 Webhook 层 (internal/webhook/v1alpha1/)

12. **Webhook 基础**: 选择一个 CRD 的 webhook 实现，理解 `Default()`, `ValidateCreate()`, `ValidateUpdate()`, `ValidateDelete()` 方法

### 第五阶段：理解工具函数 (internal/utils/)

13. **SyncObject**: `objectutils/` - 核心的创建/更新 Kubernetes 对象模式
14. **RefResolver**: `refresolver/` - 解析 `ObjectReference` 到实际 Kubernetes 资源
15. **其他工具**: `podcontrol/`, `historycontrol/`, `durationstore/`

### 第六阶段：理解入口点

16. **Manager**: `cmd/manager/main.go` - 控制器管理器启动流程
17. **Webhook**: `cmd/webhook/main.go` - Webhook 服务器启动流程

### 第七阶段：测试和实践

18. **测试框架**: `internal/utils/testutils/` - 理解测试辅助函数
19. **运行测试**: 使用 `make test` 运行测试套件
20. **本地开发**: 使用 `make install-dev` 和 `make values-dev` 设置本地环境

## 关键设计模式

### 控制器协调模式
```
Reconcile() → Sync() → syncStatus()
                ↓
         Sequential SyncSteps
                ↓
         Builder.Build*()
                ↓
         objectutils.SyncObject()
```

### CRD 定义模式
- 每个 `*_types.go` 文件定义 CRD spec 和 status
- 对应的 `*_keys.go` 文件包含辅助方法 (`Key()`, `ServiceFQDN()` 等)

### Builder 模式
```go
builder.Build*() → PodTemplate → Container → Service/StatefulSet/Deployment
```

## 重要约束

### 不可变字段
- `Controller.ClusterName` 创建后不可更改
- Webhooks 强制执行不可变约束

### 命名规范
- CRD 辅助方法在 `*_keys.go` 中封装命名逻辑
- 使用 `Key()` 获取 NamespacedName，使用 `ServiceFQDN()` 获取 DNS 名称
- 标准标签在 `well_known.go` 中定义

### Cgroup 约束
- Slurm 25.05+ 需要 cgroup v2
- 确保 `CgroupPlugin=cgroup/v2` 在 slurm.conf 中
- `SlurmctldParameters=enable_configless` 用于动态节点注册

## 测试要求

- 使用 envtest (不含 kubelet 的 Kubernetes API server)
- 强制执行 67% 代码覆盖率阈值
- 每个控制器在 `suite_test.go` 中有测试套件
- 测试辅助函数在 `internal/utils/testutils/` 中
