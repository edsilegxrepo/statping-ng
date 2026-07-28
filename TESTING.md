# Statping-ng Test Suite Documentation

> [!IMPORTANT]
> **Sequential Execution Required** — All tests MUST be run with `-p 1` to prevent global state contamination:
> ```bash
> go test ./... -p 1 -count=1
> ```
> See [Why Sequential Execution](#why-sequential-execution--p1-is-required) for details.

> [!NOTE]
> **Integration Tests** require an explicit build tag and are NOT included in default runs:
> ```bash
> go test -tags=integration ./integration -p 1 -v
> ```
> See [Integration Tests](#integration-tests) for environment variables and setup.

---

## Architecture, Design and Principles

### Test Philosophy
- **Chain Isolation**: Each package uses `TestMain` to create isolated database state and stop leaked goroutines
- **Dynamic IDs**: Tests use `example.Id` or helper functions instead of hardcoded IDs for resilience
- **Realistic Data**: Integration tests use a 94MB production-scale database fixture (607K hits, 11K failures)
- **Cross-Platform**: Tests run identically on Windows (MinGW) and Linux
- **Security-First**: Security regression tests verify authentication, authorization, CSRF, and rate limiting

### Test Chain Architecture
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         TEST CHAIN ISOLATION                                │
│  Each package has TestMain that: StopAll() → ClearCache() → Fresh DB       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐           │
│  │   handlers/     │   │ types/services/ │   │  types/users/   │           │
│  │   main_test.go  │   │   main_test.go  │   │   main_test.go  │           │
│  │                 │   │                 │   │                 │           │
│  │ StopAll()       │   │ StopAll()       │   │ StopAll()       │           │
│  │ ClearCache()    │   │ ClearCache()    │   │ ClearCache()    │           │
│  │ TempDir setup   │   │ OpenTester()    │   │ OpenTester()    │           │
│  │ └→ api_test     │   │ SetDB (7 pkgs)  │   │ SetDB()         │           │
│  │ └→ services_... │   │ AutoMigrate()   │   │ AutoMigrate()   │           │
│  │ └→ users_test   │   │ └→ services_... │   │ └→ users_test   │           │
│  │ └→ ...          │   │ └→ routine_...  │   │ └→ ...          │           │
│  └─────────────────┘   └─────────────────┘   └─────────────────┘           │
│                                                                             │
│  ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐           │
│  │  types/groups/  │   │types/incidents/ │   │   Standalone    │           │
│  │   main_test.go  │   │   main_test.go  │   │   (no TestMain) │           │
│  │                 │   │                 │   │                 │           │
│  │ StopAll()       │   │ (no StopAll -   │   │ database/       │           │
│  │ ClearCache()    │   │  import cycle)  │   │ notifiers/      │           │
│  │ OpenTester()    │   │ OpenTester()    │   │ types/checkins/ │           │
│  │ SetDB()         │   │ AutoMigrate()   │   │ types/hits/     │           │
│  │ └→ groups_test  │   │ └→ incidents_.. │   │ types/failures/ │           │
│  └─────────────────┘   └─────────────────┘   └─────────────────┘           │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                    Integration Tests (build tag: integration)               │
│         (Full HTTP stack, real database fixture, live APIs)                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Test Chain Dependencies

| Package | TestMain | Isolation | Depends On |
|---------|----------|-----------|------------|
| `handlers/` | Yes | StopAll + TempDir | all types |
| `types/services/` | Yes | StopAll + OpenTester | hits, failures, checkins, incidents, messages, notifications |
| `types/users/` | Yes | StopAll + OpenTester | none |
| `types/groups/` | Yes | StopAll + OpenTester | services |
| `types/incidents/` | Yes | OpenTester only | none (import cycle prevents StopAll) |
| `database/` | No | Per-test isolation | none |
| `notifiers/` | No | Per-test isolation | none |
| `types/checkins/` | No | Per-test isolation | none |

### Key Design Decisions
1. **Per-Chain TestMain Isolation**: Each package with shared state has a `TestMain` that calls `services.StopAll()` and `services.ClearCache()` before creating a fresh database
2. **Dynamic ID Lookups**: Tests use `example.Id` or helper functions (`getFirstServiceID()`, `getServiceByIndex()`) instead of hardcoded `1`, `2`, etc.
3. **Thread-Safe Service Operations**: `Service.Running` channel protected by mutex; `checkpoint`/`sleepDuration` use getters/setters
4. **File-Based SQLite**: `database.OpenTester(tmpDir)` creates isolated file-based databases, avoiding in-memory connection issues
5. **Fixture Database for Integration**: `testdata/statping.db.xz` provides realistic data volume
6. **Goroutine Cleanup**: `StopAll()` stops service check goroutines before clearing cache to prevent data races

---

## Logic Flow of Tests

### Main Categories

| Category | Package | Purpose |
|----------|---------|---------|
| **Unit Tests** | `types/*` | Test individual type methods and database operations |
| **Handler Tests** | `handlers/` | Test HTTP endpoints, routing, authentication |
| **Integration Tests** | `integration/` | End-to-end tests with real database and HTTP server |
| **Notifier Tests** | `notifiers/` | Test notification delivery (often skipped without credentials) |
| **Utility Tests** | `utils/` | Test helper functions, crypto, file operations |

### Positive Testing
- Valid API requests return expected data
- Authentication with valid credentials succeeds
- CRUD operations create/update/delete correctly
- Service checks return proper status

### Negative Testing
- Invalid credentials return 401 Unauthorized
- Missing CSRF tokens return 403 Forbidden
- Rate-limited requests return 429 Too Many Requests
- Malformed JSON returns 400 Bad Request
- Non-admin users cannot access admin routes

---

## Technical Requirements and Setup

### Dependencies

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Backend compilation and testing |
| Node.js | 18+ | Frontend build |
| MinGW (Windows) | GCC 12+ | CGO compilation on Windows |
| xz | any | Decompress test database fixture |
| rice | latest | Embed frontend assets |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CC` | system gcc | C compiler path (MinGW on Windows) |
| `DB_CONN` | sqlite | Database type for tests |
| `GO_ENV` | test | Environment mode |
| `NODE_OPTIONS` | `--openssl-legacy-provider` | Required for webpack compatibility |

### Constraints
- **Tests must run with `-p=1` flag (sequential)** - See "Why Sequential Execution" below
- Integration tests require `-tags=integration` build tag
- Integration tests require `testdata/statping.db.xz` fixture
- Frontend assets must be built before handler tests (`source/rice-box.go`)
- Windows requires MinGW, not Cygwin GCC

---

## Why Sequential Execution (`-p=1`) is Required

### The Problem: Global State Contamination

This codebase uses several global singletons that cause test failures when packages run in parallel:

1. **`core.App`** - Single global application state shared across all packages
2. **`services.allServices`** - Global map of all services with mutex protection
3. **Database connections** - Package-level `db` variables in `types/*` packages
4. **Notifier registry** - Global `allNotifiers` map

When Go runs tests with default parallelism (`-p=N` where N = CPU cores), multiple packages execute simultaneously. Package A's `TestMain` might initialize the database while Package B's tests are mid-execution expecting different data.

### Symptoms of Parallel Execution Failures

```
--- FAIL: TestServiceCheck
    Expected 6 services, got 0
--- FAIL: TestUserLogin  
    incorrect authentication (user deleted by parallel test)
--- FAIL: TestNotifierSend
    notifier not found in registry
```

### The Solution

**Always run tests with `-p=1`:**

```bash
# Correct - sequential execution
go test ./... -p 1 -count=1

# Wrong - parallel execution causes flaky failures
go test ./...
```

### Package Isolation Strategy (Test Chain Pattern)

Each package with shared state uses `TestMain` with a **three-step isolation pattern**:

```go
func TestMain(m *testing.M) {
    // 1. STOP - Kill leaked goroutines from previous packages
    services.StopAll()      // Stops all service check goroutines
    services.ClearCache()   // Clears in-memory service cache

    // 2. SETUP - Create isolated database
    tmpDir, _ := os.MkdirTemp("", "statping-test")
    db, _ := database.OpenTester(tmpDir)
    SetDB(db)
    
    // 3. RUN - Execute tests in order (chain preserved)
    code := m.Run()
    
    // 4. CLEANUP - Stop services again, close DB
    services.StopAll()
    db.Close()
    os.RemoveAll(tmpDir)
    os.Exit(code)
}
```

**Why `StopAll()` is critical:**
- Service checks spawn background goroutines that write to `allServices` map
- Without stopping them, they continue running after package tests complete
- When the next package clears the cache, the goroutines panic on nil map access
- This caused intermittent data races that were hard to reproduce

**The `StopAll()` helper:**
```go
// types/services/database.go
func StopAll() {
    servicesLock.RLock()
    servicesCopy := make([]*Service, 0, len(allServices))
    for _, s := range allServices {
        servicesCopy = append(servicesCopy, s)
    }
    servicesLock.RUnlock()
    for _, s := range servicesCopy {
        s.Close()  // Signals Running channel to stop
    }
}
```

Even with this isolation, the global `core.App` and cross-package dependencies mean **sequential execution is the only reliable approach**.

---

## Integration Tests

### Build Tag Requirement

Integration tests are excluded from normal test runs and require an explicit build tag:

```bash
# Unit tests only (default)
go test ./... -p 1

# Integration tests only
go test -tags=integration ./integration -p 1

# All tests including integration
go test -tags=integration ./... -p 1
```

### Why Integration Tests Are Separate

1. **External Dependencies** - Integration tests may require real databases, network access, or API credentials
2. **Execution Time** - They take significantly longer (seconds vs milliseconds)
3. **CI/CD Flexibility** - Allows running fast unit tests on every commit, integration tests on merge

### Running Live Integration Tests

Integration tests use real HTTP servers and databases:

```bash
# 1. Ensure test fixture exists
ls testdata/statping.db.xz

# 2. Run integration tests
go test -tags=integration -v -p 1 ./integration

# 3. Run with specific test
go test -tags=integration -v -p 1 -run TestRealDatabaseIntegration ./integration
```

### Integration Test Environment Variables

For tests that connect to external services:

| Variable | Purpose | Example |
|----------|---------|---------|
| `SLACK_URL` | Slack webhook for notifier tests | `https://hooks.slack.com/...` |
| `DISCORD_URL` | Discord webhook URL | `https://discord.com/api/webhooks/...` |
| `TWILIO_SID` | Twilio account SID | `AC...` |
| `TWILIO_SECRET` | Twilio auth token | `...` |
| `PUSHOVER_TOKEN` | Pushover API token | `...` |
| `TELEGRAM_TOKEN` | Telegram bot token | `...` |
| `EMAIL_HOST` | SMTP server hostname | `smtp.gmail.com` |
| `EMAIL_USER` | SMTP username | `user@example.com` |
| `EMAIL_PASS` | SMTP password | `...` |

Tests automatically skip when credentials are missing:

```go
if SLACK_URL == "" {
    t.Log("Slack notifier testing skipped, missing SLACK_URL")
    t.SkipNow()
}
```

---

## List of Tests

### Integration Tests (`integration/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Database | `TestRealDatabaseIntegration` | Full API with production-scale data | All endpoints return valid JSON |
| Concurrent Load | `TestConcurrentLoadIntegration` | 30 workers × 20 requests | No race conditions, all requests succeed |
| CRUD Lifecycle | `TestFullCRUDLifecycleIntegration` | Create, Read, Update, Delete flow | Entity persists and deletes correctly |
| Security | `TestSecurityRegression` | Auth bypass, secret redaction | Unauthorized access blocked |
| Binary CLI | `TestBinaryCLI_Version` | Binary version command | Outputs version string |
| Binary CLI | `TestBinaryCLI_VersionFlag` | Binary --version flag | Outputs version string |
| Binary CLI | `TestBinaryCLI_Help` | Binary help command | Shows usage info |
| Binary CLI | `TestBinaryCLI_Env` | Binary env command | Shows environment vars |
| Binary CLI | `TestBinaryCLI_PortFlag` | Binary --port flag | Accepted by parser |
| Binary CLI | `TestBinaryCLI_IPFlag` | Binary --ip flag | Accepted by parser |
| Binary CLI | `TestBinaryCLI_VerboseFlag` | Binary --verbose flag | Accepted by parser |
| Binary CLI | `TestBinaryCLI_ConfigFlag` | Binary --config flag | Accepted by parser |
| Binary CLI | `TestBinaryCLI_GlobalFlags` | Multiple flags | Flags combined correctly |
| Binary CLI | `TestBinaryCLI_AssetsExport` | Assets export command | Creates assets directory |

### Handler Tests (`handlers/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Setup | `TestSetupRoutes` | Initial database setup via API | Config created, services loaded |
| API Routes | `TestMainApiRoutes` | Core API endpoints | 200 OK with valid responses |
| Services | `TestServicesRoutes` | Service CRUD operations | Services created/updated/deleted |
| Users | `TestUsersRoutes` | User management | Users created with hashed passwords |
| Users | `TestApiCheckUserToken` | Token validation endpoint | Missing/invalid tokens rejected |
| Groups | `TestGroupsRoutes` | Group management | Groups with services linked |
| Authentication | `TestUnAuthenticatedRoutes` | Auth required endpoints | 401 without credentials |
| CSRF | `TestCSRFProtection` | CSRF token validation | 403 without valid token |
| Rate Limiting | `TestRateLimiting` | Login rate limits | 429 after 5 attempts |
| OAuth | `TestOAuthStateValidation` | OAuth CSRF protection | Invalid state rejected |
| OAuth | `TestValidateGithub` | GitHub user/org validation | Case-insensitive matching works |
| OAuth | `TestValidateGoogle` | Google email/domain validation | Email and Hd field matching |
| OAuth | `TestValidateSlack` | Slack user/email validation | Name and email matching |
| OAuth | `TestOAuthLoginRoutes` | OAuth callback handling | Invalid state returns error |
| Dashboard | `TestDashboardAPIRoutes` | Theme/config endpoints | GET returns valid JSON |
| Dashboard | `TestUnAuthenticatedDashboardRoutes` | Unauth dashboard access | 401/403 without auth |
| Dashboard | `TestThemeAPIRoutes` | Theme save API | Invalid JSON returns 500 |
| Dashboard | `TestSettingsImportExport` | Settings import/export | Export returns JSON, import handles errors |
| JWT | `TestJwtTokenOperations` | JWT set/parse/get/remove | Tokens created and validated |
| Index | `TestIndexRoutes` | Base handler and health check | 404 page and health endpoint work |
| Middleware | `TestGzipMiddleware` | Gzip compression | Compresses when Accept-Encoding set |
| Middleware | `TestGzipResponseWriter` | Gzip writer wrapper | Writes to gzip stream |

### Type Tests (`types/*/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Services | `TestServices` | Service CRUD and checks | HTTP/TCP/GRPC checks work |
| Services | `TestServiceNotifications` | Alert triggering | Notifications fire on failure |
| Services | `TestServiceValidate` | Validation rules | Missing fields rejected |
| Services | `TestServiceValidateEdgeCases` | Edge case validation | Static/cmd exceptions work |
| Services | `TestServiceTypes` | All service types | Each type validates correctly |
| Services | `TestCheckHTTP` | HTTP checks with mock server | Status codes, body regex, headers |
| Services | `TestCheckTCP` | TCP checks with mock listener | Port connections verified |
| Services | `TestCheckCmd` | Command execution checks | Exit codes, stdout/stderr regex |
| Services | `TestCmdConfig` | Command config parsing | JSON config deserialized |
| Services | `TestParseHost` | URL parsing | Host extracted from URLs |
| Services | `TestIsIPv6` | IPv6 detection | IPv6 addresses identified |
| Services | `TestMakeCmdEnv` | Environment building | Env vars merged correctly |
| Services | `TestHTTPRedirects` | Redirect handling | Follow/no-follow works |
| Services | `TestServiceLatencyMeasurement` | Latency timing | Latency measured accurately |
| Hits | `TestHit_BeforeCreate` | CreatedAt auto-set | Sets time when zero, preserves existing |
| Hits | `TestHit_CRUD` | Hit create/update/delete | All operations succeed |
| Core | `TestNew` | Core initialization | Version/commit set, Started time set |
| Core | `TestCore_TableName` | Table name | Returns "core" |
| Core | `TestExample` | Example core creation | Creates valid test core |
| Metrics | `TestConvert` | Label conversion | Converts various types to strings |
| Metrics | `TestHisto` | Histogram metrics | Duration/bytes methods work |
| Metrics | `TestGauge` | Gauge metrics | Status code/online methods work |
| Metrics | `TestInc` | Counter increment | Failure/success counters work |
| Metrics | `TestAdd` | Counter add | Adds values to counters |
| Metrics | `TestTimer` | Timer observer | Returns valid observer |
| Metrics | `TestServiceTimer` | Service timer | Returns valid observer |
| Notifications | `TestNotification_Name` | Name formatting | Converts method to lowercase slug |
| Notifications | `TestNotification_LastSentDur` | Duration calculation | Returns time since last sent |
| Notifications | `TestNotification_CanSend` | Send eligibility | Checks enabled/limits/timeout |
| Notifications | `TestNotification_GetValue` | Field getter | Returns correct field values |
| Services | `TestServiceUptime` | Uptime duration | Calculates from LastOffline |
| Services | `TestServiceDowntime` | Downtime duration | Calculates from LastOnline |
| Services | `TestServiceOrderSort` | Service sorting | Sorts by Order field |
| Services | `TestServiceDuration` | Interval to duration | Converts Interval to time.Duration |
| Services | `TestServiceHash` | Service hashing | Consistent hash for same service |
| Services | `TestServiceRequiresTLS` | TLS requirement check | Detects SMTP/IMAP TLS ports |
| Services | `TestByTimeSort` | Time series sorting | Sorts by timestamp |
| Services | `TestCheckSmtp` | SMTP service check | Mock SMTP server protocol |
| Services | `TestCheckImap` | IMAP service check | Mock IMAP server protocol |
| Services | `TestFindWithRelations` | Preloaded relations | Loads incidents/messages/checkins |
| Services | `TestAllWithRelations` | Batch relation loading | Preloads all services |
| Services | `TestClearCache` | Cache clearing | Empties allServices map |
| Users | `TestUsers` | User CRUD | Password hashing, API keys |
| Groups | `TestGroups` | Group-service relationships | Services grouped correctly |
| Failures | `TestFailures` | Failure recording | Failures logged with details |
| Checkins | `TestCheckins` | Checkin tracking | Heartbeats recorded |
| Messages | `TestMessages` | Announcement system | Messages with date ranges |
| Incidents | `TestIncidents` | Incident management | Incidents with updates |

### Notifier Tests (`notifiers/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Webhook | `TestWebhookNotifier` | Webhook with HTTP mock | Sends to mock server correctly |
| Webhook | `TestWebhookNotifier/webhooker_CanSend` | Send eligibility | Returns true when enabled |
| Webhook | `TestWebhookNotifier/webhooker_OnFailure` | Failure notification | Sends failure payload |
| Webhook | `TestWebhookNotifier/webhooker_OnSuccess` | Success notification | Sends success payload |
| Webhook | `TestWebhookNotifier/webhooker_OnTest` | Test notification | Returns response |
| Webhook | `TestWebhookNotifier/webhooker_with_custom_headers` | Custom headers | Headers sent correctly |
| Slack | `TestSlackNotifierMock` | Slack with mock server | Posts JSON to webhook |
| Discord | `TestDiscordNotifierMock` | Discord with mock server | Posts to webhook, expects 204 |
| Mattermost | `TestMattermostNotifierMock` | Mattermost with mock | OnFailure/OnSuccess/OnTest work |
| Gotify | `TestGotifyNotifierMock` | Gotify with mock server | X-Gotify-Key header validated |
| Pushover | `TestPushoverNotifierMock` | Pushover with mock | Select/Valid/OnSave work |
| Twilio | `TestTwilioNotifierMock` | Twilio with mock server | Select/Valid/OnSave work |
| Telegram | `TestTelegramNotifierMock` | Telegram with mock | Select/Valid work |
| Line Notify | `TestLineNotifyMock` | Line Notify with mock | Select/Valid/OnSave work |
| Amazon SNS | `TestAmazonSNSNotifierMock` | AWS SNS mock | Select/Valid/OnSave work |
| Mobile | `TestMobileNotifierMock` | Mobile push mock | Select/Valid/OnSave work |
| Email | `TestEmailNotifierMock` | Email basic tests | Select/Valid/OnSave/CanSend |
| Email | `TestEmailWithMockSMTPServer` | Email with mock SMTP | Full SMTP protocol exchange |
| Command | `TestCommandNotifier` | Command notifier | Executes shell commands |
| Command | `TestRunCommand` | Command execution | Runs echo, handles errors |
| Template | `TestReplaceTemplate` | Template substitution | Service/failure vars replaced |
| Pushover | `TestPushover_Select` | Priority mapping | Priority strings converted |

### Database Tests (`database/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Connection | `TestOpenTester` | In-memory SQLite | Connection established |
| Connection | `TestOpenwSQLite` | SQLite via Openw | Database type detected |
| Connection | `TestOpenwInvalidDialect` | Empty dialect fallback | Defaults to SQLite |
| Wrapper | `TestWrap` | Wrap existing gorm.DB | Type preserved |
| Global | `TestGetSet` | Global DB accessor | Set/Get work correctly |
| CRUD | `TestDbCRUDOperations` | Create/Read/Update/Delete | All operations succeed |
| Queries | `TestDbQueryMethods` | Where/Limit/Offset/Order | Results filtered correctly |
| Config | `TestDbChunkSize` | Batch size per dialect | SQLite=100, MySQL/Postgres=3000 |
| Schema | `TestDbHasTable` | Table existence check | True after AutoMigrate |
| Schema | `TestDbHasIndex` | Index existence check | True after AddIndex |
| Transaction | `TestDbTransaction` | Commit/Rollback | Data persists/reverts |
| Time | `TestDbTimeQueries` | Since/Between methods | Time-based filtering works |
| Raw | `TestDbRawAndExec` | Raw SQL execution | Direct SQL succeeds |
| Results | `TestDbRowsAffected` | Affected row count | Correct count returned |
| Errors | `TestDbError` | Query error handling | Errors propagated |
| Errors | `TestDbRecordNotFound` | Missing record detection | RecordNotFound is true |
| Logging | `TestDbLogMode` | Log mode toggle | No panic on toggle |
| Scopes | `TestDbScopes` | Query scopes | Scopes applied correctly |

### Security Tests (`handlers/*_test.go`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| CSRF | `TestCSRFTokenGeneration` | Token generation | Unique tokens per request |
| CSRF | `TestCSRFTokenValidation` | Token verification | Invalid tokens rejected |
| CSRF | `TestCSRFCookieMatch` | Double-submit pattern | Header matches cookie |
| Rate Limit | `TestRateLimitAllow` | Under limit | Requests allowed |
| Rate Limit | `TestRateLimitBlock` | Over limit | 429 returned |
| Rate Limit | `TestRateLimitReset` | Window expiry | Limit resets after window |
| OAuth | `TestGenerateOAuthState` | State token creation | Cryptographically random |
| OAuth | `TestValidateOAuthState` | State verification | One-time use enforced |
| OAuth | `TestValidateOAuthStateEmpty` | Empty state rejection | Returns false |
| OAuth | `TestValidateOAuthStateInvalid` | Invalid state rejection | Returns false |
| OAuth | `TestValidateOAuthStateExpired` | Expired state rejection | Returns false |
| OAuth | `TestCleanupExpiredOAuthStates` | State cleanup | Expired states removed |
| OAuth | `TestOAuthRoutes` | OAuth API endpoints | Config saved/retrieved |

### CMD Tests (`cmd/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Directory | `TestStatpingDirectory` | STATPING_DIR setup | Directory path set correctly |
| CLI Commands | `TestEnvCLI` | Env command output | Environment variables listed |
| CLI Commands | `TestVersionCLI` | Version command | Version string output |
| CLI Commands | `TestAssetsCLI` | Assets export | Assets directory created |
| CLI Commands | `TestHelpCLI` | Help command | Usage info displayed |
| Flags | `TestFlagParsing` | --port, --ip, --verbose, --config | Flags parsed correctly |
| Flags | `TestFlagDefaults` | Default flag values | port=8080, ip=0.0.0.0, verbose=2 |
| Flags | `TestInvalidFlagHandling` | Unknown/invalid flags | Error returned |
| Subcommands | `TestSubcommandRouting` | version, help, env routing | Correct subcommand executed |
| Help | `TestHelpTextOutput` | Help text content | Usage, commands, flags sections |
| Environment | `TestEnvironmentVariableHandling` | STATPING_DIR, API_SECRET | Env vars accessible |
| Config | `TestConfigFileHandling` | Config file paths | Absolute, relative, spaces handled |
| Validation | `TestImportCommandValidation` | Import requires file arg | Error without file |
| Metadata | `TestRootCmdFlags` | Root command flags | All flags registered |
| Metadata | `TestSubCommands` | All subcommands | 10 subcommands registered |

### Config Tests (`types/configs/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| YAML Loading | `TestLoadConfigs_ValidYAML` | Basic SQLite config | Config fields populated |
| YAML Loading | `TestLoadConfigs_PostgresYAML` | Postgres-specific fields | Host, port, SSL parsed |
| YAML Loading | `TestLoadConfigs_MySQLYAML` | MySQL config | MySQL fields parsed |
| YAML Loading | `TestLoadConfigs_WithLetsEncrypt` | Let's Encrypt settings | Domain, email parsed |
| Env Override | `TestLoadConfigs_EnvOverridesYAML` | DB_CONN env override | Env takes precedence |
| Env Override | `TestLoadConfigs_SqliteEnvVariants` | sqlite/sqlite3 normalization | Both variants work |
| Defaults | `TestDefaultValues` | Params defaults | Sensible defaults set |
| Defaults | `TestDbConfig_DefaultConnectionPoolSettings` | Connection pool defaults | MaxOpen, MaxIdle set |
| Validation | `TestLoadConfigs_EmptyConnection_SetupMode` | Empty connection | Setup mode triggered |
| Validation | `TestDbConfig_ValidConnectionTypes` | Valid connection types | sqlite, mysql, postgres |
| Connection | `TestConnectionString_Memory` | Memory DSN | file::memory:?cache=shared |
| Connection | `TestConnectionString_SQLite` | SQLite file path | Correct path generated |
| Connection | `TestConnectionString_MySQL` | MySQL DSN | user:pass@tcp(host:port)/db |
| Connection | `TestConnectionString_Postgres` | Postgres DSN | host=x port=y format |
| Connection | `TestConnectionString_PostgresSSLModes` | SSL modes | disable, require, verify-ca |
| Edge Cases | `TestLoadConfigs_MissingFile` | Missing config file | Error returned |
| Edge Cases | `TestLoadConfigs_MalformedYAML` | Invalid YAML | Parse error |
| Edge Cases | `TestLoadConfigs_EmptyFile` | Empty config file | Defaults used |
| Persistence | `TestSaveConfig` | Config file creation | YAML written correctly |
| Persistence | `TestUpdateConfig` | Config update | Fields updated |

### Notification Tests (`types/notifications/`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Rate Limiting | `TestNotification_CanSend` | Send eligibility | Checks enabled/limits/timeout |
| Rate Limiting | `TestNotification_CanSend_EnabledDisabledState` | Toggle enabled state | State affects CanSend |
| Rate Limiting | `TestNotification_CanSend_LimitsHandling` | Zero/negative limits | Boundary conditions |
| Rate Limiting | `TestNotification_CanSend_RateLimitReset` | 60-minute reset window | Count decrements after timeout |
| Last Sent | `TestNotification_LastSentDur` | Duration calculation | Time since last sent |
| Last Sent | `TestNotification_LastSentDur_EdgeCases` | Zero/future time | Edge cases handled |
| Values | `TestNotification_Values` | Field extraction | All fields returned |
| Values | `TestNotification_Values_EmptyFields` | Empty fields | Null/empty handled |
| Update | `TestNotification_UpdateFields` | Field updates from source | Fields copied correctly |
| Sorting | `TestNotificationOrder_Sorting` | Sort by ID | Sorted correctly |
| Sorting | `TestNotificationOrder_Empty` | Empty collection | No panic |
| Logging | `TestNotificationLog` | Success/failure logs | Log entries created |
| Forms | `TestNotificationForm` | Form field structure | Fields accessible |
| Database | `TestNotification_Create` | Create notification | ID assigned |
| Database | `TestNotification_Update` | Update notification | Fields persisted |
| Database | `TestNotification_Find` | Find by ID | Notification retrieved |
| Database | `TestNotification_All` | List all | All notifications returned |

### Notifier Edge Case Tests (`notifiers/edge_cases_test.go`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| Timeouts | `TestWebhookNetworkTimeout` | 10s timeout with slow server | Times out without hanging |
| Timeouts | `TestSlackNetworkTimeout` | Slack timeout handling | Timeout error returned |
| Malformed | `TestWebhookMalformedResponses` | Invalid JSON, 500, 404 | Errors handled gracefully |
| Malformed | `TestSlackMalformedResponses` | Slack non-"ok" response | Validation works |
| Retry | `TestWebhookRetryBehavior` | Multiple consecutive failures | Request count accurate |
| Retry | `TestWebhookTransientFailureRecovery` | Server recovers | Recovery detected |
| Rate Limit | `TestWebhookRateLimitResponse` | 429 with Retry-After | Rate limit handled |
| Rate Limit | `TestDiscordRateLimitResponse` | Discord rate limit format | Discord-specific handling |
| Null Fields | `TestWebhookEmptyNullFields` | Empty Host, null data | No panic on null |
| Null Fields | `TestSlackEmptyNullFields` | Empty Slack fields | Defaults applied |
| Templates | `TestTemplateRenderingErrors` | Invalid template syntax | Error not panic |
| Templates | `TestWebhookTemplateRenderingEdgeCases` | Non-existent fields | Safe field access |
| Connection | `TestWebhookConnectionRefused` | Server not running | Connection refused error |
| URL | `TestWebhookInvalidURL` | Malformed URL | URL parse error |
| Payload | `TestWebhookLargePayload` | 100KB payload | Large payloads sent |
| Methods | `TestWebhookHTTPMethods` | GET, POST, PUT, PATCH, DELETE | All methods work |
| Headers | `TestWebhookHostHeaderOverride` | Custom Host header | Header overridden |

### Service Check Edge Case Tests (`types/services/edge_cases_test.go`)

| Logical Group | Test Name | Purpose | Success Criteria |
|---------------|-----------|---------|------------------|
| TLS | `TestTLSCertificateValidation` | Self-signed, expired certs | VerifySSL respected |
| DNS | `TestDNSResolution` | Non-existent domain, timeout | DNS errors handled |
| Redirects | `TestRedirectLoops` | Redirect loop detection | Max redirects enforced |
| Redirects | `TestRedirectLoops/Ping-pong` | Two-server loop | Loop detected |
| Timeouts | `TestConnectionTimeout` | Non-routable IP timeout | Timeout fires |
| Timeouts | `TestConnectionTimeout/SlowHeaders` | Server slow to respond | Header timeout |
| Timeouts | `TestConnectionTimeout/SlowBody` | Slow body streaming | Body timeout |
| Response | `TestLargeResponseBody` | 10MB, 15MB responses | Large responses handled |
| Response | `TestLargeResponseBody/Truncated` | Response > 10MB | Truncated to limit |
| Invalid HTTP | `TestInvalidHTTPResponses` | Garbage response, partial | Invalid HTTP handled |
| Invalid HTTP | `TestInvalidHTTPResponses/CloseImmediately` | Server closes connection | EOF handled |
| IPv6 | `TestIPv6AddressHandling` | IPv6 localhost, formatting | IPv6 addresses work |
| IPv6 | `TestIPv6AddressHandling/ZoneID` | fe80::1%eth0 format | Zone ID parsed |
| URL | `TestURLEdgeCases` | Special chars, auth in URL | URL edge cases |
| Connection | `TestConnectionStates` | Accept but no response | Half-close, RST |
| Headers | `TestHeaderEdgeCases` | Multiple headers, empty values | Custom headers sent |

---

## Code Coverage Report

### Current Coverage by Package

| Package | Coverage | Target | Status |
|---------|----------|--------|--------|
| `types/null` | 98.4% | 80%+ | PASS |
| `types/users` | 95.2% | 80%+ | PASS |
| `types/failures` | 95.1% | 80%+ | PASS |
| `types/notifications` | 95.0% | 80%+ | PASS |
| `types/groups` | 94.4% | 80%+ | PASS |
| `types/messages` | 88.9% | 80%+ | PASS |
| `types/incidents` | 88.8% | 80%+ | PASS |
| `types/hits` | 85.0% | 80%+ | PASS |
| `types/core` | 84.7% | 80%+ | PASS |
| `source` | 81.4% | 80%+ | PASS |
| `types/services` | 80.1% | 80%+ | PASS |
| `database` | 80.1% | 80%+ | PASS |
| `types/checkins` | 80.0% | 80%+ | PASS |
| `utils` | 73.8% | 80%+ | BELOW |
| `handlers` | 68.3% | 80%+ | BELOW |
| `types/metrics` | 64.3% | 80%+ | BELOW |
| `notifiers` | 46.7% | 80%+ | BELOW |
| `types` | 44.4% | 80%+ | BELOW |
| `types/configs` | 41.9% | 80%+ | BELOW |
| `cmd` | 19.1% | 80%+ | BELOW |
| **Total** | **66.4%** | **80%+** | Baseline was ~36% |

**Test Architecture Improvements (July 2026):**
- `handlers/main_test.go` - Per-chain isolation with StopAll/ClearCache
- `types/services/main_test.go` - Per-chain isolation, sets DB for 7 packages
- `types/users/main_test.go` - Per-chain isolation
- `types/groups/main_test.go` - Per-chain isolation
- `types/incidents/main_test.go` - Per-chain isolation (no StopAll due to import cycle)
- `types/services/database.go` - Added `StopAll()` helper for goroutine cleanup
- `types/services/struct.go` - Thread-safe `runningMu` mutex for `Running` channel
- `types/services/methods.go` - Thread-safe getters/setters for `checkpoint`/`sleepDuration`
- Dynamic ID lookups in all test files (replaced hardcoded `1`, `2` with `example.Id`)

**Test Additions (July 2026):**
- `notifiers/*_test.go` - Mock HTTP/SMTP server tests for all notifiers
- `notifiers/edge_cases_test.go` - Timeout, retry, rate limit, template edge cases
- `handlers/oauth_test.go` - OAuth validation tests (GitHub, Google, Slack)
- `handlers/jwt_test.go` - JWT token operations tests
- `handlers/middleware_test.go` - Gzip middleware tests
- `handlers/index_test.go` - Index and health check routes
- `handlers/incidents_test.go` - Dynamic ID helpers (`getFirstServiceID`, `getLatestIncidentID`)
- `handlers/services_test.go` - Dynamic helpers (`getServiceByIndex`, `getPrivateService`)
- `types/services/routine_test.go` - Mock SMTP/IMAP server tests
- `types/services/routine_comprehensive_test.go` - Comprehensive routine coverage
- `types/services/database_test.go` - FindWithRelations and cache tests
- `types/services/methods_test.go` - Service method unit tests
- `types/services/edge_cases_test.go` - TLS, DNS, redirect, timeout edge cases
- `types/notifications/notifications_test.go` - Rate limiting, state, DB operations
- `types/configs/config_test.go` - YAML loading, env overrides, connection strings
- `database/database_test.go` - Transactions, connections, migrations, query builders
- `cmd/cli_test.go` - Flag parsing, subcommands, env vars, help text

### How to Generate Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out -p=1 ./...

# View coverage by function
go tool cover -func=coverage.out

# View coverage in browser (HTML report)
go tool cover -html=coverage.out -o coverage.html

# Get total coverage percentage
go tool cover -func=coverage.out | grep total | awk '{print $3}'
```

### Coverage Requirements
- **Minimum**: 80% coverage required for new code
- **Security-critical code**: 90%+ coverage (auth, CSRF, rate limiting)
- **Integration tests**: 100% of user-facing functionality covered

---

## Realistic Data Simulation

### Test Database Fixture

| Table | Row Count | Purpose |
|-------|-----------|---------|
| `hits` | 607,649 | Performance testing time-range queries |
| `failures` | 10,963 | Failure analytics and grouping |
| `services` | 7 | Multi-service scenarios |
| `users` | 3 | Auth and permission testing |
| `groups` | 3 | Group-based filtering |
| `notifications` | 13 | Notifier configuration |

### Fixture Location
- **Compressed**: `testdata/statping.db.xz` (5.7MB)
- **Uncompressed**: 94MB SQLite database

### Live System Integration
Integration tests use `httptest.Server` with the real router stack:
- Real HTTP requests via `net/http`
- Real database operations via GORM
- Real authentication middleware
- Real CSRF protection

---

## How to Run Tests

### PowerShell (Windows)

```powershell
# Set MinGW compiler
$env:CC = "D:\dev\mingw64\bin\gcc.exe"
$env:NODE_OPTIONS = "--openssl-legacy-provider"

# Run all tests
go test -v -p=1 ./...

# Run with coverage
go test -coverprofile=coverage.out -p=1 ./...

# Run specific package
go test -v ./handlers/...

# Run specific test
go test -v -run TestCSRFProtection ./handlers/...

# Run integration tests only
go test -v ./integration/...

# Run with race detector
go test -race -short ./handlers/... ./types/services/...
```

### Bash (Linux/macOS/Git Bash)

```bash
# Set compiler (Windows only)
export CC=/d/dev/mingw64/bin/gcc.exe  # Windows
# export CC=gcc                        # Linux/macOS

export NODE_OPTIONS=--openssl-legacy-provider

# Run all tests
go test -v -p=1 ./...

# Run with coverage
go test -coverprofile=coverage.out -p=1 ./...

# Run specific package
go test -v ./handlers/...

# Run specific test
go test -v -run TestCSRFProtection ./handlers/...

# Run integration tests only
go test -v ./integration/...

# Run with race detector
go test -race -short ./handlers/... ./types/services/...

# Full quality audit
./code_quality.sh
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Run Tests
  env:
    DB_CONN: sqlite
    GO_ENV: test
  run: |
    go test -v -p=1 -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | grep total
```

---

## Maintenance and Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| `don't use the cygwin compiler` | Wrong GCC on Windows | Set `CC` to MinGW path |
| `testdata/statping.db.xz not found` | Missing fixture | Ensure fixture is committed |
| `TempDir cleanup failed` | DB connection not closed | Add cleanup function with `db.Close()` |
| `CSRF token mismatch` | Cookie not sent | Include cookies in HTTP client |
| `rice-box.go missing` | Frontend not built | Run `make frontend-build` or build manually |

### Adding New Tests

1. **Create test file**: `*_test.go` in appropriate package
2. **Use `t.TempDir()`**: For any file/database operations
3. **Close resources**: Return cleanup function if needed
4. **Update this document**: Add test to the table above
5. **Check coverage**: Ensure 80%+ on new code

### Updating Coverage Stats

After modifying code:

```bash
# Regenerate coverage
go test -coverprofile=coverage.out -p=1 ./...

# Update stats in this file
go tool cover -func=coverage.out | grep total
```

### Test Database Maintenance

To update the test fixture:

```bash
# Export from a running instance
sqlite3 statping.db ".dump" > dump.sql

# Or copy and compress
xz -9 statping.db -c > testdata/statping.db.xz
```

---

## Mock Server Patterns

### HTTP Mock Server (Notifiers)

Used for testing notifiers without external dependencies:

```go
mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    receivedBody = string(body)
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"status":"ok"}`))
}))
defer mockServer.Close()

// Point notifier at mock server
testNotifier.Host = null.NewNullString(mockServer.URL)
```

### SMTP Mock Server (Email/Service Checks)

Used for testing email notifier and SMTP service checks:

```go
listener, _ := net.Listen("tcp", "127.0.0.1:0")
go func() {
    conn, _ := listener.Accept()
    conn.Write([]byte("220 mock.smtp.server ESMTP\r\n"))
    // Handle EHLO, MAIL FROM, RCPT TO, DATA, QUIT
}()
```

### IMAP Mock Server (Service Checks)

Used for testing IMAP service checks:

```go
listener, _ := net.Listen("tcp", "127.0.0.1:0")
go func() {
    conn, _ := listener.Accept()
    conn.Write([]byte("* OK IMAP4rev1 Service Ready\r\n"))
    // Handle CAPABILITY, LOGIN, LOGOUT
}()
```

---

*Last Updated: 2026-07-27*
*Coverage: 66.4% (target: 80%+)*
