# Statping-ng Modernization & Hardening Report - PHASE 1

This document details the technical modernization and security hardening of the Statping-ng build orchestration. The objective was to achieve a production-grade, secure, and reproducible build by stabilizing a complex legacy dependency graph.

## 1. Patches Applied
The following legacy patches were integrated and applied to the source code before the modernization phase:

- **`statping-ng_000.patch`**: Dockerfile and Build script adjustments.
- **`statping-ng_001.patch`**: Visual removal of branding.
- **`statping-ng_003.patch`**: Notifier (email) branding cleanup.
- **`statping-ng_257.patch`**: Frontend asset path corrections (PR257).
- **`statping-ng_284.patch`**: Database connection pool optimizations (PR284).
- **`statping-ng_356.patch`**: Security hardening for legacy handlers (PR356).

---

## 2. Module Modernization Strategy

### The Hardening Model
Instead of a high-risk global upgrade (`go get -u`), we implemented a **Targeted Hardening** strategy. This model isolates the "System Stack" (foundational libraries like gRPC and Google Cloud) while upgrading the top-level modules (business logic and utility libraries).

### Implementation Phases:
1.  **Source Path Modernization**: Utilized `sed`-based source transformations to update legacy import identities (`bbolt`, `atomic`, `sarama`, `kingpin`, `jwt-go`) to their modern canonical paths. This resolved deep-level identity collisions without manual refactoring.
2.  **CGO & Hardened Linking**:
    -   Integrated the build with a modern **SQLite 3 (/opt/lib/sqlite3)** via `pkg-config`.
    -   Injected hardened linker flags: **Full RELRO**, **Immediate Binding (`-z now`)**, and **Symbol Stripping (`-s`)**.
3.  **Curated Upgrade Loop**: Explicitly upgraded 20+ high-impact libraries to their latest **2024/2025 releases**, while pinning the project's legacy gRPC core to its compatible 2020 era.

---

## 3. Upgraded Module Manifest

### Core Infrastructure & Logic
| Module | Version | Reasoning |
| :--- | :--- | :--- |
| `github.com/spf13/viper` | `@latest` (v1.21.0) | Modern configuration engine security. |
| `github.com/spf13/cobra` | `@latest` (v1.10.2) | Hardened CLI command handling. |
| `github.com/sirupsen/logrus` | `@latest` (v1.9.4) | Final stable release of the logrus engine. |
| `github.com/prometheus/client_golang` | `@latest` (v1.23.2) | Compatibility with modern metrics scrapers. |

### API & HTTP Engines
| Module | Version | Reasoning |
| :--- | :--- | :--- |
| `github.com/gin-gonic/gin` | `@latest` (v1.12.0) | Critical security patches for the core API. |
| `github.com/labstack/echo/v4` | `@latest` (v4.15.2) | Modernized routing and middleware security. |
| `github.com/valyala/fasthttp` | `@latest` (v1.71.0) | Performance and security for notifiers. |
| `github.com/go-resty/resty/v2` | `@latest` (v2.17.2) | Improved timeout/retry logic for health checks. |

### Database & Persistence
| Module | Version | Reasoning |
| :--- | :--- | :--- |
| `github.com/jinzhu/gorm` | **v1.9.16** | Final stable patch of the GORM v1 series. |
| `github.com/mattn/go-sqlite3` | **v1.14.44** | Latest security-patched stable v1 release. |
| `github.com/lib/pq` | `@latest` (v1.12.3) | Hardened Postgres communication. |
| `github.com/go-sql-driver/mysql` | `@latest` (v1.10.0) | Modern MySQL driver stability. |

### Security & Reliability
| Module | Version | Reasoning |
| :--- | :--- | :--- |
| `github.com/golang-jwt/jwt` | **v3.2.2** | Secure drop-in replacement for `dgrijalva/jwt-go`. |
| `github.com/microcosm-cc/bluemonday` | `@latest` (v1.0.27) | **Critical XSS Protection** for status pages. |
| `github.com/yuin/goldmark` | `@latest` (v1.8.2) | ReDoS-safe Markdown parsing engine. |
| `github.com/go-playground/validator/v10` | `@latest` (v10.30.2) | Secure input validation for all API endpoints. |
| `github.com/nats-io/nats.go` | `@latest` (v1.52.0) | Reliable distributed monitoring messaging. |
| `github.com/go-ping/ping` | `@latest` (v1.2.0) | Modern raw-socket ICMP handling. |

---

## 4. Final Build Integrity
- **Synchronization Status**: All modules verified via `go mod verify`.
- **Era Anchoring**: Foundational gRPC and Google Cloud modules are anchored to the mid-2020 era to maintain monolithic compatibility.
- **Binary Status**: Production-ready, stripped, and hardened with Full RELRO.
