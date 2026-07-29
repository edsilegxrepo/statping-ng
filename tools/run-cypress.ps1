# Cypress E2E Test Runner
# Usage: .\run-cypress.ps1 [-Open] [-Spec <pattern>]

param(
    [switch]$Open,           # Use -Open flag to open Cypress UI instead of headless
    [string]$Spec = ""       # Optional: run specific spec file(s), e.g. "groups_spec.js"
)

$ErrorActionPreference = "Continue"
$ProjectRoot = $PSScriptRoot

# Fix UTF-8 encoding for Cypress box-drawing characters
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$env:LANG = "en_US.UTF-8"
$LogFile = "$ProjectRoot\cypress-test.log"
$Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Start logging
"[$Timestamp] Cypress E2E Test Run Started" | Out-File $LogFile

Write-Host "Killing any existing processes..." -ForegroundColor Cyan
"[$Timestamp] Killing existing processes (statping, node, Cypress)" | Add-Content $LogFile
try { taskkill /F /IM statping.exe 2>&1 | Out-Null } catch {}
try { taskkill /F /IM node.exe 2>&1 | Out-Null } catch {}
try { taskkill /F /IM Cypress.exe 2>&1 | Out-Null } catch {}
Start-Sleep -Seconds 2

Write-Host "Cleaning up previous test state..." -ForegroundColor Cyan
Remove-Item -Force -ErrorAction SilentlyContinue "$ProjectRoot\statping.db"
Remove-Item -Force -ErrorAction SilentlyContinue "$ProjectRoot\config.yml"
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue "$ProjectRoot\logs"
Remove-Item -Force -ErrorAction SilentlyContinue $LogFile

"[$Timestamp] Cleaning up: statping.db, config.yml, logs, cypress-test.log" | Out-File $LogFile

# Check if Cypress is installed and working
Write-Host "Checking Cypress installation..." -ForegroundColor Cyan
Push-Location "$ProjectRoot\frontend"

# First check if cypress package exists in node_modules
if (-not (Test-Path "$ProjectRoot\frontend\node_modules\cypress")) {
    Write-Host "Cypress package not found in node_modules. Run 'npx yarn add --dev cypress@12.17.4' first." -ForegroundColor Red
    "[$Timestamp] ERROR: Cypress not in node_modules" | Add-Content $LogFile
    Pop-Location
    exit 1
}

# Verify the binary cache is valid
npx yarn cypress verify 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Cypress binary missing or corrupted. Installing..." -ForegroundColor Yellow
    "[$Timestamp] Cypress verify failed, installing binary..." | Add-Content $LogFile

    # Install Cypress binary (downloads to cache)
    npx yarn cypress install 2>&1 | ForEach-Object { Write-Host $_ }

    # Verify installation
    npx yarn cypress verify 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Cypress binary installation failed!" -ForegroundColor Red
        "[$Timestamp] ERROR: Cypress binary installation failed" | Add-Content $LogFile
        Pop-Location
        exit 1
    }
    Write-Host "Cypress binary installed successfully." -ForegroundColor Green
    "[$Timestamp] Cypress binary installed" | Add-Content $LogFile
} else {
    Write-Host "Cypress verified." -ForegroundColor Green
    "[$Timestamp] Cypress verified" | Add-Content $LogFile
}

Pop-Location

# Check if frontend needs rebuilding
Write-Host "Checking if frontend rebuild is needed..." -ForegroundColor Cyan
$distDir = "$ProjectRoot\frontend\dist"
$srcDir = "$ProjectRoot\frontend\src"

$needsBuild = $false
if (-not (Test-Path $distDir)) {
    Write-Host "No dist folder found - build required" -ForegroundColor Yellow
    $needsBuild = $true
} else {
    $distMarker = Get-Item "$distDir\base.gohtml" -ErrorAction SilentlyContinue
    if (-not $distMarker) {
        Write-Host "No dist/base.gohtml found - build required" -ForegroundColor Yellow
        $needsBuild = $true
    } else {
        # Check if any src file is newer than dist
        $newestSrc = Get-ChildItem -Path $srcDir -Recurse -File | Sort-Object LastWriteTime -Descending | Select-Object -First 1
        if ($newestSrc -and ($newestSrc.LastWriteTime -gt $distMarker.LastWriteTime)) {
            Write-Host "Source files newer than dist - rebuild required" -ForegroundColor Yellow
            Write-Host "  Newest: $($newestSrc.FullName)" -ForegroundColor Gray
            $needsBuild = $true
        }
    }
}

if ($needsBuild) {
    Write-Host "Building frontend..." -ForegroundColor Yellow
    "[$Timestamp] Rebuilding frontend" | Add-Content $LogFile
    Push-Location "$ProjectRoot\frontend"
    $env:NODE_OPTIONS = "--openssl-legacy-provider --max-old-space-size=4096"
    npx yarn build 2>&1 | ForEach-Object { Write-Host $_ }
    $buildExit = $LASTEXITCODE
    $env:NODE_OPTIONS = $null  # Clear to avoid affecting Cypress
    if ($buildExit -ne 0) {
        Write-Host "ERROR: Frontend build failed!" -ForegroundColor Red
        "[$Timestamp] ERROR: Frontend build failed" | Add-Content $LogFile
        Pop-Location
        exit 1
    }
    Pop-Location
    Write-Host "Frontend build complete." -ForegroundColor Green
    "[$Timestamp] Frontend build complete" | Add-Content $LogFile
} else {
    Write-Host "Frontend is up to date." -ForegroundColor Green
    "[$Timestamp] Frontend up to date - skipping build" | Add-Content $LogFile
}

# Start statping in background
Write-Host "Starting statping backend..." -ForegroundColor Cyan
"[$Timestamp] Starting statping backend in background" | Add-Content $LogFile
$statpingProcess = Start-Process -FilePath "$ProjectRoot\statping.exe" -ArgumentList "--port 8080" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 3

# Wait for API to be ready
Write-Host "Waiting for API to be ready..." -ForegroundColor Cyan
$maxWait = 30
$waited = 0
while ($waited -lt $maxWait) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/api" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            Write-Host "API is ready." -ForegroundColor Green
            "[$Timestamp] API ready after ${waited}s" | Add-Content $LogFile
            break
        }
    } catch {
        Start-Sleep -Seconds 1
        $waited++
    }
}
if ($waited -ge $maxWait) {
    Write-Host "ERROR: API did not start in time!" -ForegroundColor Red
    "[$Timestamp] ERROR: API timeout" | Add-Content $LogFile
    Stop-Process -Id $statpingProcess.Id -Force -ErrorAction SilentlyContinue
    exit 1
}

Write-Host "Starting Cypress E2E tests..." -ForegroundColor Green
"[$Timestamp] Starting tests (Open=$Open, Spec=$Spec)" | Add-Content $LogFile

Push-Location "$ProjectRoot\frontend"
try {
    if ($Open) {
        Write-Host "Opening Cypress UI..." -ForegroundColor Yellow
        npx cypress open 2>&1 | ForEach-Object {
            Write-Host $_
            $_ | Add-Content $LogFile -Encoding UTF8
        }
    } else {
        $cypressArgs = @("cypress", "run", "--browser", "D:\inet\www\chromium\bin\chrome.exe")
        if ($Spec) {
            # Handle comma-separated specs by prefixing each with the integration path
            $specPaths = ($Spec -split ',') | ForEach-Object { "cypress/integration/$($_.Trim())" }
            $cypressArgs += "--spec"
            $cypressArgs += ($specPaths -join ',')
        }
        Write-Host "Running: npx $($cypressArgs -join ' ')" -ForegroundColor Yellow
        & npx @cypressArgs 2>&1 | ForEach-Object {
            Write-Host $_
            $_ | Add-Content $LogFile -Encoding UTF8
        }
    }
    $ExitCode = $LASTEXITCODE
} finally {
    Pop-Location
    # Stop statping
    Write-Host "Stopping statping..." -ForegroundColor Cyan
    Stop-Process -Id $statpingProcess.Id -Force -ErrorAction SilentlyContinue
}

$Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
"[$Timestamp] Test run completed with exit code: $ExitCode" | Add-Content $LogFile
Write-Host "`nLog saved to: $LogFile" -ForegroundColor Cyan
exit $ExitCode
