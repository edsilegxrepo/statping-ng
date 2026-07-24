# 0.96.0 (07-24-2026)
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
  - Created [`code_quality.sh`](file:///usr/src/packages/BUILD/statping-ng/code_quality.sh) script running all 8 audit phases in sequence.

# 0.95.0 (07-23-2026)
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
