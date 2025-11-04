# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

Unreleased section should follow [Release Toolkit](https://github.com/newrelic/release-toolkit#render-markdown-and-update-markdown)

## Unreleased

## v2.11.3 - 2025-11-04

### ⛓️ Dependencies
- Updated golang version to v1.25.3

## v2.11.2 - 2025-08-29

### ⛓️ Dependencies
- Updated golang patch version to v1.24.6

## v2.11.1 - 2025-07-01

### ⛓️ Dependencies
- Updated golang version to v1.24.4

## v2.11.0 - 2025-03-11

### 🚀 Enhancements
- Add FIPS compliant packages

### ⛓️ Dependencies
- Updated golang patch version to v1.23.6

## v2.10.2 - 2025-01-21

### ⛓️ Dependencies
- Updated golang patch version to v1.23.5

## v2.10.1 - 2024-11-18

### 🐞 Bug fixes
- Prevent forceful exit on invalid output commands

## v2.10.0 - 2024-10-15

### dependency
- Upgrade go to 1.23.2

### 🚀 Enhancements
- Upgrade integrations SDK so the interval is variable and allows intervals up to 5 minutes

## v2.9.7 - 2024-09-10

### ⛓️ Dependencies
- Updated golang version to v1.23.1

## v2.9.6 - 2024-07-09

### ⛓️ Dependencies
- Updated golang version to v1.22.5

## v2.9.5 - 2024-05-14

### ⛓️ Dependencies
- Updated golang version to v1.22.3

## v2.9.4 - 2024-04-16

### ⛓️ Dependencies
- Updated golang version to v1.22.2

## v2.9.3 - 2024-03-12

### 🐞 Bug fixes
- Updated golang to version v1.21.7 to fix a vulnerability

## v2.9.2 - 2024-02-27

### ⛓️ Dependencies
- Updated github.com/newrelic/infra-integrations-sdk to v3.8.2+incompatible

## v2.9.1 - 2023-10-31

### ⛓️ Dependencies
- Updated golang version to 1.21

## 2.9.0 (2023-06-06)
### Changed
- Upgrade Go version to 1.20

## 2.8.3  (2022-07-05)
### Changed
- Bump dependencies
### Added
Added support for more distributions:
- RHEL(EL) 9
- Ubuntu 22.04


## 2.8.2  (2022-06-07)
### Changed
- Bump Go to 1.18
- Bump dependencies

## 2.8.1 (2021-09-20)
### Changed
- Added windows `nagios-config.sample` file 

## 2.8.0 (2021-08-27)
### Added

Moved default config.sample to [V4](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-newer-configuration-format/), added a dependency for infra-agent version 1.20.0

Please notice that old [V3](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-standard-configuration-format/) configuration format is deprecated, but still supported.

## 2.7.1 (2021-06-10)
### Changed
- ARM support.

## 2.7.0 (2021-05-10)
### Changed
- Update Go to v1.16.
- Migrate to Go Modules
- Update Infrastracture SDK to v3.6.7.
- Update other dependecies.

## 2.6.1 (2021-03-25)
### Changed
- Release pipeline has been moved to Github Action
- Code of conduct has been removed
- Dependency yaml.v2 bumped to remedy a severity

## 2.6.0 (2020-01-28)
### Added
- `output_table_name` argument

## 2.5.0 (2020-01-14)
### Added
- `concurrency` argument

## 2.4.1 (2019-12-04)
### Changed
- Relaxed required permissions for the service checks file to 0640

## 2.3.0 (2019-11-22)
### Changed
- Renamed the integration executable from nr-nagios to nri-nagios in order to be consistent with the package naming. **Important Note:** if you have any security module rules (eg. SELinux), alerts or automation that depends on the name of this binary, these will have to be updated.

## 2.1.3 - 2019-10-16
### Fixed
- Windows installer GUIDs

## 2.1.2 - 2019-08-06
### Added
- Hostname as attribute

## 2.1.1 - 2019-07-30
### Added
- Windows build scripts for packaging

## 2.1.0 - 2019-06-21
### Added
- Optional best-effort metric parsing for service check output
- serviceCheck.serviceOutput and serviceCheck.longServiceOutput attributes

## 2.0.0 - 2019-04-25
### Changed
- Updated the SDK
- Added executing_host ID Attribute

## 1.0.1 - 2019-03-19
### Changed
- GA Release

## 0.1.4 - 2019-03-19
### Changed
- Add log line when unable to execute service checks

## 0.1.3 - 2019-01-08
### Changed
- Rename sample service checks file for consistency
- Change from sample check_yum to more standard check_ssh

## 0.1.2 - 2018-12-10
### Changed
- Add environment variables for packaging

## 0.1.1 - 2018-12-7
### Changed
- Fixed some test causing builds to fail

## 0.1.0 - 2018-11-15
### Added
- Initial version: Includes Metrics and Inventory data
