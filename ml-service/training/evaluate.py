import os
import json
import joblib
import pandas as pd
from feature_engineering import engineer_features, get_feature_columns

def evaluate_saved_models():
    base_dir = os.path.dirname(os.path.abspath(__file__))
    artifacts_dir = os.path.join(base_dir, "..", "artifacts")
    meta_path = os.path.join(artifacts_dir, "model_metadata.json")

    if not os.path.exists(meta_path):
        print(f"[Evaluate] No metadata found at {meta_path}. Please run train.py first.")
        return

    with open(meta_path, "r") as f:
        meta = json.load(f)

    print("=" * 65)
    print(f"RAZORPAY RECOVERY INTELLIGENCE - MODEL EVALUATION REPORT ({meta['version']})")
    print("=" * 65)
    print(f"Training Date: {meta['training_date']}")
    print(f"Total Records: {meta['num_records']}")
    print("-" * 65)
    print(f"{'Metric':<16} | {'Baseline (Logistic Reg)':<23} | {'Primary (XGBoost)':<18}")
    print("-" * 65)
    for k in ["accuracy", "precision", "recall", "f1_score", "roc_auc", "pr_auc"]:
        print(f"{k:<16} | {meta['baseline_model'][k]:<23} | {meta['primary_model'][k]:<18}")
    print("=" * 65)
    print("\nTOP 8 FEATURE IMPORTANCES (XGBoost):")
    for feat, score in list(meta['feature_importance'].items())[:8]:
        bar = "#" * int(score * 40)
        print(f"  {feat:<25} : {score:.4f} {bar}")

if __name__ == "__main__":
    evaluate_saved_models()
