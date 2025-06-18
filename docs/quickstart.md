# QuickStart Guide

## Table of Contents

<!-- mdformat-toc start --slug=github --no-anchors --maxlevel=6 --minlevel=1 -->

- [QuickStart Guide](#quickstart-guide)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Install](#install)
    - [Optional](#optional)
    - [Slurm Operator](#slurm-operator)
    - [Slurm Cluster](#slurm-cluster)
    - [Testing](#testing)

<!-- mdformat-toc end -->

## Overview

This quickstart guide will help you get the slurm-operator running and deploy
Slurm clusters to Kubernetes.

## Install

### Optional

Install the optional helm charts.

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
# Install certificate manager for slurm-operator's webhook.
helm install cert-manager jetstack/cert-manager \
	--namespace cert-manager --create-namespace
# Install Prometheus stack for slurm-exporter.
helm install prometheus prometheus-community/kube-prometheus-stack \
	--namespace prometheus --create-namespace
```

### Slurm Operator

Download values and install the slurm-operator from OCI package.

```bash
curl -L https://raw.githubusercontent.com/SlinkyProject/slurm-operator/refs/tags/v0.4.0/helm/slurm-operator/values.yaml \
  -o values-operator.yaml
helm install slurm-operator oci://ghcr.io/slinkyproject/charts/slurm-operator \
  --values=values-operator.yaml --version=0.4.0 --namespace=slinky --create-namespace
```

If a cert-manager is not installed, then make sure to set
`certManager.enabled=false` in values or helm install with
`--set 'certManager.enabled=false'`.

```sh
helm install slurm-operator oci://ghcr.io/slinkyproject/charts/slurm-operator \
  --version=0.4.0 --set 'certManager.enabled=false' --namespace=slinky --create-namespace
```

Make sure the cluster deployed successfully with:

```sh
kubectl --namespace=slinky get pods
```

Output should be similar to:

```sh
NAME                                      READY   STATUS    RESTARTS   AGE
slurm-operator-7444c844d5-dpr5h           1/1     Running   0          5m00s
slurm-operator-webhook-6fd8d7857d-zcvqh   1/1     Running   0          5m00s
```

### Slurm Cluster

Download values and install a Slurm cluster from OCI package.

```bash
curl -L https://raw.githubusercontent.com/SlinkyProject/slurm-operator/refs/tags/v0.4.0/helm/slurm/values.yaml \
  -o values-slurm.yaml
helm install slurm oci://ghcr.io/slinkyproject/charts/slurm \
  --values=values-slurm.yaml --version=0.4.0 --namespace=slurm --create-namespace
```

Make sure the slurm cluster deployed successfully with:

```sh
kubectl --namespace=slurm get pods
```

Output should be similar to:

```sh
NAME                                  READY   STATUS    RESTARTS   AGE
slurm-accounting-0                    1/1     Running   0          2m
slurm-compute-slinky-0                2/2     Running   0          2m
slurm-controller-0                    3/3     Running   0          2m
slurm-exporter-6ffb9fdbbd-547zj       1/1     Running   0          2m
slurm-login-slinky-7ff66445b5-wdjkn   1/1     Running   0          2m
slurm-restapi-77b9f969f7-kh4r8        1/1     Running   0          2m
```

### Testing

SSH through the login service:

```sh
SLURM_LOGIN_IP="$(kubectl get services -n slurm slurm-login-slinky -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"
SLURM_LOGIN_PORT="$(kubectl get services -n slurm slurm-login-slinky -o jsonpath='{.status.loadBalancer.ingress[0].ports[0].port}')"
## Assuming your public SSH key was configured in `login.rootSshAuthorizedKeys`.
ssh -p ${SLURM_LOGIN_PORT:-22} root@${SLURM_LOGIN_IP}
## Assuming SSSD is configured.
ssh -p ${SLURM_LOGIN_PORT:-22} ${USER}@${SLURM_LOGIN_IP}
```

Then, from a login pod, run Slurm commands to quickly test that Slurm is
functioning:

```sh
sinfo
srun hostname
sbatch --wrap="sleep 60"
squeue
sacct
```

See [Slurm Commands][slurm-commands] for more details on how to interact with
Slurm.

<!-- Links -->

[slurm-commands]: https://slurm.schedmd.com/quickstart.html#commands
