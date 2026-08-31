Write-Host "[RRI ML] Starting dataset generation & model training..." -ForegroundColor Cyan
Set-Location ml-service
python training/generate_dataset.py
python training/train.py
python training/evaluate.py
Set-Location ..
Write-Host "[RRI ML] Training completed successfully!" -ForegroundColor Green
