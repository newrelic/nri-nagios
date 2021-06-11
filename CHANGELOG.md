# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

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
