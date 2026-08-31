#!/usr/bin/env bash
set -e

echo "[RRI Seeder] Populating database via synthetic event generation..."
python simulator/event-generator/generator.py 25
echo "[RRI Seeder] Database seeded successfully."
