# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.7.0] - 2021-10-29

### Changed

- version bump to make this compliant with our new sem ver standards

## [1.6.6] - 2021-09-13

### Changed

- Changed the docker image to run as the user nobody

## [1.6.5] - 2021-08-12

### Changed

- Fixed the chart target in the Makefile

## [1.6.4] - 2021-08-10

### Changed

- cleaned up dockerfiles, added .github files


## [1.6.3] - 2021-07-26

### Changed

- Migrated to github


## [1.6.2] - 2021-07-22

### Changed

- Added Jenkins pipeline configuration for GitHub

## [1.6.1] - 2021-06-30


### Security

- CASMHMS-4898 - Updated base container images for security updates.

## [1.6.0] - 2021-06-18

### Changed
- Bump minor version for CSM 1.2 release branch



## [1.5.0] - 2021-06-18

### Changed
- Bump minor version for CSM 1.1 release branch

## [1.4.3] - 2021-05-05

### Changed

- Updated docker-compose files to pull images from Artifactory instead of DTR.

## [1.4.2] - 2021-04-20

### Changed

- Updated Dockerfile to pull base images from Artifactory instead of DTR.

## [1.4.1] - 2021-03-09

### Changed

- CASMHMS-4329 - Updated Azure go-autorest vendor package from v0.11.10 to v0.11.18 for security updates.

## [1.4.0] - 2021-02-02

### Changed

- Added license headers to code.

## [1.3.0] - 2021-01-14

### Changed

- Updated license file.

## [1.2.4] - 2020-11-04

### Security

- CASMHMS-4109 - Updated Azure go-autorest vendor package for security updates.

## [1.2.3] - 2020-11-02

### Security

- CASMHMS-4148 - Updated Go module vendor code for security update.

## [1.2.2] - 2020-10-21

### Security

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.2.1] - 2020-09-10

### Security

- CASMHMS-3998 - More trusted baseOS updates for hms-trs-operator.

## [1.2.0] - 2020-08-12

### Changed

- CASMHMS-3350 - Updated hms-trs-operator to use the latest trusted baseOS images.

## [1.1.0] - 2020-06-03

### Changed

- Support for online upgrade and rollbacks with the TRS operator helm chart.

## [1.0.2] - 2020-03-01

### Fixed

- Added product field to Jenkins file.

## [1.0.1] - 2020-02-26

### Added

- RBAC for Kafka topic watcher service account so services can block on topic creation and not race the operator.

## [1.0.0] - 2020-02-26

### Added

- First version.

### Changed

### Deprecated

### Removed

### Fixed

### Security

