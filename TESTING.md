# Statping-ng Test Suite Documentation

## Architecture, Design and Principles

### Test Philosophy
- **Isolation**: Each test uses `t.TempDir()` for database and file operations, ensuring no cross-test contamination
- **Realistic Data**: Integration tests use a 94MB production-scale database fixture (607K hits, 11K failures)
- **Cross-Platform**: Tests run identically on Windows (MinGW) and Linux
- **Security-First**: Security regression tests verify authentication, authorization, CSRF, and rate limiting

### Test Layers
```
┌─────────────────────────────────────────────────────────────┐
│                    Integration Tests                        │
│         (Full HTTP stack, real database, live APIs)         │
├─────────────────────────────────────────────────────────────┤
│                     Handler Tests                           │
│           (HTTP handlers, routing, middleware)              │
├─────────────────────────────────────────────────────────────┤
│                      Type Tests                             │
│        (Services, Users, Groups, Checkins, etc.)            │
├─────────────────────────────────────────────────────────────┤
│                    Notifier Tests                           │
│         (Slack, Discord, Email, Webhook, etc.)              │
├─────────────────────────────────────────────────────────────┤
│                     Utility Tests                           │
│              (Crypto, HTTP, File operations)                │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions
1. **File-Based SQLite for Unit Tests**: `database.OpenTester(tmpDir)` creates isolated file-based databases in temp directories, avoiding in-memory SQLite connection isolation issues
2. **Fixture Database for Integration**: `testdata/statping.db.xz` provides realistic data volume
3. **TestMain for Package Setup**: Handler tests use `TestMain` to configure shared temp directories with `sync.Once` to ensure single setup
4. **Cleanup Functions**: Database connections are properly closed before temp directory cleanup
5. **Cache Clearing**: `services.ClearCache()` ensures fresh state before test database setup

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
- Tests must run with `-p=1` flag (sequential) to avoid database conflicts
- Integration tests require `testdata/statping.db.xz` fixture
- Frontend assets must be built before handler tests (`source/rice-box.go`)
- Windows requires MinGW, not Cygwin GCC

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
| Groups | `TestGroupsRoutes` | Group management | Groups with services linked |
| Authentication | `TestUnAuthenticatedRoutes` | Auth required endpoints | 401 without credentials |
| CSRF | `TestCSRFProtection` | CSRF token validation | 403 without valid token |
| Rate Limiting | `TestRateLimiting` | Login rate limits | 429 after 5 attempts |
| OAuth | `TestOAuthStateValidation` | OAuth CSRF protection | Invalid state rejected |
| Dashboard | `TestDashboardAPIRoutes` | Theme/config endpoints | GET returns valid JSON |
| Dashboard | `TestUnAuthenticatedDashboardRoutes` | Unauth dashboard access | 401/403 without auth |

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

---

## Code Coverage Report

### Current Coverage by Package

| Package | Coverage | Target | Notes |
|---------|----------|--------|-------|
| `handlers` | 59.4% | 80%+ | HTTP handlers |
| `source` | 76.9% | 80%+ | Asset management |
| `types/groups` | 88.9% | 80%+ | Group types |
| `types/messages` | 88.9% | 80%+ | Message types |
| `types/incidents` | 72.5% | 80%+ | Incident types |
| `types/metrics` | 64.3% | 80%+ | Prometheus metrics |
| `types/services` | 57.8% | 80%+ | Service types and checks |
| `types/users` | 56.0% | 80%+ | User types |
| `utils` | 55.8% | 80%+ | Utilities |
| `types/checkins` | 46.7% | 80%+ | Checkin types |
| `types/failures` | 42.7% | 80%+ | Failure types |
| `types/notifications` | 28.8% | 80%+ | Notification types |
| `database` | 27.4% | 80%+ | Database abstraction |
| `notifiers` | 18.3% | 80%+ | Notifier implementations |
| `types/configs` | 15.5% | 80%+ | Config types |
| **Total** | **45.2%** | **80%+** | Baseline was ~36% |

**Recent Test Additions:**
- `database/database_test.go` - 18 comprehensive database layer tests
- `types/services/validation_test.go` - Service validation with edge cases
- `types/services/check_test.go` - HTTP/TCP/Command check tests with mocks

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

*Last Updated: 2026-07-24*
*Coverage: 45.2% (target: 80%+)*
