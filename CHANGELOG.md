# 0.97.1 (07-31-2026)
- **UI: Sticky Top Navigation Bar**:
  - Added professional top navigation banner to public pages
  - Real-time status summary (X/Y services online)
  - 24-hour uptime percentage with color-coded thresholds
  - Last check timestamp with human-readable format
  - Active incident count indicator
  - Dashboard and Login buttons always visible
  - Responsive design with graceful mobile degradation
- **Theme System Simplification**:
  - Replaced SASS compilation with simple CSS-only custom themes
  - Removed external `sass` binary dependency
  - Custom CSS saved directly to `assets/custom.css`
  - Simplified theme API handlers (create/save/delete)
- **Email Templates Modernization**:
  - Replaced MJML-generated bloat with clean, enterprise-grade HTML
  - Deleted `emails_local/` module (separate go.mod, ~700 lines)
  - Deleted `notifiers/email_rendered.go` (unused duplicate)
  - New responsive templates: ServiceOnline, ServiceOffline (~200 lines total)
  - Modern gradient headers, status badges, info cards
  - Compatible with Gmail, Outlook, Apple Mail
- **Test Infrastructure**:
  - Added `testfiles/` directory for persistent test artifacts (gitignored)
  - Build outputs now go to `testfiles/statping.exe`
  - All tests use `t.TempDir()` for ephemeral files
  - Added fully documented `testdata/config.yml` example
  - Cleaned `testdata/` to only immutable fixtures
- **Code Cleanup**:
  - Deleted unused generators (`generate_help.go`, `generate_languages.go`, `generate_version.go`)
  - Removed legacy wiki submodule reference
  - Reduced source package by ~1000 lines

# 0.97.0 (07-30-2026)
- **Frontend: Vue 3 Migration**:
  - Complete migration from Vue 2.7 to Vue 3.5 with Composition API
  - Replaced Vuex with Pinia for state management
  - Migrated from Webpack 4 to Vite 6 (10-100x faster HMR)
  - Upgraded vue-router 3.x → 4.x, vue-i18n 8.x → 10.x
  - Migrated all 55 SFCs to Composition API
  - Fixed Login.vue session persistence bug (no longer clears cookie on mount)
  - Replaced Cypress with Playwright + Vitest for E2E testing
- **Security: Encryption Key Separation**:
  - Added dedicated `EncryptionKey` field separate from `ApiSecret`
  - API secret rotation no longer breaks encrypted data
  - Encryption key never exposed via API (`json:"-"`)
  - Race-safe key generation with conditional UPDATE for clustered deployments
  - Encryption failures now block save (GORM hooks return errors)
- **Microsoft Teams Notifier**:
  - New Teams notifier using Power Automate Workflows webhooks
  - Modern Adaptive Cards for service alerts (up/down with latency)
  - Implements `DigestNotifier` interface for daily digest support
- **Multi-Channel Daily Digest**:
  - Added `DigestNotifier` interface for digest-capable notifiers
  - New `receive_digest` toggle on all notifiers (not just email)
  - Teams, Slack, Discord can now receive daily service summaries
  - `sync.Once` pattern for digest scheduler initialization
- **LDAP/LDAPS Authentication**:
  - Full LDAP/LDAPS support with configurable bind DN and search base
  - User/group attribute mapping (OpenLDAP, Active Directory, FreeIPA templates)
  - Encrypted bind password storage (AES-256-GCM)
  - Test Connection button for validating LDAP settings
  - Semicolon-separated authorized groups (commas are part of DN syntax)
- **Build System**:
  - Merged `build.sh` and `code_quality.sh` into unified `tools/build.sh`
  - Added `--audit`, `--test`, `--all`, `--clean`, `--clean-all`, `--extra-scans` flags
  - 8 quality checks: shellcheck, go vet, gosec, govulncheck, race detection, tests, oxlint, biome
  - Grype supply chain scanning available via `--extra-scans`
- **Security Fixes**:
  - `rand.Read` failures now panic instead of silent empty return
  - `NewSHA256Hash()` panics on crypto/rand failure; `GenerateSHA256Hash()` returns errors
  - `IsEncrypted()` uses `enc:` prefix marker (not fragile heuristics)
  - Session timeout capped at 30 days maximum
  - Encryption failures now block database save with proper error propagation
- **Code Quality**:
  - Fixed all `go vet` mutex copy warnings across codebase
  - Notifier interface now uses `*Service` pointers (OnSuccess/OnFailure)
  - Service methods converted to pointer receivers (Hash, Downtime, Uptime, etc.)
  - `AllInOrder()` returns `[]*Service` instead of copying values
  - New `ServicePtrOrder` sort type replaces deprecated `ServiceOrder`
  - Fixed IPv6 address format in SMTP diagnostics (`net.JoinHostPort`)
  - Added explicit 30-second timeout to digest SMTP connections
- **Test Coverage**:
  - Added unit tests for digest handler API endpoints
  - Added unit tests for LDAP handler (settings, templates, group membership)
  - Added unit tests for Teams notifier (card structure, interfaces)
  - Added unit tests for digest notifier (email parsing, formatting, rendering)
- **Bug Fixes**:
  - Fixed `MigrationId` type mismatch causing database errors
  - Fixed service chart rendering issues
  - Removed deprecated Vue 2 frontend directory

# 0.96.6 (07-25-2026, unreleased)
- **Security Hardening**:
  - Added CSRF double-submit cookie protection with cryptographic tokens
  - Added rate limiting middleware (5 requests/minute for login endpoint)
  - Added OAuth state validation with one-time use enforcement and expiry cleanup
  - Added input validation for service types (HTTP, TCP, gRPC, ICMP, static, command)
- **Test Architecture Overhaul**:
  - Switched from in-memory SQLite to file-based SQLite with WAL mode for test isolation
  - Added `sync.Once` pattern for thread-safe handler test setup, eliminating race conditions
  - Added `services.ClearCache()` to reset stale cache before test database setup
  - Added 94MB production-scale test fixture (607K hits, 11K failures) in `testdata/`
  - Test coverage improved from ~36% to 45.2%
- **New Test Coverage**:
  - `database/database_test.go`: 18 comprehensive database layer tests
  - `handlers/csrf_test.go`: CSRF token generation and validation tests
  - `handlers/ratelimit_test.go`: Rate limiting allow/block/reset tests
  - `handlers/dashboard_test.go`: Theme and config API tests
  - `types/services/check_test.go`: HTTP/TCP/Command check tests with mocks
  - `types/services/validation_test.go`: Service validation with edge cases
  - `types/core/core_test.go`, `types/hits/hits_test.go`, `types/metrics/metrics_test.go`, `types/notifications/notifications_test.go`
- **Documentation**:
  - Added `ARCHITECTURE.md` with Mermaid diagrams (system, data flow, deployment)
  - Rewrote `README.md` with security assessment, CLI args, deployment examples
  - Added `TESTING.md` with complete test matrix, coverage stats, troubleshooting
  - Renamed original README.md to `LEGACY.md`
- **Bug Fixes**:
  - Fixed `dashboard.go` config file path: `configs.yml` → `config.yml`
- **Frontend**:
  - Added CSRF token handling to `API.js`
  - Improved form validation and error handling in Login.vue

# 0.96.0 (07-24-2026, unreleased)

# 0.95.0 (07-23-2026, unreleased)
- **Security & Backdoor Removal**:
  - Removed `GO_ENV=test` authentication backdoor in `handlers/authentication.go` for 100% environment parity.
  - Fixed `IsUser()` API key and Bearer token privilege resolution on private service queries.
  - Added live HTTP security regression test suite enforcing 401 unauthenticated responses and secret redaction.
- **Database & Schema Indexing Optimization**:
  - Added composite indexes (`idx_hits_service_created_at`, `idx_failures_service_created_at`, `idx_checkin_hits_created_at`) for sub-millisecond time-series queries.
  - Added indexed `api_key` lookup on `users` table for rapid API token authentication.
  - Enforced foreign key constraints (`ON DELETE CASCADE`) across relational tables.
- **Frontend & Supply Chain Hardening**:
  - Upgraded Vue to **Vue 2.7.16** (final official 2.x release) and `axios` to `^0.33.0`.
  - Configured NPM `overrides` (`lodash`, `qs`, `serialize-javascript`, `elliptic`, `express`, `node-forge`) and committed `package-lock.json`.
  - Resolved all JS/Vue diagnostics across 84 application source files (`oxlint` / `biome` 0 errors).
- **Tooling & Code Hygiene**:
  - Cleaned all repository shell scripts (`install.sh`, `dev/*.sh`) for 100% `ShellCheck` compliance.
  - Created [`code_quality.sh`](file:///code_quality.sh) script running all 8 audit phases in sequence.
- **Test Suite Compliance & Stability**:
  - Updated test suites across `types/users`, `types/services`, `utils`, and `handlers` for 100% compliance with 30-character password complexity rules.
  - Added nil pointer guard in `types/services/notifications.go` (`logMessage`).
  - Fixed test directory path assertions in `utils/utils_test.go`.

# 0.94.0 (05-22-2026, unreleased)
- **Core Modernization & ORM Migration**:
  - Migrated core database layer from legacy `jinzhu/gorm` (v1) to `gorm.io/gorm` (v2).
  - Migrated Cloud SDKs to `aws-sdk-go-v2` (SNS and Translate modules).
  - Security hardening: SHA256 service hashing, bcrypt cost 14, `crypto/rand` token generation, and `golang-jwt/jwt/v5` upgrade (CVE-2025-3553).
  - Replaced legacy bulk insertion with native `CreateInBatches()`.
- **Database Scalability & Indexing**:
  - Comprehensive query performance optimizations and high-volume index creation (`idx_hits_created_at`, `idx_failures_created_at`, `idx_checkin_hits_created_at`).
  - Added native foreign key constraints with `ON DELETE CASCADE` across relational tables.
  - Automatic secret credential collection (`statping.secrets` mode `0400`).
- **UI & Dashboard Overhaul**:
  - Widescreen `container-fluid` layout expansion removing 1140px fixed width bounds.
  - Interactive 6-month Service Heatmap with month dividers and hourly failure breakdown drill-down.
  - Advanced Dual Y-Axis Mixed Chart overlaying failure columns directly onto latency/ping curves.
  - Global card-level click navigation using Bootstrap `.stretched-link` pattern.
  - Total failure count dynamically displayed in chart titles.

# 0.93.0 (06-04-2025)
- **Base Release (Commit `c02149f0`)**:
  - Implemented proper checkin monitoring routines (`proper-checkins`).
  - Added support for custom OAuth scope parameters.

# 0.92.0 (01-15-2025)
- **CI/CD & Dependencies**:
  - Updated GitHub Actions dependencies (`@actions/upload-artifact` v4, `@actions/setup-go` v5).
  - Reworked multi-arch container build matrix and release step workflows.

# 0.91.0 (07-10-2023)
- **Monitoring & Protocol Expansion**:
  - Added SMTP and IMAP service check types.
  - Added HTTP `HEAD` method support for service monitoring.
  - Added notification icons to status page for active service announcements.
- **UI & Accessibility Improvements**:
  - Moved service messages to top of status pages for higher visibility.
  - Added Swedish language translation and updated Czech language translations.
- **Bug Fixes**:
  - Fixed local login authentication logic and checkin latency calculations for static services during timeouts.
  - Fixed healthcheck ports and Docker multi-architecture builds (`build.yaml`).

# 0.90.80 (01-26-2022)
- Fixed permissions on /app directory - Thanks twouters

# 0.90.79 (01-24-2022)
- Updated Russian Language - Thanks meatlayer
- Docker file fix for BASE_PATH and health checks - Thanks michaelkrieger
- Removed statping emailer notifier (not SMTP Mail)
- Fixes for notification failures (Issue statping#911) - Thanks glanchow
- Updated Home page uptime wording (24hr/7days) - Thanks Jonathanrbarney & thatInfrastructureGuy
- [GITHUB] Removed mailer tests

# 0.90.78 (09-15-2021)
- HTTP Webhooks accept multiple HTTP Headers
- Modified Telegram notifier to allow chat_ids
- New Notifier - Mattermost
- Updated German Language - Thanks Flofeld
- Czech Language - Thanks Fjuro
- Some minor branding Changes
- Moved some asset dependancies from assets.statping.com
- Fixed the (Ubuntu) Snap Store build script
- Retrospectively updated the Changelog
- [GITHUB] Fixed Windows/Mac autobuilds
- [GITHUB] Unstable container build
- [GITHUB] Triggers SNAP builds

# 0.90.77 (08-18-2021)
- More branding changes
- Fix for go statping-ng/email deps (https://github.com/statping-ng/statping-ng/issues/9)
- [GITHUB] Fixed autobuilds

# 0.90.76 (08-13-2021)
- Forked statping and renamed to statping-ng
- Branding changes
