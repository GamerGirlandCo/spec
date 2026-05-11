# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-05-11

### Added
- OpenAPI 3.2.0 support and expanded version-specific validation checks.
- Validation severity levels (`Error`, `Warning`, `Info`) with best-practice reporting.
- `ValidateReport` support across the core router and all adapters.
- Dirty-state tracking for incremental build behavior.
- Automatic injection of missing path parameters.
- Support for `application/x-www-form-urlencoded` querystring parameter content.
- Media type tag support in schema reflection.
- New Iris adapter: `irisopenapi`.

### Changed
- Rebuilt core generator, reflector, and validator internals.
- Aligned all adapters with the updated core behavior.
- Upgraded OpenAPI handling to improve version consistency and validation behavior (including 3.1.2 alignment).
- Prefixed package name in default schema definition names.
- Centralized operation method validation logic.

### Fixed
- Hardened path and HTTP method generation behavior.

### Removed
- Removed `httprouter` from the supported adapter list.
- Removed dependency on external OpenAPI generator packages (including `github.com/swaggest/openapi-go`) as part of the core generator rebuild.

## [0.4.2] - 2026-04-22

### Fixed
- Shortened generic type definition names in generated JSON schema.
- Resolved Echo v5 adapter test breakage caused by upstream library changes.

### Changed
- Updated `swaggest/openapi-go` to `v0.2.61`.

## [0.4.1] - 2026-04-12

### Added
- Config option to strip trailing slashes from paths.

### Fixed
- URL paths with trailing slashes now generate correctly.

### Changed
- Updated Fiber v2 and Echo v5 dependencies (core and adapters).

## [0.4.0] - 2026-04-05

### Added
- Echo v5 adapter (`echov5openapi`).
- Fiber v3 adapter (`fiberv3openapi`).
- Custom `spec-ui` option support, including `spec-ui` upgrade to `v0.2.0`.

### Changed
- Merged duplicate status-code responses into a single `oneOf` response schema.

## [0.3.6] - 2025-11-05

### Added
- Custom parameter tag configuration support.

### Fixed
- Echo route syntax mapping now correctly supports `:id` style parameters.

## [0.3.5] - 2025-10-25

### Added
- Ability to configure the OpenAPI `encoding` object for request bodies.
- Added `RELEASE.md` release workflow documentation.

### Changed
- Updated adapter dependency versions.

[0.5.0]: https://github.com/oaswrap/spec/compare/v0.4.2...v0.5.0
[0.4.2]: https://github.com/oaswrap/spec/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/oaswrap/spec/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/oaswrap/spec/compare/v0.3.6...v0.4.0
[0.3.6]: https://github.com/oaswrap/spec/compare/v0.3.5...v0.3.6
[0.3.5]: https://github.com/oaswrap/spec/compare/v0.3.4...v0.3.5
