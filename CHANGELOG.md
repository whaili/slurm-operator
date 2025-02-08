# ChangeLog

All notable changes to this project will be documented in this file.

## \[Unreleased\]

### Added

- Added container image labels.

### Fixed

- Fixed HTTP/2 enabled by default. [CVE-2023-44487] [CVE-2023-39325]
- Fixed Slurm helm chart using incorrect imagePullPolicy in values file.

### Changed

- Changed Slurm images to new schema.

### Removed

<!-- Links -->

[CVE-2023-44487]: https://github.com/advisories/GHSA-qppj-fm5r-hxr3
[CVE-2023-39325]: https://github.com/advisories/GHSA-4374-p667-p6c8
