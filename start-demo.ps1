# Copyright 2026 José Antonio Garibay Marcelo
# SPDX-License-Identifier: Apache-2.0
#
# Startup script for KyberLink Demo (Windows PowerShell)

Write-Host "Starting Mock Go Backend on :34890..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd examples/mock-backend; go run main.go"

Write-Host "Starting KyberLink Gateway on :45782..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd gateway; go run main.go"

Write-Host "Starting Frontend Example on :5173..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd examples/frontend; npm run dev"

Write-Host "All systems starting. Please wait a few seconds and then open http://localhost:5173" -ForegroundColor Green
