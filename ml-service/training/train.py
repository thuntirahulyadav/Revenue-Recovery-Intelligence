import os
import json
import joblib
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import (
    accuracy_score,
    precision_score,
    recall_score,
    f1_score,
    roc_auc_score,
    average_precision_score,
    confusion_matrix,
)
from xgboost import XGBClassifier
import shap

from feature_engineering import engineer_features, get_feature_columns
from generate_dataset import generate_synthetic_dataset

def train_recovery_models():
    print("[Training] 1. Generating or loading synthetic dataset...")
    base_dir = os.path.dirname(os.path.abspath(__file__))
    artifacts_dir = os.path.join(base_dir, "..", "artifacts")
    data_dir = os.path.join(base_dir, "..", "data", "synthetic")
    os.makedirs(artifacts_dir, exist_ok=True)
    os.makedirs(data_dir, exist_ok=True)

    csv_path = os.path.join(data_dir, "payments_recovery_synthetic.csv")
    if os.path.exists(csv_path):
        df_raw = pd.read_csv(csv_path)
    else:
        df_raw = generate_synthetic_dataset(num_records=12500, seed=42)
        df_raw.to_csv(csv_path, index=False)

    print(f"[Training] Loaded {len(df_raw)} records. Transforming features...")
    df_feat = engineer_features(df_raw)
    feature_cols = get_feature_columns()

    X = df_feat[feature_cols].values
    y = df_feat["recovery_success"].values

    # Train / Test split (80% train, 20% test, stratified)
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.20, random_state=42, stratify=y
    )

    # Standardize numerical features for Logistic Regression
    scaler = StandardScaler()
    X_train_scaled = scaler.fit_transform(X_train)
    X_test_scaled = scaler.transform(X_test)

    # 1. Baseline Model: Logistic Regression
    print("[Training] 2. Training Baseline Model: Logistic Regression...")
    lr_model = LogisticRegression(max_iter=1000, random_state=42)
    lr_model.fit(X_train_scaled, y_train)

    lr_pred = lr_model.predict(X_test_scaled)
    lr_prob = lr_model.predict_proba(X_test_scaled)[:, 1]

    lr_metrics = {
        "model_name": "Logistic Regression (Baseline)",
        "accuracy": round(float(accuracy_score(y_test, lr_pred)), 4),
        "precision": round(float(precision_score(y_test, lr_pred)), 4),
        "recall": round(float(recall_score(y_test, lr_pred)), 4),
        "f1_score": round(float(f1_score(y_test, lr_pred)), 4),
        "roc_auc": round(float(roc_auc_score(y_test, lr_prob)), 4),
        "pr_auc": round(float(average_precision_score(y_test, lr_prob)), 4),
        "confusion_matrix": confusion_matrix(y_test, lr_pred).tolist(),
    }

    # 2. Primary Model: XGBoost Classifier
    print("[Training] 3. Training Primary Model: XGBoost Classifier...")
    xgb_model = XGBClassifier(
        n_estimators=180,
        max_depth=5,
        learning_rate=0.08,
        subsample=0.85,
        colsample_bytree=0.85,
        random_state=42,
        eval_metric="logloss",
    )
    xgb_model.fit(X_train, y_train)

    xgb_pred = xgb_model.predict(X_test)
    xgb_prob = xgb_model.predict_proba(X_test)[:, 1]

    xgb_metrics = {
        "model_name": "XGBoost Classifier (Primary)",
        "accuracy": round(float(accuracy_score(y_test, xgb_pred)), 4),
        "precision": round(float(precision_score(y_test, xgb_pred)), 4),
        "recall": round(float(recall_score(y_test, xgb_pred)), 4),
        "f1_score": round(float(f1_score(y_test, xgb_pred)), 4),
        "roc_auc": round(float(roc_auc_score(y_test, xgb_prob)), 4),
        "pr_auc": round(float(average_precision_score(y_test, xgb_prob)), 4),
        "confusion_matrix": confusion_matrix(y_test, xgb_pred).tolist(),
    }

    # Feature importances from XGBoost
    importances = xgb_model.feature_importances_
    feat_importance_dict = {
        col: round(float(imp), 4) for col, imp in zip(feature_cols, importances)
    }
    sorted_importances = dict(sorted(feat_importance_dict.items(), key=lambda item: item[1], reverse=True))

    # 3. SHAP TreeExplainer
    print("[Training] 4. Fitting SHAP TreeExplainer on background sample...")
    background_sample = X_train[:300]
    shap_explainer = shap.TreeExplainer(xgb_model, data=background_sample)

    # Save artifacts
    print("[Training] 5. Saving model artifacts...")
    joblib.dump(xgb_model, os.path.join(artifacts_dir, "xgb_model.joblib"))
    joblib.dump(lr_model, os.path.join(artifacts_dir, "lr_model.joblib"))
    joblib.dump(scaler, os.path.join(artifacts_dir, "scaler.joblib"))
    joblib.dump(shap_explainer, os.path.join(artifacts_dir, "shap_explainer.joblib"))

    metadata = {
        "version": "v1.0.0",
        "training_date": pd.Timestamp.now().isoformat(),
        "num_records": len(df_raw),
        "feature_columns": feature_cols,
        "baseline_model": lr_metrics,
        "primary_model": xgb_metrics,
        "feature_importance": sorted_importances,
    }

    with open(os.path.join(artifacts_dir, "model_metadata.json"), "w") as f:
        json.dump(metadata, f, indent=2)

    print("=" * 60)
    print("MODEL COMPARISON RESULTS:")
    print("=" * 60)
    print(f"Metric        | Baseline (Logistic Reg) | Primary (XGBoost)")
    print("-" * 60)
    for m in ["accuracy", "precision", "recall", "f1_score", "roc_auc", "pr_auc"]:
        print(f"{m:<13} | {lr_metrics[m]:<23} | {xgb_metrics[m]:<18}")
    print("=" * 60)
    print(f"[Training] Saved model artifacts to: {artifacts_dir}")
    return metadata

if __name__ == "__main__":
    train_recovery_models()
