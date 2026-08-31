import pytest
from fastapi.testclient import TestClient
import sys
import os
import numpy as np

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
APP_DIR = os.path.join(PROJECT_ROOT, "app")

# Allow this test module to run from either the repository root or ml-service.
for path in (PROJECT_ROOT, APP_DIR):
    if path not in sys.path:
        sys.path.insert(0, path)
from main import app
from services.predictor import predictor
from training.feature_engineering import get_feature_columns

client = TestClient(app)

def test_health_check():
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "healthy"
    assert data["service"] == "rri-ml-service"

def test_predict_endpoint():
    payload = {
        "transaction_amount": 25000.0,
        "payment_method": "card",
        "failure_reason": "BANK_TIMEOUT",
        "attempt_count": 1,
        "customer_success_rate": 0.92,
        "customer_failure_rate": 0.08,
        "customer_value": 50000.0,
        "hour_of_day": 14,
        "day_of_week": 2
    }
    response = client.post("/predict", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert "recovery_probability" in data
    assert 0.0 <= data["recovery_probability"] <= 1.0
    assert "confidence" in data
    assert 0.0 <= data["confidence"] <= 1.0
    assert "shap_factors" in data
    assert len(data["shap_factors"]) > 0

def test_explain_endpoint():
    payload = {
        "transaction_amount": 75000.0,
        "payment_method": "upi",
        "failure_reason": "NETWORK_ERROR",
        "attempt_count": 1,
        "customer_success_rate": 0.88,
        "customer_failure_rate": 0.12,
        "customer_value": 80000.0,
        "hour_of_day": 11,
        "day_of_week": 3
    }
    response = client.post("/explain", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert "positive_factors" in data
    assert "negative_factors" in data
    assert "summary" in data

def test_predict_endpoint_handles_single_row_shap_matrix(monkeypatch):
    payload = {
        "transaction_amount": 25000.0,
        "payment_method": "card",
        "failure_reason": "BANK_TIMEOUT",
        "attempt_count": 1,
        "customer_success_rate": 0.92,
        "customer_failure_rate": 0.08,
        "customer_value": 50000.0,
        "hour_of_day": 14,
        "day_of_week": 2
    }

    feature_cols = get_feature_columns()
    fake_shap = np.linspace(-0.2, 0.2, len(feature_cols), dtype=float).reshape(len(feature_cols), 1)

    monkeypatch.setattr(predictor.shap_explainer, "shap_values", lambda _: fake_shap)

    prob, conf, factors = predictor.predict(payload)

    assert 0.0 <= prob <= 1.0
    assert 0.0 <= conf <= 1.0
    assert len(factors) >= 3
    feature_names = {factor.feature for factor in factors}
    assert "transaction_amount" in feature_names
    assert "attempt_count" in feature_names


def test_normalize_shap_values_handles_legacy_binary_class_list():
    feature_count = len(get_feature_columns())
    negative_class = np.zeros((1, feature_count), dtype=float)
    positive_class = np.linspace(-0.2, 0.2, feature_count, dtype=float).reshape(1, feature_count)

    normalized = predictor._normalize_shap_values(
        [negative_class, positive_class], feature_count
    )

    np.testing.assert_allclose(normalized, positive_class[0])


def test_normalize_shap_values_handles_current_three_dimensional_output():
    feature_count = len(get_feature_columns())
    values = np.zeros((1, feature_count, 2), dtype=float)
    values[0, :, 1] = np.linspace(-0.2, 0.2, feature_count, dtype=float)

    normalized = predictor._normalize_shap_values(values, feature_count)

    np.testing.assert_allclose(normalized, values[0, :, 1])


def test_model_metrics_endpoint():
    response = client.get("/metrics")
    assert response.status_code == 200
    data = response.json()
    assert "primary_model" in data
    assert "baseline_model" in data
    assert "feature_importance" in data
