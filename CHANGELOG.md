# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->

[keep a changelog]: https://keepachangelog.com/en/1.0.0/
[semantic versioning]: https://semver.org/spec/v2.0.0.html

## [0.4.0]

### Added

- Added `env` key to Helm chart values

### Fixed

- Handle `null` values in `attributes` field

### Changed

- Controller now loads all values from the `proclaim` secret as environment variables
- **[BC]** Changed some Helm chart values for consistency:
  - Added `deployment.labels`
  - Renamed `deploymentAnnotations` to `deployment.annotations`
  - Renamed `podAnnotations` to `pod.annotations`
  - Renamed `podLabels` to `pod.labels`
  - Renamed `commonLabels` to `common.labels`

### Removed

- Removed various unused Helm chart values

## [0.3.0] - 2023-03-20

### Changed

- **[BC]** Changed `attributes` field to accept any scalar type
  - Empty string values are longer treated as "flags", use `true` instead
  - Regular associative attributes with empty values are now supported

## [0.2.0] - 2023-03-20

### Added

- The controller now verifies instances are advertised correctly using DNS-SD queries
- Added "conditions" to CRD status, as per Kubernetes API design recommendations
  - The `Adopted` condition indicates whether a suitable DNS provider has been found
  - The `Advertised` condition indicates whether the DNS records have been successfully created/updated
  - The `Discoverable` condition indicates whether the service is discoverable via DNS-SD
- Added more granular events
- Added `targets` field to CRD, allowing (future) support for multiple targets per instance

### Fixed

- Marked `DNSIMPLE_TOKEN` environment variable as "sensitive" to avoid showing its value in validation output

### Changed

- **[BC]** Moved DNS-SD properties in CRD into `instance` subkey of `spec`
- **[BC]** Renamed `service` fields in CRD to `serviceType` for better alignment with the DNS-SD spec
- **[BC]** Removed `targetHost` and `targetPort` fields from CRD, see new `targets` field instead

## [0.1.0] - 2023-03-16

- Initial release

<!-- references -->

[unreleased]: https://github.com/dogmatiq/proclaim
[0.1.0]: https://github.com/dogmatiq/proclaim/releases/tag/v0.1.0
[0.2.0]: https://github.com/dogmatiq/proclaim/releases/tag/v0.2.0
[0.3.0]: https://github.com/dogmatiq/proclaim/releases/tag/v0.3.0
[0.4.0]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.0

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
