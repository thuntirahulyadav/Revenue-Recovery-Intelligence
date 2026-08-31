#!/usr/bin/env bash
set -e

echo "[RRI ML] Starting dataset generation & model training..."
cd ml-service
python training/generate_dataset.py
python training/train.py
python training/evaluate.py
echo "[RRI ML] Training completed successfully!"
