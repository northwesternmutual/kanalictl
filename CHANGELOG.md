# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2017-10-07
### Added
- Added decrypt subcommand.
### Changed
- Fixed bug that created an invalid API key if not providing an existing one.
- Upgraded to Kanali version `v1.2.0`
- Modified subcommands

## [1.0.2] - 2017-08-22
### Added
- Support for multiple YAML documents in a single file.
### Changed
- Generating API key of length 32 instead of 16.
- Fixed bug that used wrong capitalization for Kanali kinds.

## [1.0.1] - 2017-08-11
### Changed
- Gracefully handling case where configuration isn't being used.

## [1.0.0] - 2017-08-10
### Added
- Initial release