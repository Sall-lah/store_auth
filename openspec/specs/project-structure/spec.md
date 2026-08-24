# project-structure Specification

## Purpose
Feature-based code organization in internal/ packages.

## Requirements

### Requirement: Feature-based code organization
The system codebase SHALL be organized into feature-centric packages (`internal/auth`, `internal/otp`, `internal/jwt`) rather than layer-centric directories.

#### Scenario: Code compilation in feature-based layout
- **WHEN** the application is compiled using `go build ./...`
- **THEN** all feature packages compile without circular import errors
