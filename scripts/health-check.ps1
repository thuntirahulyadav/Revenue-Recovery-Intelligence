Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Razorpay Recovery Intelligence - Health Check" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# Backend
try {
    $res = Invoke-RestMethod -Uri "http://localhost:8082/health" -Method Get -TimeoutSec 3
    Write-Host "Go Backend (http://localhost:8082/health): OK ($($res.status))" -ForegroundColor Green
} catch {
    Write-Host "Go Backend (http://localhost:8082/health): DOWN" -ForegroundColor Red
}

# ML Service
try {
    $res = Invoke-RestMethod -Uri "http://localhost:8000/health" -Method Get -TimeoutSec 3
    Write-Host "Python ML Service (http://localhost:8000/health): OK ($($res.status))" -ForegroundColor Green
} catch {
    Write-Host "Python ML Service (http://localhost:8000/health): DOWN" -ForegroundColor Red
}

# Frontend
try {
    $res = Invoke-WebRequest -Uri "http://localhost:5173" -Method Get -TimeoutSec 3
    Write-Host "Frontend UI (http://localhost:5173): OK" -ForegroundColor Green
} catch {
    Write-Host "Frontend UI (http://localhost:5173): DOWN" -ForegroundColor Red
}

Write-Host "==========================================" -ForegroundColor Cyan
