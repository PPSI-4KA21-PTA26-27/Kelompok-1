# Jalankan semua pengecekan backend (versi Windows, tanpa perlu make).
# Pakai:  powershell -ExecutionPolicy Bypass -File .\scripts\check.ps1

Set-Location (Join-Path $PSScriptRoot "..")

function Ensure-Tool($name, $pkg) {
    if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
        Write-Host "Meng-install $name ..." -ForegroundColor Yellow
        go install $pkg
        $env:Path += ";" + (go env GOPATH) + "\bin"
    }
}

function Step($title, [scriptblock]$cmd) {
    Write-Host "`n== $title ==" -ForegroundColor Cyan
    & $cmd
    if ($LASTEXITCODE -ne 0) {
        Write-Host "GAGAL: $title" -ForegroundColor Red
        exit $LASTEXITCODE
    }
}

Ensure-Tool "golangci-lint" "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
Ensure-Tool "govulncheck" "golang.org/x/vuln/cmd/govulncheck@latest"

Step "Build"                         { go build ./... }
Step "Vet"                           { go vet ./... }
Step "Test"                          { go test ./... -cover }
Step "Lint + Security (gosec)"       { golangci-lint run ./... }
Step "Dependency vulnerabilities"    { govulncheck ./... }

Write-Host "`nSemua pengecekan lolos." -ForegroundColor Green
