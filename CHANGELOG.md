# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Removed
- **Config.Sanitize()** method removed - metrics config contains no secrets
  - This was implementing the method unnecessarily for "consistency"
  - Module now only implements `Prefix()` and `ConfigSummary()`
  - No breaking change for users who weren't calling `Sanitize()` directly

## [0.2.0] - 2025-10-29

### Added
- Release version 0.2.0
### Changed
- Updated gostratum dependencies to latest versions

### Fixed
- Fix: update go.mod / module metadata (from release)

## [0.1.5] - 2025-10-26

### Added
- Release version 0.1.5


## [0.1.4] - 2025-10-21

### Added
- Sanitize and ConfigSummary methods for `Config` to improve security and diagnostics.

### Fixed
- Update NewConfig to return sanitized metrics configuration
- Update gostratum/core dependency from v0.1.7 to v0.1.8

## [0.1.3] - 2025-10-20

### Fixed

- Update gostratum/core dependency from v0.1.5 to v0.1.7

## [0.1.2] - 2025-10-17

### Fixed

- Update gostratum/core dependency from v0.1.4 to v0.1.5

### Changed

- Add .gitignore to exclude coverage.out file
- Update module name from "metricsx" to "metrics"
- Update go.mod dependencies and refactor test logger implementation

## [0.1.1] - 2025-10-16

### Added

- Implement metrics interface and Prometheus provider

### Test

- Update assertions for empty labels and noop timer duration