# ChangeLog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Added NodeSet level authcred configuration.
- Added topology.yaml to config files

### Fixed

- Fixed token job `ttlSecondsAfterFinished` being too low for helm
  `--wait-for-jobs`.
- Fixed nodeset pod's sackd image tag default value.
- Fixed webhook timeout being read from unintended values path.
- Fixed operator replicas being read form unintended values path.

### Changed

- Changed default storageClassName to empty.

### Removed
