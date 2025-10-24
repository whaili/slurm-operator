# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Kubernetes operator for managing Slurm HPC clusters. It implements custom resource definitions (CRDs) and controllers to deploy, configure, and manage Slurm components (slurmctld, slurmd, slurmdbd, slurmrestd) as Kubernetes workloads.

- **Framework**: Kubebuilder v4 with controller-runtime
- **Language**: Go 1.24
- **Minimum Kubernetes**: v1.29
- **Minimum Slurm**: 25.05
- **CRD Domain**: `slinky.slurm.net`

## Common Commands

### Building and Testing

```bash
# Run tests (requires 67% code coverage)
make test

# Run a single test
KUBEBUILDER_ASSETS="$(./bin/setup-envtest-* use 1.29 --bin-dir ./bin -p path)" \
  go test -v ./internal/controller/nodeset -run TestNodeSetReconciler

# Build container images and Helm charts
make build

# Format, vet, and lint code
make fmt
make vet
make golangci-lint

# Generate CRDs and deep copy methods
make manifests
make generate
```

### Development Tools

```bash
# Install development binaries (dlv, kind, cloud-provider-kind)
make install-dev

# Validate Helm charts
make helm-validate

# Update Helm dependencies
make helm-dependency-update

# Generate Helm documentation
make helm-docs

# Create values-dev.yaml files for local development
make values-dev
```

### Running Locally

The operator consists of two separate binaries:
- Manager (controller reconciliation loops): `cmd/manager/main.go`
- Webhook (admission validation): `cmd/webhook/main.go`

## Architecture Overview

### Custom Resource Definitions (CRDs)

Six primary CRDs form the API (`api/v1alpha1/`):

1. **Controller** (`controller_types.go`) - Manages the Slurm controller daemon (slurmctld)
   - Central orchestration resource for a Slurm cluster
   - Creates StatefulSet, Service, and ConfigMap for slurmctld
   - Key field: `ClusterName` (immutable after creation)
   - References: `SlurmKeyRef`, `JwtHs256KeyRef` for auth, optional `AccountingRef`

2. **NodeSet** (`nodeset_types.go`) - Manages Slurm worker nodes (slurmd)
   - Creates StatefulSet with ordered pods for compute nodes
   - Most complex controller - handles pod lifecycle with Slurm state awareness
   - Implements node draining, revision history, and hostname-based resolution
   - References: `ControllerRef` (required)

3. **LoginSet** (`loginset_types.go`) - Manages Slurm login/submit nodes
   - Creates Deployment for user-facing SSH login nodes
   - Integrates with SSSD for user identity management
   - References: `ControllerRef` (required), `SssdConfRef` (optional)

4. **Accounting** (`accounting_types.go`) - Manages Slurm accounting daemon (slurmdbd)
   - Creates StatefulSet for slurmdbd with database configuration
   - One Accounting resource can serve multiple Controller resources
   - Key field: `StorageConfig` with database credentials

5. **RestApi** (`restapi_types.go`) - Manages Slurm REST API server (slurmrestd)
   - Creates Deployment for HTTP REST interface
   - References: `ControllerRef` (required)

6. **Token** (`token_types.go`) - Generates JWT tokens for Slurm authentication
   - Utility resource that creates/manages JWT tokens in Secrets
   - Supports automatic refresh
   - References: `JwtHs256KeyRef` (required)

**Important patterns in CRD definitions:**
- Each `*_types.go` file defines the CRD spec and status
- Corresponding `*_keys.go` file has helper methods (`Key()`, `ServiceFQDN()`, etc.)
- `well_known.go` defines standard annotations and labels
- `base_types.go` contains shared types like `ObjectReference`, `PodTemplate`, etc.

### Controller Architecture

Controllers follow a consistent reconciliation pattern (`internal/controller/`):

```
Reconcile() → Sync() → syncStatus()
                ↓
         Sequential SyncSteps
                ↓
         Builder.Build*()
                ↓
         objectutils.SyncObject()
```

**Key controllers:**
- **ControllerReconciler** (`controller/`) - Reconciles Controller CRD, creates slurmctld StatefulSet
- **NodeSetReconciler** (`nodeset/`) - Most complex, manages worker pod lifecycle with Slurm awareness
- **LoginSetReconciler** (`loginset/`) - Reconciles LoginSet CRD
- **AccountingReconciler** (`accounting/`) - Reconciles Accounting CRD
- **RestapiReconciler** (`restapi/`) - Reconciles RestApi CRD
- **TokenReconciler** (`token/`) - Generates JWT tokens
- **SlurmClientReconciler** (`slurmclient/`) - Internal controller for managing Slurm client connections via ClientMap

**Important controller utilities:**
- `clientmap.ClientMap` (`internal/clientmap/`) - Thread-safe map of Slurm client connections, keyed by Controller resource
- `DurationStore` (`internal/utils/durationstore/`) - Tracks requeue durations across reconciliation steps
- `PodControl` (`internal/utils/podcontrol/`) - Interface for pod creation/deletion with expectations tracking
- `HistoryControl` (`internal/utils/historycontrol/`) - Manages ControllerRevision objects for rollout history

### Builder Pattern

The `Builder` struct (`internal/builder/builder.go`) constructs Kubernetes resources from CRDs:
- Each CRD type has corresponding `Build*App()`, `Build*Service()`, `Build*Config()` methods
- Builders merge user-provided specs with operator-managed defaults
- Sub-packages: `labels/` for consistent labeling, `metadata/` for metadata management

**Core builder pattern:**
```go
builder.Build*() → PodTemplate → Container → Service/StatefulSet/Deployment
```

### Webhook Layer

Separate binary (`cmd/webhook/main.go`) implements validation and defaulting (`internal/webhook/v1alpha1/`):
- Each CRD has a webhook with `Default()`, `ValidateCreate()`, `ValidateUpdate()`, `ValidateDelete()`
- Validates immutability constraints (e.g., Controller.ClusterName cannot change)
- Validates cross-resource references exist

### Critical Utility Patterns

1. **SyncObject** (`internal/utils/objectutils/`) - Core pattern for creating/updating Kubernetes objects:
   - Get existing object → If not found: Create → If found: Strategic merge patch
   - Used throughout all controllers

2. **RefResolver** (`internal/utils/refresolver/`) - Resolves `ObjectReference` to actual Kubernetes resources
   - Handles namespace fallback
   - Used for dereferencing cross-CRD references

3. **Pod Expectation Tracking** (`NodeSetReconciler`) - Prevents reconciliation until pod creates/deletes complete
   - Uses `UIDTrackingControllerExpectations`
   - Critical for avoiding timing issues

## Important Implementation Details

### Immutable Fields
- `Controller.ClusterName` cannot change after creation
- Webhooks enforce immutability constraints

### Resource Relationships
```
Controller → NodeSet (1:N)
Controller → LoginSet (1:N)
Controller → RestApi (1:N)
Controller → Accounting (N:1 optional)
Token → JWT Secret (1:1)
```

### Namespace Handling
- `ObjectReference` fields include namespace
- If namespace empty, defaults to referencing object's namespace
- RefResolver handles fallback logic

### Naming Conventions
- CRD helper methods in `*_keys.go` encapsulate naming logic
- Use `Key()` for NamespacedName, `ServiceFQDN()` for DNS names
- Standard labels in `well_known.go`

### Configuration Management
- `ExtraConf` fields allow appending custom Slurm configuration
- `ConfigFileRefs`, `PrologScriptRefs`, `EpilogScriptRefs` reference ConfigMaps/Secrets
- Builder layer merges base configs with user-provided extras

### Pod Lifecycle for NodeSets
- NodeSet pods get predictable hostnames for direct communication
- Controller queries Slurm state before terminating pods
- Pods marked as "drain" in Slurm before deletion
- Pod conditions reflect Slurm node state

### Status Conditions
- All CRDs use `metav1.Condition` for status tracking
- Controllers update conditions to indicate reconciliation progress/errors
- Follow standard Kubernetes Conditions API

### Testing
- Tests use envtest (Kubernetes API server without kubelet)
- Coverage threshold enforced at 67%
- Suite tests per controller in `suite_test.go`
- Test helpers in `internal/utils/testutils/`

## Codebase Organization

```
api/v1alpha1/          # CRD definitions and helper methods
cmd/
  manager/             # Controller manager entry point
  webhook/             # Admission webhook entry point
config/                # Generated CRD and RBAC manifests
internal/
  builder/             # Resource construction (StatefulSets, Services, etc.)
  clientmap/           # Slurm client connection management
  controller/          # Reconciliation logic for each CRD
  utils/               # Shared utilities (objectutils, refresolver, etc.)
  webhook/             # Validation and defaulting webhooks
helm/                  # Helm charts
  slurm-operator-crds/ # CRD installation chart
  slurm-operator/      # Operator deployment chart
  slurm/               # Reference Slurm cluster chart
hack/                  # Development scripts and resources
docs/                  # Project documentation
```

## Notes on Slurm Integration

- The operator manages Kubernetes resources, not Slurm directly
- Container images run actual Slurm daemons (slurmctld, slurmd, slurmdbd, slurmrestd)
- `ClientMap` manages gRPC/REST connections to slurmctld for querying cluster state
- Slurm configuration generated from CRD specs and mounted as ConfigMaps
- Authentication via shared `SlurmKeyRef` (munge key) and `JwtHs256KeyRef` (JWT signing key)

## Cgroup Constraints

Slurm 25.05+ requires cgroup v2. Ensure:
- `CgroupPlugin=cgroup/v2` in slurm.conf
- `SlurmctldParameters=enable_configless` for dynamic node registration


## Code Architecture
The detailed project understanding documents are organized into five parts under the `docs/claude/` directory:  
- 01-overview.md – Project Overview (directories, responsibilities, build/run methods, external dependencies, newcomer reading order)  
- 02-entrypoint.md – Program Entry & Startup Flow (entry functions, CLI commands, initialization and startup sequence)  
- 03-callchains.md – Core Call Chains (function call tree, key logic explanations, main sequence diagram)  
- 04-modules.md – Module Dependencies & Data Flow (module relationships, data structures, request/response processing, APIs)  
- 05-architecture.md – System Architecture (overall structure, startup flow, key call chains, module dependencies, external systems, configuration)  
When answering any questions related to source code structure, module relationships, or execution flow, **always refer to these five documents first**, and include file paths and function names for clarity.

## Reply Guidelines
- Always reference **file path + function name** when explaining code.
- Use **Mermaid diagrams** for flows, call chains, and module dependencies.
- If context is missing, ask explicitly which files to `/add`.
- Never hallucinate non-existing functions or files.
- Always reply in **Chinese**

## Excluded Paths
- vendor/
- build/
- dist/
- .git/
- third_party/

