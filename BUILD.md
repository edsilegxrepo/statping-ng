# Statping-ng Build Orchestration Guide

This document details the build orchestration workflow for Statping-ng, specifically focusing on the `statping_build.sh` pipeline and the modernized build requirements.

## 1. Orchestration Overview

The build process is managed by `statping_build.sh` (located in the RPM `SOURCES` directory). The script has been modernized to eliminate legacy hacks and leverage the hardened repository state.

### Key Orchestration Principles
- **Single Source of Truth**: All patches, de-branding templates, and source modernization are now integrated directly into the `edsilegxrepo` repository. The build script no longer applies external patches.
- **Clean Environment**: The script enforces a clean build by performing a fresh clone and running `go mod tidy` before compilation.
- **Native Pipeline**: Optimized for native execution on RHEL/CentOS systems, using pre-configured toolsets in `/u01`.
- **Security Hardening**: Enforces modern security standards including PIE (Position Independent Executables), Full RELRO, and ASLR compatibility.

## 2. Build Environment & Requirements

The build orchestrator expects a specific set of tools and environment configurations:

### Toolchain Dependencies
- **Node.js / Yarn**: Required for the Vue.js frontend build. (Node v24+, Yarn v1.22+).
- **Go**: Required for the backend build (Go 1.26+).
- **Rice**: Used for embedding static assets into the Go binary.
- **Sass**: Required for CSS preprocessing in the frontend.

### Critical Environment Variables
- `NODE_OPTIONS="--openssl-legacy-provider"`: Set to maintain compatibility with modern Node versions and legacy Webpack configurations.
- `GOPROXY="https://proxy.golang.org,direct"`: Ensures reliable and secure dependency fetching.
- `PATH`: Automatically extended to include `/u01/yarn/bin`, `/u01/node/bin`, and `/u01/tools`.

## 3. Frontend Compilation (Vue.js)

The frontend build workflow has been standardized to ensure consistency across different build environments.

- **Dependency Management**: Uses `yarn install --frozen-lockfile` to ensure that exact dependency versions are used without modifying project files during the build.
- **Browser Compatibility**: Automatically updates the browserslist database (`npx update-browserslist-db`) to ensure the latest CSS autoprefixer telemetry is used.
- **Asset Integration**: Compiled assets are automatically moved to the Go `source/` directory for embedding via `go.rice embed-go`.

## 4. Backend Compilation (Golang)

The backend build utilizes the modernized Go module graph.

- **Dependency Anchoring**: Relies on the `tools.go` pattern within the repository to maintain build-time generator dependencies (Minify, Markdown, AWS Translate).
- **Hardened Linking**:
    - **CGO_LDFLAGS**: Injects `-Wl,-z,relro -Wl,-z,now` for immediate binding and Full RELRO.
    - **Stripping**: Uses `-s -w` linker flags to remove DWARF and symbol tables, reducing binary size.
    - **ASLR/PIE**: Uses `-buildmode=pie` to ensure the resulting binary supports Address Space Layout Randomization.
- **CGO Configuration**: Linked against a custom, hardened SQLite 3 instance located at `/opt/lib/sqlite3/lib`.

## 5. Build Command Syntax

To build a production archive:

```bash
./statping_build.sh <VERSION> <GIT_HASH> [--native] [--purge]
```

- **`--native`**: Executes the native build pipeline (default).
- **`--purge`**: Automatically cleans up the build directory after a successful archive creation.
