# ChangeLog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Added NodeSet level authcred configuration.
- Added topology.yaml to config files
- Added Accounting, Controller, Restapi, and LoginSet CRDs.
- Added Slurm ClusterName override, otherwise derived from Controller CR
  metadata.
- Added disaggregated configuration for each sidecar, no longer overloading the
  authcred configuration.

### Fixed

- Fixed token job `ttlSecondsAfterFinished` being too low for helm
  `--wait-for-jobs`.
- Fixed nodeset pod's sackd image tag default value.
- Fixed webhook timeout being read from unintended values path.
- Fixed operator replicas being read form unintended values path.
- Fixed race condition where a stale NodeSet pod status leads to the Slurm node
  being terminated instead of drained.

### Changed

- Changed default storageClassName to empty.
- Changed to Slurm v43 API.
- Changed how a Slurm clusters are expressed via CRDs.
- Changed lifetime of JWT tokens created by operator from infinite to 15
  minutes.
- Changed how Slurm config files and secrets are set up in the pod, mount
  volumes with `securityContext.fsGroup` and remove initconf sidecar.
- Changed logfile sidecar image to alpine.
- Changed reconfigure sidecar image to slurmctld.

### Removed

- Removed the Cluster CRD.
- Removed `bitnami/mariadb` dependency from Slurm helm chart.
