Write-Host "[RRI Seeder] Populating database via synthetic event generation..." -ForegroundColor Cyan
python simulator/event-generator/generator.py 25
Write-Host "[RRI Seeder] Database seeded successfully." -ForegroundColor Green
