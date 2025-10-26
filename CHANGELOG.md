# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


## [0.1.5] - 2025-10-26

### Added
- Release version 0.1.5

### Changed
- Updated gostratum dependencies to latest versions


## [0.1.4] - 2025-10-26

### Fixed

- Update gostratum/core dependency from v0.1.7 to v0.1.8

### Added

- Add Sanitize and ConfigSummary methods to metrics Config

## [0.1.3] - 2025-10-26

### Fixed

- Update gostratum/core dependency from v0.1.5 to v0.1.7

## [0.1.2] - 2025-10-26

### Fixed

- Update gostratum/core dependency from v0.1.4 to v0.1.5

### Changed

- Add .gitignore to exclude coverage.out file

## [0.1.1] - 2025-10-26

### Fixed

- Update module name from "metricsx" to "metrics"

### Changed

- Update go.mod dependencies and refactor test logger implementation

### Fixed

- Update assertions for empty labels and noop timer duration

## [0.1.0] - 2025-10-26

### Added

- Implement metrics interface and Prometheus provider