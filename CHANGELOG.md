# ChangeLog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Added NodeSet level authcred configuration.

### Fixed

- Fixed token job `ttlSecondsAfterFinished` being too low for helm
  `--wait-for-jobs`.
- Fixed nodeset pod's sackd image tag default value.

### Changed

- Changed default storageClassName to empty.

### Removed
