#!/usr/bin/env bash
set -e

echo "=========================================="
echo "Razorpay Recovery Intelligence - Health Check"
echo "=========================================="

echo -n "Checking Go Backend (http://localhost:8082/health)... "
curl -sf http://localhost:8082/health > /dev/null && echo "✅ HEALTHY" || echo "❌ DOWN"

echo -n "Checking Python ML Service (http://localhost:8000/health)... "
curl -sf http://localhost:8000/health > /dev/null && echo "✅ HEALTHY" || echo "❌ DOWN"

echo -n "Checking Frontend UI (http://localhost:5173)... "
curl -sf http://localhost:5173 > /dev/null && echo "✅ HEALTHY" || echo "❌ DOWN"

echo "=========================================="
