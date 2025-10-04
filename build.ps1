# build.ps1 - Build script for Windows with version injection

param(
    [string]$OutputName = "aws-ssm.exe"
)

# Get version information
try {
    $Version = git describe --tags --exact-match 2>$null
    if (-not $Version) {
        $ShortCommit = git rev-parse --short HEAD
        $Version = "dev-$ShortCommit"
    }
} catch {
    $Version = "dev-unknown"
}

try {
    $GitCommit = git rev-parse HEAD
} catch {
    $GitCommit = "unknown"
}

$BuildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$Maintainer = "Nicolas_HYPOLITE"

# Build flags - using underscore instead of space to avoid PowerShell issues
$LdFlags = "-w -s -X main.Version=$Version -X main.GitCommit=$GitCommit -X main.BuildTime=$BuildTime -X main.Maintainer=$Maintainer"

Write-Host "Building aws-ssm with version information:"
Write-Host "  Version:    $Version"
Write-Host "  Commit:     $GitCommit"
Write-Host "  Time:       $BuildTime"
Write-Host "  Maintainer: Nicolas HYPOLITE"
Write-Host ""

# Build the binary
$env:GOOS = ""
$env:GOARCH = ""
go build -ldflags $LdFlags -o $OutputName main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful: $OutputName"
    Write-Host ""
    Write-Host "Test the version:"
    Write-Host "  ./$OutputName --version"
} else {
    Write-Host "❌ Build failed"
    exit 1
}