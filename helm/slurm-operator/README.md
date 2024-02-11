# slurm-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 24.05](https://img.shields.io/badge/AppVersion-24.05-informational?style=flat-square)

Helm Chart for Slurm HPC Workload Manager Operator

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| certManager.duration | string | `"43800h0m0s"` |  Duration of certificate life. |
| certManager.enabled | bool | `true` |  Enables cert-manager for certificate management. |
| certManager.renewBefore | string | `"8760h0m0s"` |  Certificate renewal time. Should be before the expiration. |
| certManager.secretName | string | `"slurm-operator-webhook-ca"` |  The secret to be (created and) mounted. |
| fullnameOverride | string | `""` |  Overrides the full name of the release. |
| image.repository | string | `"ghcr.io/slinkyproject/slurm-operator"` |  Sets the image repository to use. |
| image.tag | string | The Release appVersion. |  Sets the image tag to use. |
| imagePullPolicy | string | `"IfNotPresent"` |  Set the image pull policy. |
| imagePullSecrets | list | `[]` |  Sets the image pull secrets. Ref: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/ |
| nameOverride | string | `""` |  Overrides the name of the release. |
| namespaceOverride | string | `""` |  Overrides the namespace of the release. |
| operator.affinity | object | `{}` |  Set affinity for Kubernetes Pod scheduling. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity |
| operator.clusterWorkers | integer | `1` |  Set the max concurrent workers for the Cluster controller. |
| operator.enabled | bool | `true` |  Enables the operator. |
| operator.logLevel | string | `"info"` |  Set the log level by string (e.g. error, info, debug) or number (e.g. 1..5). |
| operator.nodesetWorkers | integer | `1` |  Set the max concurrent workers for the NodeSet controller. |
| operator.replicas | integer | `1` |  Set the number of replicas to deploy. |
| operator.resources | object | `{}` |  Set container resource requests and limits for Kubernetes Pod scheduling. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-requests-and-limits-of-pod-and-container |
| operator.serviceAccount.create | bool | `true` |  Allows chart to create the service account. |
| operator.serviceAccount.name | string | `""` |  Set the service account to use (and create). |
| priorityClassName | string | `""` |  Set the priority class to use. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass |
| webhook.affinity | object | `{}` |  Set affinity for Kubernetes Pod scheduling. Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity |
| webhook.enabled | bool | `true` |  Enables the webhook. |
| webhook.logLevel | string | `"info"` |  Set the log level by string (e.g. error, info, debug) or number (e.g. 1..5). |
| webhook.replicas | integer | `1` |  Set the number of replicas to deploy. |
| webhook.resources | object | `{}` |  Set container resource requests and limits for Kubernetes Pod scheduling. Ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-requests-and-limits-of-pod-and-container |
| webhook.serviceAccount.create | bool | `true` |  Allows chart to create the service account. |
| webhook.serviceAccount.name | string | `""` |  Set the service account to use (and create). |

