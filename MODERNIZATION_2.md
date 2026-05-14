# Statping-ng Modernization & Hardening Report - PHASE 2

This document details the final transition of the Statping-ng database layer to GORM v2 and the resolution of the modern dependency graph.

## 1. Database Architecture (GORM v2 Migration)

The core database layer has been successfully migrated from `jinzhu/gorm` (v1) to `gorm.io/gorm` (v2).

### Compatibility Layer
To maintain stability across the application's legacy query patterns, a **Compatibility Layer** was implemented in `database/database.go`.
- **Polymorphic `Update`**: Automatically handles both single-struct and column/value update signatures.
- **State Preservation**: Refactored the shim to use internal `wrap()` methods that preserve `Type` and `ReadOnly` fields across chained queries.
- **Migrator Shims**: Restored `CreateTable`, `DropTableIfExists`, and `AddIndex` as wrappers for the GORM v2 Migrator API.
- **Enhanced Count**: Standardized all `Count()` operations to use `int64` pointers, fixing return-type mismatches.

### Native Bulk Insertion
The project was decoupled from the legacy `github.com/t-tiger/gorm-bulk-insert` module.
- **Implementation**: Refactored hits and failures sampling to use native GORM v2 `.CreateInBatches()`.
- **Outcome**: Improved data ingestion reliability and eliminated a legacy ORM dependency.

### Search & Retrieval Standardization
Refactored all single-object `Find` methods (Users, Groups, Incidents, Messages, Checkins) to use `.First()`. This ensures that `gorm.ErrRecordNotFound` is correctly triggered and handled by the API handlers.

## 2. Cloud SDK Modernization (AWS SDK v2)

The project has been migrated from the deprecated `aws-sdk-go` (v1) to the modern `aws-sdk-go-v2` stack.
- **Amazon SNS**: Refactored the SNS notifier to use the v2 client with context-aware calls and modern configuration patterns.
- **Language Utility**: Updated the language generation tool to use the v2 Translate client, eliminating legacy v1 dependencies.

## 3. Phase 2 Module Manifest

Building upon Phase 1, the following modules were refactored for modern SDK compatibility:

| Category | Module | Modern Version | Note |
| :--- | :--- | :--- | :--- |
| **ORM** | `gorm.io/gorm` | **v1.25.12+** | Core GORM v2 engine. |
| **Cloud SDK** | `github.com/aws/aws-sdk-go-v2` | **v1.36.0+** | Modern AWS SDK v2 transition. |
| **Security** | `getsentry/sentry-go` | **v0.31.1+** | Refactored for modern structured context. |
| **Protobuf** | `google.golang.org/protobuf` | **v1.36.0+** | Modern proto3 canonical implementation. |
| **Drivers** | `gorm.io/driver/*` | **Latest** | Canonical v2 drivers for MySQL, Postgres, SQLite. |

## 4. Code Quality & Security Hardening

A comprehensive code audit was performed to resolve technical debt and security vulnerabilities identified in the audit pipeline.

### Security Hardening
- **Cryptographic Primitives**:
    - Replaced weak `math/rand` with `crypto/rand` for secure token/hash generation.
    - Upgraded `Service.Hash()` from SHA1 to SHA256.
    - **JWT Security**: Upgraded `github.com/golang-jwt/jwt` to `v5` to remediate `GO-2025-3553` (CVE-2025-3553).
- **TLS/SSL Hardening**:
    - **Cipher Suites**: Removed weak CBC-mode ciphers from the SSL server, enforcing modern GCM-based suites.
    - **InsecureSkipVerify**: Fixed insecure defaults in the Email notifier; verification is now enabled by default.
- **Session Security**:
    - Hardened JWT cookies by enabling `HttpOnly` and `SameSite=Lax` attributes to mitigate XSS and CSRF risks.
- **Permissions**: Tightened directory creation permissions from `0777` to a secure `0750`.
- **Integer Safety**: Resolved multiple `G115` integer overflow vulnerabilities in type conversion and system UID/GID comparisons.

### Code Hygiene & Audit Remediation
- **Error Handling (`errcheck`)**:
    - Systematically resolved over 50+ unhandled error cases (`G104`) across the codebase.
    - Explicitly acknowledged non-critical errors in tests (`InitLogs`, `Setenv`, `Unsetenv`) and production paths (`Body.Close()`, `gz.Close()`, `w.Write()`, `conn.Close()`, `conn.Logout()`) using blank identifiers.
    - Added error logging for background server instances (pprof and redirect servers).
- **Dead Code Removal (`unused`)**:
    - Purged unused functions and types across the `cmd`, `handlers`, `notifiers`, and `database` packages (e.g., `importAll`, `updateDisplay`, `cacheJson`).
    - Removed unused constants (`emailSuccess`), fields (`loggable`, `mu`), and internal types (`notificationHits`) to satisfy "Zero Error" audit requirements.
- **Static Analysis (`staticcheck`)**:
    - Modernized gRPC usage by replacing deprecated `WithInsecure` and `grpc.DialContext` with `grpc.NewClient`.
    - Refactored shadowed `err` variables in walk functions and fixed duplicate JSON struct tags.
    - Refactored ticker-based maintenance loops to use modern `for range` patterns.
    - Fixed duplicate struct tags, invalid `bitSize` arguments in `ParseFloat`, and inconsistent error string capitalization.
- **Cleanliness**: Pruned all unused imports and standardized the entire codebase using `gofumpt`.

## 5. Final Verification
- **Compilation**: Successfully produced a production binary using the GORM v2 stack.
- **Dependency Audit**: Verified a clean, synchronized graph via `go mod tidy`.
- **Binary Status**: Production-ready and verified with `statping version`.
