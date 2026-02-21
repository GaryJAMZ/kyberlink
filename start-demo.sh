#!/bin/bash
# Copyright 2026 José Antonio Garibay Marcelo
# SPDX-License-Identifier: Apache-2.0
#
# Startup script for KyberLink Demo (Linux/macOS)

echo "Starting Mock Go Backend on :34890..."
(cd examples/mock-backend && go run main.go) &

echo "Starting KyberLink Gateway on :45782..."
(cd gateway && go run main.go) &

echo "Starting Frontend Example on :5173..."
(cd examples/frontend && npm run dev) &

echo ""
echo "All systems starting. Please wait a few seconds and then open http://localhost:5173"
echo "Press Ctrl+C to stop all services."
wait
