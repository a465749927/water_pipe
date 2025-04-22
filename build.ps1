# PowerShell build script for Water Pipe proxy

# Variables
$BinaryName = "water_pipe.exe"

# Build function
function Build-WaterPipe {
    Write-Host "Building Water Pipe proxy..."
    go build -ldflags="-s -w" -o $BinaryName .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "Build successful!" -ForegroundColor Green
}

# Test function
function Test-WaterPipe {
    Write-Host "Running tests..."
    go test -v ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "Tests passed!" -ForegroundColor Green
}

# Clean function
function Clean-WaterPipe {
    Write-Host "Cleaning build artifacts..."
    if (Test-Path $BinaryName) {
        Remove-Item $BinaryName
    }
    Write-Host "Clean completed!" -ForegroundColor Green
}

# Main script
if ($args.Count -eq 0) {
    Build-WaterPipe
} else {
    switch ($args[0]) {
        "build" { Build-WaterPipe }
        "test" { Test-WaterPipe }
        "clean" { Clean-WaterPipe }
        default {
            Write-Host "Unknown command: $($args[0])" -ForegroundColor Red
            Write-Host "Available commands: build, test, clean" -ForegroundColor Yellow
        }
    }
}
