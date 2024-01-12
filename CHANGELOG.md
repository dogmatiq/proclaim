# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->

[keep a changelog]: https://keepachangelog.com/en/1.0.0/
[semantic versioning]: https://semver.org/spec/v2.0.0.html

<!-- REMEMBER TO BUMP THE VERSIONS IN THE CHART FILE -->

## [0.4.9] - 2024-01-12

### Fixed

- Fixed "reconciler returned both a non-zero result and a non-nil error" warning
- Fixed "the object has been modified" error by always requeing on conflict

## [0.4.8] - 2023-11-16

### Added

- Added `common.annotations` value to Helm chart, which adds annotations to all Kubernetes resources
- Added `rbac.annotations` value to Helm chart, which adds annotations to RBAC-related Kubernetes resources

## [0.4.7] - 2023-11-07

### Fixed

- Use case-insensitive comparison when detecting hostname drift

## [0.4.6] - 2023-11-07

### Fixed

- Ensure Helm Chart version is bumped in lockstep with Proclaim version

## [0.4.5] - 2023-11-07

### Fixed

- Always use TCP (not UDP) for DNS-SD reconcilation queries to avoid truncated responses
- Fixed false-negative reported for the `Discoverable` condition

### Changed

- Added brief description of the drifted values when `Discoverable` is `False` due to `LookupResultOutOfSync`

## [0.4.4] - 2023-11-05

### Changed

- The dependency on the `proclaim` secret is now optional

## [0.4.3] - 2023-08-15

### Changed

- Proclaim is now built against Go v1.21
- Updated AWS, Kubernetes and DNS-SD related dependencies

## [0.4.2] - 2023-04-22

### Changed

- The Helm chart in versioned in lockstep with Proclaim itself

## [0.4.1] - 2023-04-22

### Changed

- Drastically reduced the re-reconciliation interval from 10 hours (the
  Kubernetes default), to the TTL of the DNS-SD instance (typically ~1 minute).
  This provides much more practical drift-detection behavior. Assuming there is
  no DNS record drift, the only overhead is a DNS query every TTL period.
- The `Discovered` event is now only emitted when a service instance is first
  discovered, or returns to being discoverable after a period of
  undiscoverability or drift. Prior to this change the event was emitted every
  time drift detection was performed.

### Added

- Added granular debug logging for advertise, unadvertise, discover and
  finalization operations.

## [0.4.0] - 2023-03-21

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
[0.4.1]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.1
[0.4.2]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.2
[0.4.3]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.3
[0.4.4]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.4
[0.4.5]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.5
[0.4.6]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.6
[0.4.7]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.7
[0.4.8]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.8
[0.4.9]: https://github.com/dogmatiq/proclaim/releases/tag/v0.4.9

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
