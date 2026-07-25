# Statping-ng

**Status Page & Monitoring Server**

An easy-to-use, self-hosted status page for monitoring websites, APIs, and applications. Statping-ng automatically checks your services and renders a beautiful status page with uptime statistics, failure tracking, and multi-channel notifications.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Security Assessment](#security-assessment)
- [Code Quality Assessment](#code-quality-assessment)
- [Installation](#installation)
- [Command Line Arguments](#command-line-arguments)
- [Environment Variables](#environment-variables)
- [Deployment Examples](#deployment-examples)
- [API Examples](#api-examples)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Statping-ng monitors your services using multiple protocols:

| Protocol | Use Case |
|----------|----------|
| **HTTP/HTTPS** | Websites, REST APIs, health endpoints |
| **TCP** | Databases, message queues, custom services |
| **UDP** | DNS servers, game servers |
| **ICMP** | Network connectivity (ping) |
| **gRPC** | Microservices with gRPC health checks |
| **SMTP/IMAP** | Email servers |
| **Command** | Custom scripts and local checks |
| **Static** | Manual status updates via API |

### Objectives

1. **Reliability**: Status page remains online even when monitored services fail
2. **Simplicity**: Single binary deployment with embedded assets
3. **Flexibility**: Support for SQLite, MySQL, PostgreSQL, and SQL Server
4. **Extensibility**: Plugin-based notifier architecture
5. **Security**: Modern authentication, CSRF protection, and encrypted communications

---

## Features

- Real-time service monitoring with configurable intervals
- Public and private status pages
- Service grouping and ordering
- Incident management with updates
- Scheduled maintenance messages
- 12+ notification channels (Slack, Discord, Email, Telegram, etc.)
- Prometheus metrics exporter (`/metrics`)
- OAuth login (GitHub, Google, Slack, custom OIDC)
- Custom SASS theming
- Multi-language support (i18n)
- Let's Encrypt automatic TLS certificates
- REST API for automation
- Docker and Kubernetes ready

---

## Security Assessment

### Encryption in Transit

| Component | Implementation | Status |
|-----------|----------------|--------|
| **HTTPS** | TLS 1.2+ via Let's Encrypt or reverse proxy | Supported |
| **Database** | TLS connections configurable per dialect | Supported |
| **API Calls** | All sensitive endpoints require HTTPS in production | Recommended |
| **Notifier Webhooks** | HTTPS URLs enforced for external services | Configurable |

**TLS Configuration:**
```bash
# Automatic Let's Encrypt
LETSENCRYPT_ENABLE=true
LETSENCRYPT_HOST=status.example.com
LETSENCRYPT_EMAIL=admin@example.com

# Or use reverse proxy (nginx, Traefik, Caddy)
```

### Secret Management

| Secret | Storage | Protection |
|--------|---------|------------|
| **Database credentials** | `config.yml` or environment variables | File permissions (0600) |
| **API Secret** | Database `core` table | Auto-generated SHA256, rotatable |
| **JWT Signing Key** | Derived from API Secret | In-memory only |
| **OAuth Client Secrets** | Database `oauth` table | Encrypted at rest (recommended) |
| **Notifier Credentials** | Database `notifications` table | Per-notifier encryption (varies) |

**Best Practices:**
- Use environment variables for secrets in containerized deployments
- Never commit `config.yml` with credentials to version control
- Rotate API Secret periodically via Dashboard > Settings > Renew API Keys
- Use external secret managers (Vault, AWS Secrets Manager) in production

### Authentication Configuration

| Method | Implementation | Security Level |
|--------|----------------|----------------|
| **Local Auth** | bcrypt (cost 14) password hashing | High |
| **JWT Tokens** | HMAC-SHA256, HTTP-only cookies, 72h expiry | High |
| **OAuth 2.0** | GitHub, Google, Slack, custom OIDC | High |
| **API Key** | Bearer token in `Authorization` header | Medium |
| **Basic Auth** | Optional HTTP Basic (via `AUTH_USERNAME`/`AUTH_PASSWORD`) | Medium |

**CSRF Protection:**
- Double-submit cookie pattern
- `X-CSRF-Token` header required for all mutating requests
- 10-minute state token TTL for OAuth flows

**Rate Limiting:**
- Login endpoint: 5 attempts per 5 minutes per IP
- Configurable via middleware

### Role-Based Access Control (RBAC)

| Role | Permissions |
|------|-------------|
| **Public** | View public services, groups, messages, incidents |
| **User** | Above + view private services, limited dashboard access |
| **Admin** | Full access: CRUD all entities, settings, notifiers, users |

**Field-Level Scoping:**
Private fields (credentials, internal IPs) are stripped from API responses for non-admin users using struct tags:
```go
Domain string `json:"domain" scope:"user,admin" private:"true"`
```

### Library Security

**Go Dependencies (Backend):**

| Library | Version | CVE Status | Notes |
|---------|---------|------------|-------|
| `golang.org/x/crypto` | 0.54.0 | Clean | bcrypt, secure random |
| `github.com/golang-jwt/jwt/v5` | 5.3.1 | Clean | JWT implementation |
| `gorm.io/gorm` | 1.31.2 | Clean | Parameterized queries |
| `github.com/gorilla/mux` | 1.8.1 | Clean | HTTP router |
| `github.com/gorilla/securecookie` | 1.1.2 | Clean | Cookie encryption |

**NPM Dependencies (Frontend):**

| Library | Version | CVE Status | Notes |
|---------|---------|------------|-------|
| `vue` | 2.7.16 | Clean | UI framework |
| `axios` | 0.33.0 | Clean | HTTP client |
| `dompurify` | 3.0.6 | Clean | XSS sanitization |

**Overrides Applied:**
```json
{
  "lodash": "^4.17.21",
  "minimatch": "^3.1.2",
  "qs": "^6.14.2",
  "express": "^4.20.0"
}
```

Run `npm audit` and `go mod verify` regularly to check for new vulnerabilities.

### Unprivileged Execution

Statping-ng runs as an unprivileged user by default:

```dockerfile
# Docker image runs as non-root
USER statping
```

**Required Capabilities:**
- `CAP_NET_RAW` - Only if using ICMP ping checks
- No other elevated privileges required

**File System:**
- Working directory: `/app` (configurable via `STATPING_DIR`)
- Write access needed for: `config.yml`, `statping.db` (SQLite), `logs/`, `assets/`

---

## Code Quality Assessment

### Architecture

- **Clean separation**: Handlers, business logic, and data access layers
- **Interface-driven**: Database abstraction via `Database` interface
- **Dependency injection**: Services receive database connections explicitly

### Code Review Summary

| Aspect | Assessment | Notes |
|--------|------------|-------|
| **Error Handling** | Good | Wrapped errors with context via `pkg/errors` |
| **Input Validation** | Good | Request validation in handlers, struct tags |
| **SQL Injection** | Protected | GORM parameterized queries throughout |
| **XSS Prevention** | Good | DOMPurify sanitization in frontend |
| **CSRF Protection** | Good | Double-submit cookie pattern |
| **Race Conditions** | Mitigated | `sync.RWMutex` on service cache |
| **Resource Cleanup** | Good | Proper `defer` for DB connections, file handles |
| **Logging** | Good | Structured logging with logrus |

### Best Practices Followed

- Go idioms: error returns, defer cleanup, interface composition
- Vue patterns: Vuex for state, computed properties, single-file components
- RESTful API design with consistent JSON responses
- Comprehensive struct tags for JSON, GORM, validation
- Middleware chain for cross-cutting concerns

### Test Coverage

| Package | Coverage |
|---------|----------|
| `handlers` | 59.4% |
| `types/services` | 57.8% |
| `types/groups` | 88.9% |
| `utils` | 55.8% |
| **Total** | **45.2%** |

Target: 80%+ coverage. See [TESTING.md](TESTING.md) for details.

---

## Installation

### Binary Download

```bash
# Linux (amd64)
curl -L https://github.com/statping-ng/statping-ng/releases/latest/download/statping-linux-amd64.tar.gz | tar xz
chmod +x statping
./statping

# macOS (arm64)
curl -L https://github.com/statping-ng/statping-ng/releases/latest/download/statping-darwin-arm64.tar.gz | tar xz
chmod +x statping
./statping

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/statping-ng/statping-ng/releases/latest/download/statping-windows-amd64.zip" -OutFile statping.zip
Expand-Archive statping.zip -DestinationPath .
.\statping.exe
```

### Docker

```bash
docker run -d \
  --name statping \
  -p 8080:8080 \
  -v statping_data:/app \
  adamboutcher/statping-ng:latest
```

### Build from Source

```bash
# Prerequisites: Go 1.21+, Node.js 18+, Make

git clone https://github.com/statping-ng/statping-ng.git
cd statping-ng

# Build frontend
cd frontend && npm install && npm run build && cd ..

# Embed assets and build binary
rice embed-go -i ./source
go build -o statping ./cmd
```

---

## Command Line Arguments

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--ip` | `-s` | string | `0.0.0.0` | IP address to bind the HTTP server |
| `--port` | `-p` | int | `8080` | Port to listen on |
| `--verbose` | `-v` | int | `2` | Logging verbosity (1=error, 2=info, 3=debug) |
| `--config` | `-c` | string | `./config.yml` | Path to configuration file |
| `--version` | | | | Print version and exit |
| `--help` | `-h` | | | Show help message |

### Subcommands

| Command | Description |
|---------|-------------|
| `statping` | Start the server (default) |
| `statping version` | Print version information |
| `statping env` | Display current configuration |
| `statping assets` | Export embedded assets to `./assets/` for customization |
| `statping sass` | Compile custom SCSS files |
| `statping export` | Export settings to `statping-export.json` |
| `statping import <file>` | Import settings from JSON backup |
| `statping reset` | Delete all data and start fresh |
| `statping once` | Check all services once and exit |
| `statping update` | Update to the latest version |
| `statping systemctl install <path> <port>` | Install systemd service |
| `statping systemctl uninstall` | Remove systemd service |

---

## Environment Variables

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `STATPING_DIR` | `.` (cwd) | Working directory for config, database, logs |
| `SERVER_IP` | `0.0.0.0` | Bind address (same as `--ip`) |
| `SERVER_PORT` | `8080` | Listen port (same as `--port`) |
| `BASE_PATH` | (empty) | URL path prefix (e.g., `/status`) |
| `DISABLE_HTTP` | `false` | Disable HTTP server (metrics/API only) |
| `READ_ONLY` | `false` | Disable all write operations |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_CONN` | `sqlite` | Database type: `sqlite`, `mysql`, `postgres`, `mssql` |
| `DB_HOST` | `localhost` | Database server hostname |
| `DB_PORT` | (varies) | Database port (5432 for postgres, 3306 for mysql) |
| `DB_USER` | (empty) | Database username |
| `DB_PASS` | (empty) | Database password |
| `DB_DATABASE` | `statping` | Database name |
| `DB_DSN` | (empty) | Full DSN string (overrides individual settings) |
| `POSTGRES_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `MAX_OPEN_CONN` | `25` | Maximum open database connections |
| `MAX_IDLE_CONN` | `25` | Maximum idle connections |
| `MAX_LIFE_CONN` | `5m` | Connection maximum lifetime |

### TLS / Let's Encrypt

| Variable | Default | Description |
|----------|---------|-------------|
| `LETSENCRYPT_ENABLE` | `false` | Enable automatic TLS certificates |
| `LETSENCRYPT_HOST` | (empty) | Domain for certificate |
| `LETSENCRYPT_EMAIL` | (empty) | Email for Let's Encrypt notifications |
| `LETSENCRYPT_LOCAL` | `false` | Use Let's Encrypt staging environment |

### Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_USERNAME` | (empty) | HTTP Basic Auth username (optional) |
| `AUTH_PASSWORD` | (empty) | HTTP Basic Auth password (optional) |
| `ADMIN_USER` | `admin` | Default admin username (setup only) |
| `ADMIN_PASSWORD` | (empty) | Default admin password (setup only) |
| `ADMIN_EMAIL` | `info@admin.com` | Default admin email (setup only) |

### Application Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `NAME` | `Platform Monitoring` | Status page title |
| `DESCRIPTION` | `Enterprise Status Page` | Status page description |
| `DOMAIN` | `http://localhost:8080` | Public URL for OAuth callbacks |
| `LANGUAGE` | `en` | Default language |
| `SAMPLE_DATA` | `false` | Load sample services on first run |
| `ALLOW_REPORTS` | `false` | Send anonymous error reports to Sentry |
| `DEMO_MODE` | `false` | Enable demo mode (limited functionality) |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `DISABLE_LOGS` | `false` | Disable all logging |
| `DISABLE_COLORS` | `false` | Disable colored log output |
| `LOGS_MAX_COUNT` | `5` | Maximum number of rotated log files |
| `LOGS_MAX_AGE` | `28` | Days to retain log files |
| `LOGS_MAX_SIZE` | `16` | Maximum log file size in MB |
| `DEBUG` | `false` | Enable debug logging and pprof endpoint |

### Data Retention

| Variable | Default | Description |
|----------|---------|-------------|
| `REMOVE_AFTER` | `8760h` (1 year) | Delete hits/failures older than this |
| `CLEANUP_INTERVAL` | `1h` | How often to run cleanup routine |

---

## Deployment Examples

### Docker Compose with PostgreSQL

```yaml
# docker-compose.yml
version: '3.8'

services:
  statping:
    image: adamboutcher/statping-ng:latest
    container_name: statping
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      DB_CONN: postgres
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: statping
      DB_PASS: secretpassword
      DB_DATABASE: statping
      NAME: "My Status Page"
      DESCRIPTION: "Service Health Dashboard"
      DOMAIN: "https://status.example.com"
    volumes:
      - statping_data:/app
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    container_name: statping-db
    restart: unless-stopped
    environment:
      POSTGRES_USER: statping
      POSTGRES_PASSWORD: secretpassword
      POSTGRES_DB: statping
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  statping_data:
  postgres_data:
```

**Start:**
```bash
docker-compose up -d
```

**Output:**
```
Creating network "statping_default" with the default driver
Creating statping-db ... done
Creating statping    ... done
```

**Verify:**
```bash
curl -s http://localhost:8080/api | jq '.name, .version'
```
```json
"My Status Page"
"0.91.0"
```

### Docker Compose with Automatic SSL

```yaml
# docker-compose-ssl.yml
version: '3.8'

services:
  statping:
    image: adamboutcher/statping-ng:latest
    restart: unless-stopped
    environment:
      DB_CONN: sqlite
      LETSENCRYPT_ENABLE: "true"
      LETSENCRYPT_HOST: status.example.com
      LETSENCRYPT_EMAIL: admin@example.com
      DOMAIN: "https://status.example.com"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - statping_data:/app
      - letsencrypt:/etc/letsencrypt

volumes:
  statping_data:
  letsencrypt:
```

### Kubernetes Deployment

```yaml
# statping-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: statping
  labels:
    app: statping
spec:
  replicas: 1
  selector:
    matchLabels:
      app: statping
  template:
    metadata:
      labels:
        app: statping
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      containers:
      - name: statping
        image: adamboutcher/statping-ng:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_CONN
          value: postgres
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: statping-secrets
              key: db-host
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: statping-secrets
              key: db-user
        - name: DB_PASS
          valueFrom:
            secretKeyRef:
              name: statping-secrets
              key: db-pass
        - name: DB_DATABASE
          value: statping
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        volumeMounts:
        - name: data
          mountPath: /app
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: statping-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: statping
spec:
  selector:
    app: statping
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

### Systemd Service (Linux)

```bash
# Install service
sudo statping systemctl install /opt/statping 8080

# Or manually create service file
sudo tee /etc/systemd/system/statping.service << 'EOF'
[Unit]
Description=Statping-ng Status Page
After=network.target

[Service]
Type=simple
User=statping
Group=statping
WorkingDirectory=/opt/statping
ExecStart=/opt/statping/statping --port 8080
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable statping
sudo systemctl start statping

# Check status
sudo systemctl status statping
```

**Output:**
```
● statping.service - Statping-ng Status Page
     Loaded: loaded (/etc/systemd/system/statping.service; enabled)
     Active: active (running) since Fri 2026-07-25 10:00:00 UTC; 5s ago
   Main PID: 12345 (statping)
      Tasks: 8 (limit: 4096)
     Memory: 45.2M
     CGroup: /system.slice/statping.service
             └─12345 /opt/statping/statping --port 8080
```

---

## API Examples

### Authentication

```bash
# Login and get JWT token
curl -X POST http://localhost:8080/api/login \
  -d "username=admin&password=admin" \
  -c cookies.txt

# Use cookie for authenticated requests
curl -b cookies.txt http://localhost:8080/api/services

# Or use API key
curl -H "Authorization: Bearer YOUR_API_SECRET" \
  http://localhost:8080/api/services
```

### Create a Service

```bash
curl -X POST http://localhost:8080/api/services \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_SECRET" \
  -H "X-CSRF-Token: $(curl -s -c - http://localhost:8080/api/csrf | grep csrf_token | awk '{print $7}')" \
  -d '{
    "name": "Google",
    "domain": "https://google.com",
    "type": "http",
    "method": "GET",
    "expected_status": 200,
    "check_interval": 60,
    "timeout": 30,
    "public": true
  }'
```

**Response:**
```json
{
  "status": "success",
  "type": "service",
  "method": "create",
  "id": 1,
  "output": {
    "id": 1,
    "name": "Google",
    "domain": "https://google.com",
    "type": "http",
    "online": true,
    "latency": 142
  }
}
```

### Get Service Status

```bash
curl -s http://localhost:8080/api/services/1 | jq
```

**Response:**
```json
{
  "id": 1,
  "name": "Google",
  "domain": "https://google.com",
  "type": "http",
  "online": true,
  "latency": 142,
  "online_24_hours": 99.98,
  "online_7_days": 99.95,
  "avg_response": 156,
  "failures_24_hours": 2,
  "last_success": "2026-07-25T10:05:00Z",
  "stats": {
    "failures": 15,
    "hits": 43200,
    "first_hit": "2026-06-25T00:00:00Z"
  }
}
```

### Prometheus Metrics

```bash
curl -H "Authorization: Bearer YOUR_API_SECRET" \
  http://localhost:8080/metrics
```

**Output:**
```
# HELP statping_service_online Service online status
# TYPE statping_service_online gauge
statping_service_online{id="1",name="Google",type="http"} 1
statping_service_online{id="2",name="API Server",type="http"} 1

# HELP statping_service_latency Service response latency in milliseconds
# TYPE statping_service_latency gauge
statping_service_latency{id="1",name="Google"} 142
statping_service_latency{id="2",name="API Server"} 23

# HELP statping_service_failures_total Total service failures
# TYPE statping_service_failures_total counter
statping_service_failures_total{id="1",name="Google"} 15
statping_service_failures_total{id="2",name="API Server"} 3
```

### Prometheus Scrape Config

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'statping'
    bearer_token: 'YOUR_API_SECRET'
    static_configs:
      - targets: ['statping:8080']
    metrics_path: /metrics
    scrape_interval: 30s
```

---

## Testing

Comprehensive test documentation is available in [TESTING.md](TESTING.md), including:

- Test architecture and philosophy
- Running tests locally
- Coverage reports
- Adding new tests
- CI/CD integration

**Quick Start:**
```bash
# Run all tests
go test -v -p=1 ./...

# Run with coverage
go test -coverprofile=coverage.out -p=1 ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Contributing

Contributions are welcome! Please submit Pull Requests to the `dev` branch.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test -p=1 ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

See [ARCHITECTURE.md](ARCHITECTURE.md) for technical details about the codebase structure.

---

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

---

*Documentation last updated: 2026-07-25*
