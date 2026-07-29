#!/usr/bin/env bash
# Statping-ng Code Quality Audit Suite
# Executes all quality, safety, and security verification tools in optimal order.
# Works on Linux, macOS, and Windows (Git Bash/MSYS2/Cygwin with MinGW).
set -e

GREEN="\033[32m"
CYAN="\033[36m"
YELLOW="\033[33m"
RESET="\033[0m"

# Use MinGW GCC on Windows to avoid Cygwin compiler issues
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
  MINGW_BIN=""
  MINGW_BIN="$(cygpath -u "$CC" 2>/dev/null | xargs dirname 2>/dev/null)" || MINGW_BIN='/d/dev/mingw64/bin'
  export PATH="$MINGW_BIN:$PATH"
  CC_PATH=""
  CC_PATH="$(cygpath -u "$CC" 2>/dev/null)" || CC_PATH="$MINGW_BIN/gcc.exe"
  export CC="$CC_PATH"
  export CXX="${CC_PATH/gcc/g++}"
fi

# Node.js legacy OpenSSL provider for older webpack/babel tooling
export NODE_OPTIONS=--openssl-legacy-provider

log_step() {
  printf "\n%b=========================================================%b\n" "$CYAN" "$RESET"
  printf "%b  [AUDIT] %s%b\n" "$GREEN" "$1" "$RESET"
  printf "%b=========================================================%b\n" "$CYAN" "$RESET"
}

# Build frontend assets and embed them for go.rice (required for integration tests)
build_frontend() {
  log_step "0/10: Building Frontend Assets"
  rm -rf source/dist frontend/dist
  cd frontend && npm install --engine-strict=false && npm run build && cd ..
  cp -r frontend/dist source/
  cp -r frontend/src/assets/scss source/dist/
  cp frontend/public/robots.txt source/dist/
  cd source && rice embed-go && cd ..
  echo "Frontend assets built and embedded in source/rice-box.go"
}

# Decompress test database fixture if needed
setup_test_db() {
  if [[ -f "testdata/statping.db.xz" ]] && [[ ! -f "testdata/statping.db" ]]; then
    log_step "0/10: Decompressing Test Database Fixture"
    xz -dkf testdata/statping.db.xz
  fi
}

# Check if frontend needs building
if [[ ! -f "source/rice-box.go" ]] || [[ ! -d "source/dist" ]]; then
  build_frontend
fi

# Decompress test database if needed (for integration tests)
setup_test_db

# 1. Shell Script Quality (ShellCheck)
log_step "1/10: Shell Script Linting (shellcheck)"
find . -name "*.sh" -not -path "*/node_modules/*" -exec shellcheck -e SC2329 {} +

# 2. Go Backend Compiler & Linter (go vet)
log_step "2/10: Go Compiler & Linter (go vet)"
go vet ./...

# 3. Go Security AST Scanner (gosec)
log_step "3/10: Go Security AST Analysis (gosec)"
gosec ./...

# 4. Go Vulnerability Scanner (govulncheck)
log_step "4/10: Go Vulnerability Analysis (govulncheck)"
govulncheck ./...

# 5. Go Race Condition Detector (go test -race)
log_step "5/10: Go Race Condition Detection (go test -race)"
go test -race -short ./handlers/... ./types/services/...

# 6. Go Unit & Integration Test Suite (go test)
log_step "6/10: Go Unit & Integration Tests (go test)"
go test -p=1 ./...

# 7. Frontend Dependencies Install
log_step "7/10: Frontend Dependencies (npm install)"
cd frontend && npm install --engine-strict=false && cd ..

# 8. Frontend Linter (oxlint)
log_step "8/10: Frontend JS/Vue Linter (oxlint)"
oxlint frontend/src

# 9. Frontend Code Quality & Formatting (biome)
log_step "9/10: Frontend Quality & Formatting (biome)"
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

# 10. Supply Chain & Vulnerability Scanner (grype)
log_step "10/10: Supply Chain Vulnerability Scan (grype)"
grype dir:.

printf "\n%b=========================================================%b\n" "$GREEN" "$RESET"
printf "%b  ALL CODE QUALITY & SECURITY CHECKS PASSED SUCCESSFULLY! %b\n" "$GREEN" "$RESET"
printf "%b=========================================================%b\n\n" "$GREEN" "$RESET"

# Release Info
printf "%b  Release Info:%b\n" "$YELLOW" "$RESET"
printf "  - Version: %s\n" "$(cat version.txt 2>/dev/null || git describe --tags 2>/dev/null || echo 'dev')"
printf "  - Commit:  %s\n" "$(cat commit.txt 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
printf "  - Go:      %s\n" "$(go version | awk '{print $3}')"
printf "  - Node:    %s\n" "$(node --version)"
printf "\n"
