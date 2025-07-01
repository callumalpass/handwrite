# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-07-01

### Changed
- **BREAKING:** Complete rewrite from Python to Go
- **BREAKING:** New CLI interface using Cobra framework
- **BREAKING:** Configuration file format updated for Go compatibility

### Added
- Go implementation with significantly improved performance
- Concurrent processing with configurable worker pools
- Real-time progress bars for batch operations
- Cross-platform binary releases (Linux, macOS, Windows)
- Comprehensive test suite
- GitHub Actions CI/CD pipeline
- Makefile for development workflow
- golangci-lint configuration for code quality

### Improved
- Performance improvement through Go's concurrency
- Single binary distribution - no dependency management
- Better error handling and logging
- Type safety with Go's strong typing system
- Memory efficiency for large PDF processing

### Removed
- Python implementation and dependencies
- setuptools/pip installation method
- Python-specific configuration options

## [0.x.x] - Previous Python Versions

See git history for Python version changelog.

