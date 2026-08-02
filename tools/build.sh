#!/bin/bash
#=============================================================================
# Statping-ng Build & Quality Audit Script
#=============================================================================
#
# A unified build, test, and audit tool for the Statping-ng monitoring system.
# Supports both Windows (MSYS2/Cygwin) and Linux environments.
#
# USAGE:
#   ./tools/build.sh [options]
#
# OPTIONS:
#   (no args)              Quick build: frontend + Go binary
#   --audit                Run full code quality & security audit (8 checks)
#   --test                 Run Go tests with full isolation (serial, no cache)
#   --all                  Build + audit combined
#   --clean                Remove build artifacts (binaries, dist, test data)
#   --clean-all            Full reset including node_modules
#   --extra-scans          Include slow scans (grype supply chain analysis)
#   --update-modules       Update all dependencies (Go + Node)
#   --update-modules=go    Update Go modules only (go get -u, go mod tidy)
#   --update-modules=node  Update Node modules only (npm update, npm audit fix)
#   --help, -h             Show this help message
#
# EXAMPLES:
#   ./tools/build.sh                      # Quick build
#   ./tools/build.sh --all                # Build + full audit
#   ./tools/build.sh --update-modules     # Update all deps, then review changes
#   ./tools/build.sh --clean --all        # Clean slate rebuild with audit
#
# REQUIREMENTS:
#   - Go 1.21+
#   - Node.js 18+ with npm
#   - Optional: gosec, govulncheck, shellcheck, oxlint, biome, grype
#
# OUTPUT:
#   - Binary:   testfiles/statping[.exe]
#   - Frontend: source/dist/
#
#=============================================================================
set -e

#-----------------------------------------------------------------------------
# Terminal Colors
#-----------------------------------------------------------------------------
GREEN="\033[32m"
CYAN="\033[36m"
YELLOW="\033[33m"
RESET="\033[0m"

#-----------------------------------------------------------------------------
# Helper Functions
#-----------------------------------------------------------------------------
log_step() {
  printf "\n%b[%s]%b %s\n" "$CYAN" "$1" "$RESET" "$2"
}

log_success() {
  printf "%b✓%b %s\n" "$GREEN" "$RESET" "$1"
}

#-----------------------------------------------------------------------------
# Platform Detection
#-----------------------------------------------------------------------------
IS_WINDOWS=false
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
  IS_WINDOWS=true
fi

# Set binary extension based on platform
BIN_EXT=""
if $IS_WINDOWS; then
  BIN_EXT=".exe"
fi

#-----------------------------------------------------------------------------
# Argument Parsing
#-----------------------------------------------------------------------------
DO_BUILD=true
DO_AUDIT=false
DO_TEST=false
DO_EXTRA=false
DO_CLEAN=false
DO_CLEAN_ALL=false
DO_UPDATE=""

for arg in "$@"; do
  case $arg in
    --audit)
      DO_BUILD=false
      DO_AUDIT=true
      ;;
    --test)
      DO_BUILD=false
      DO_TEST=true
      ;;
    --all)
      DO_BUILD=true
      DO_AUDIT=true
      ;;
    --clean)
      DO_CLEAN=true
      DO_BUILD=false
      ;;
    --clean-all)
      DO_CLEAN=true
      DO_CLEAN_ALL=true
      DO_BUILD=false
      ;;
    --extra-scans)
      DO_EXTRA=true
      ;;
    --update-modules)
      DO_BUILD=false
      DO_UPDATE="both"
      ;;
    --update-modules=go)
      DO_BUILD=false
      DO_UPDATE="go"
      ;;
    --update-modules=node)
      DO_BUILD=false
      DO_UPDATE="node"
      ;;
    --help | -h)
      echo "Usage: ./tools/build.sh [options]"
      echo "  (no args)       Quick build only"
      echo "  --audit         Run code quality & security checks"
      echo "  --test          Run tests with full isolation (serial, no cache, pollution check)"
      echo "  --all           Build + audit"
      echo "  --clean         Remove build artifacts and data for fresh start"
      echo "  --clean-all     Full reset (includes node_modules)"
      echo "  --extra-scans   Include slow scans (grype supply chain)"
      echo "  --update-modules [=go|=node]  Update dependencies (both if no arg)"
      exit 0
      ;;
  esac
done

#=============================================================================
# CLEAN
#=============================================================================
if $DO_CLEAN; then
  echo "=== Cleaning ==="

  # Build artifacts
  log_clean() { printf "  %b✓%b Removed %s\n" "$GREEN" "$RESET" "$1"; }

  # Clean testfiles/ (build outputs, test artifacts)
  rm -f testfiles/statping{,.exe} && log_clean "testfiles/statping[.exe]"
  rm -f testfiles/statping.db && log_clean "testfiles/statping.db"
  rm -f testfiles/statping.secrets && log_clean "testfiles/statping.secrets"
  rm -f testfiles/config.yml && log_clean "testfiles/config.yml"
  rm -rf testfiles/logs/ && log_clean "testfiles/logs/"
  rm -rf testfiles/assets/ && log_clean "testfiles/assets/"

  # Clean legacy locations (repo root)
  rm -f statping.exe statping.db statping.secrets config.yml 2> /dev/null
  rm -rf logs/ assets/ 2> /dev/null
  rm -rf frontend/dist/ && log_clean "frontend/dist/"
  rm -rf source/dist/assets/ && log_clean "source/dist/assets/"
  rm -f source/dist/index.html && log_clean "source/dist/index.html"

  if $DO_CLEAN_ALL; then
    echo ""
    echo "=== Full Reset ==="
    rm -rf frontend/node_modules/ && log_clean "frontend/node_modules/"
    rm -f frontend/package-lock.json && log_clean "frontend/package-lock.json"
  fi

  echo ""
  echo "Clean complete. Run './tools/build.sh' to rebuild."
  exit 0
fi

#=============================================================================
# UPDATE MODULES
#=============================================================================
if [[ -n "$DO_UPDATE" ]]; then
  echo "=== Updating Dependencies ==="

  if [[ "$DO_UPDATE" == "both" || "$DO_UPDATE" == "go" ]]; then
    log_step "UPDATE" "Updating Go modules..."
    go get -u ./...
    go mod tidy
    log_success "Go modules updated"
  fi

  if [[ "$DO_UPDATE" == "both" || "$DO_UPDATE" == "node" ]]; then
    log_step "UPDATE" "Updating Node modules..."
    cd frontend
    npm outdated || true
    npm update
    npm audit fix || true
    cd ..
    log_success "Node modules updated"
  fi

  echo ""
  echo "Update complete. Review changes with 'git diff go.mod go.sum frontend/package*.json'"
  exit 0
fi

#-----------------------------------------------------------------------------
# Windows Compiler Setup (MinGW)
# Avoids Cygwin compiler issues with CGO
#-----------------------------------------------------------------------------
if $IS_WINDOWS; then
  MINGW_BIN=""
  MINGW_BIN="$(cygpath -u "$CC" 2> /dev/null | xargs dirname 2> /dev/null)" || MINGW_BIN='/d/dev/mingw64/bin'
  export PATH="$MINGW_BIN:$PATH"
  CC_PATH=""
  CC_PATH="$(cygpath -u "$CC" 2> /dev/null)" || CC_PATH="$MINGW_BIN/gcc.exe"
  export CC="$CC_PATH"
  export CXX="${CC_PATH/gcc/g++}"
fi

#=============================================================================
# BUILD
#=============================================================================
if $DO_BUILD; then
  echo "=== Building Statping ==="

  # Kill all running statping instances
  log_step "BUILD" "Stopping any running statping..."
  if $IS_WINDOWS; then
    taskkill //F //IM statping.exe 2> /dev/null || true
  else
    pkill -f statping 2> /dev/null || true
  fi

  # Build frontend
  log_step "BUILD" "Building frontend..."
  cd frontend
  npm run build
  cd ..

  # Sync built assets to source/dist
  log_step "BUILD" "Syncing to source/dist..."
  rm -rf source/dist/assets source/dist/index.html
  cp -r frontend/dist/assets source/dist/
  cp frontend/dist/index.html source/dist/

  # Update base.gohtml with correct asset hashes
  INDEX_JS=$(basename source/dist/assets/index-*.js)
  INDEX_CSS=$(basename source/dist/assets/index-*.css)
  VENDOR_JS=$(basename source/dist/assets/vendor-*.js 2> /dev/null || echo "")

  sed -i "s|assets/index-[^\"]*\.js|assets/${INDEX_JS}|g" source/dist/base.gohtml
  sed -i "s|assets/index-[^\"]*\.css|assets/${INDEX_CSS}|g" source/dist/base.gohtml
  if [[ -n "$VENDOR_JS" ]]; then
    sed -i "s|assets/vendor-[^\"]*\.js|assets/${VENDOR_JS}|g" source/dist/base.gohtml
  fi

  echo "  JS:  ${INDEX_JS}"
  echo "  CSS: ${INDEX_CSS}"

  # Build Go binary to testfiles/
  log_step "BUILD" "Building Go binary..."
  mkdir -p testfiles

  # Get version and commit for ldflags
  VERSION=$(cat version.txt 2> /dev/null || echo "dev")
  COMMIT=$(git rev-parse --short=7 HEAD 2> /dev/null || echo "unknown")
  LDFLAGS="-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT}"

  go build -ldflags "$LDFLAGS" -o "testfiles/statping${BIN_EXT}" ./cmd

  log_success "Build complete (binary: testfiles/statping${BIN_EXT})"
fi

#=============================================================================
# TEST (full isolation)
#=============================================================================
if $DO_TEST; then
  log_step "TEST" "Running tests with full isolation..."

  # Clean any stale test artifacts from repo root
  rm -f statping.db statping.log config.yml statping_config.yml 2> /dev/null || true

  # Run tests: serial packages (-p=1), no cache (-count=1), extended timeout
  go test ./... -p=1 -count=1 -timeout=600s

  # Verify no pollution
  if ls statping.db statping.log config.yml statping_config.yml 2> /dev/null; then
    printf "%bWARNING:%b Test pollution detected - files written to repo root\n" "$YELLOW" "$RESET"
    exit 1
  fi

  log_success "All tests passed"
fi

#=============================================================================
# FULL AUDIT
#=============================================================================
if $DO_AUDIT; then
  echo ""
  echo "=== Code Quality & Security Audit ==="

  # Decompress test database fixture if needed
  if [[ -f "testdata/statping.db.xz" ]] && [[ ! -f "testdata/statping.db" ]]; then
    log_step "SETUP" "Decompressing test database fixture..."
    xz -dkf testdata/statping.db.xz
  fi

  # 1. Shell Script Quality
  log_step "1/8" "Shell script linting (shellcheck)..."
  find . -name "*.sh" -not -path "*/node_modules/*" -exec shellcheck -e SC2329 {} +
  log_success "Shell scripts OK"

  # 2. Go vet
  log_step "2/8" "Go compiler & linter (go vet)..."
  go vet ./...
  log_success "Go vet OK"

  # 3. Go Security Scanner
  log_step "3/8" "Go security analysis (gosec)..."
  if command -v gosec &> /dev/null; then
    gosec -quiet ./...
    log_success "Gosec OK"
  else
    echo "  (skipped - gosec not installed)"
  fi

  # 4. Go Vulnerability Scanner
  log_step "4/8" "Go vulnerability analysis (govulncheck)..."
  if command -v govulncheck &> /dev/null; then
    govulncheck ./...
    log_success "Govulncheck OK"
  else
    echo "  (skipped - govulncheck not installed)"
  fi

  # 5. Go Race Detector
  log_step "5/8" "Go race condition detection..."
  go test -race -short ./handlers/... ./types/services/...
  log_success "Race detection OK"

  # 6. Go Tests
  log_step "6/8" "Go unit & integration tests..."
  go test -p=1 ./...
  log_success "Go tests OK"

  # 7. Frontend Linter
  log_step "7/8" "Frontend linting (oxlint)..."
  if command -v oxlint &> /dev/null; then
    oxlint frontend/src
    log_success "Oxlint OK"
  else
    echo "  (skipped - oxlint not installed)"
  fi

  # 8. Frontend Quality & Formatting
  log_step "8/8" "Frontend quality & formatting (biome)..."
  if command -v biome &> /dev/null; then
    biome check frontend/src
    log_success "Biome OK"
  else
    echo "  (skipped - biome not installed)"
  fi

  printf "\n%b========================================%b\n" "$GREEN" "$RESET"
  printf "%b  ALL CHECKS PASSED SUCCESSFULLY!%b\n" "$GREEN" "$RESET"
  printf "%b========================================%b\n" "$GREEN" "$RESET"
fi

#=============================================================================
# EXTRA SCANS (slow, optional)
#=============================================================================
if $DO_EXTRA; then
  echo ""
  echo "=== Extra Scans ==="

  log_step "EXTRA" "Supply chain vulnerability scan (grype)..."
  if command -v grype &> /dev/null; then
    grype dir:. --quiet
    log_success "Grype OK"
  else
    echo "  (skipped - grype not installed)"
  fi
fi

#=============================================================================
# SUMMARY
#=============================================================================
if $DO_BUILD || $DO_AUDIT; then
  echo ""
  printf "%bRelease Info:%b\n" "$YELLOW" "$RESET"
  printf "  Version: %s\n" "$(cat version.txt 2> /dev/null || git describe --tags 2> /dev/null || echo 'dev')"
  printf "  Commit:  %s\n" "$(git rev-parse --short HEAD 2> /dev/null || echo 'unknown')"
  printf "  Go:      %s\n" "$(go version | awk '{print $3}')"
  echo ""
fi

echo "=== Done ==="
