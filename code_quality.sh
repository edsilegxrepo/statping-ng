#!/usr/bin/env bash
# Statping-ng Code Quality Audit Suite
# Executes all quality, safety, and security verification tools in optimal order.
set -e

GREEN="\033[32m"
CYAN="\033[36m"
RESET="\033[0m"

log_step() {
  printf "\n%b=========================================================%b\n" "$CYAN" "$RESET"
  printf "%b  [AUDIT] %s%b\n" "$GREEN" "$1" "$RESET"
  printf "%b=========================================================%b\n" "$CYAN" "$RESET"
}

# 1. Shell Script Quality (ShellCheck)
log_step "1/8: Shell Script Linting (shellcheck)"
find . -name "*.sh" -not -path "*/node_modules/*" -exec shellcheck -e SC2329 {} +

# 2. Go Backend Compiler & Linter (go vet)
log_step "2/8: Go Compiler & Linter (go vet)"
go vet ./...

# 3. Go Security AST Scanner (gosec)
log_step "3/8: Go Security AST Analysis (gosec)"
gosec ./...

# 4. Go Vulnerability Scanner (govulncheck)
log_step "4/8: Go Vulnerability Analysis (govulncheck)"
govulncheck ./...

# 5. Go Unit & Integration Test Suite (go test)
log_step "5/8: Go Unit & Integration Tests (go test)"
go test -p=1 ./...

# 6. Frontend Linter (oxlint)
log_step "6/8: Frontend JS/Vue Linter (oxlint)"
oxlint frontend/src

# 7. Frontend Code Quality & Formatting (biome)
log_step "7/8: Frontend Quality & Formatting (biome)"
biome check \
  frontend/src/API.js \
  frontend/src/App.vue \
  frontend/src/codemirror_json.js \
  frontend/src/graphing.js \
  frontend/src/icons.js \
  frontend/src/main.js \
  frontend/src/mixin.js \
  frontend/src/routes.js \
  frontend/src/store.js \
  frontend/src/components \
  frontend/src/forms \
  frontend/src/pages \
  frontend/src/languages

# 8. Supply Chain & Vulnerability Scanner (grype)
log_step "8/8: Supply Chain Vulnerability Scan (grype)"
grype dir:.

printf "\n%b=========================================================%b\n" "$GREEN" "$RESET"
printf "%b  ALL CODE QUALITY & SECURITY CHECKS PASSED SUCCESSFULLY! %b\n" "$GREEN" "$RESET"
printf "%b=========================================================%b\n\n" "$GREEN" "$RESET"
