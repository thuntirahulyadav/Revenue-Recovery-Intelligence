import os
import sys
import json
import joblib
import numpy as np
import pandas as pd
from typing import Dict, Any, List, Tuple

# Ensure training directory is on path for feature_engineering
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
TRAINING_DIR = os.path.abspath(os.path.join(CURRENT_DIR, "..", "..", "training"))
ARTIFACTS_DIR = os.path.abspath(os.path.join(CURRENT_DIR, "..", "..", "artifacts"))
if TRAINING_DIR not in sys.path:
    sys.path.insert(0, TRAINING_DIR)

from feature_engineering import engineer_features, get_feature_columns, prepare_inference_features
from schemas.prediction import SHAPFactor, PredictionResponse, ExplainabilityResponse

FEATURE_DESCRIPTIONS = {
    "customer_success_rate": "Historical customer success rate",
    "customer_failure_rate": "Historical customer failure rate",
    "customer_value": "Customer lifetime transaction value",
    "attempt_count": "Number of retry attempts",
    "transaction_amount": "Transaction value",
    "hour_of_day": "Transaction hour",
    "day_of_week": "Day of week",
    "customer_risk_ratio": "Customer risk ratio",
    "value_per_attempt": "Amount value per attempt",
    "is_peak_hour": "Payment during peak banking hours",
    "is_weekend": "Weekend processing window",
    "reason_BANK_TIMEOUT": "Transient core bank timeout",
    "reason_NETWORK_ERROR": "Gateway network connectivity glitch",
    "reason_INSUFFICIENT_FUNDS": "Account balance deficit",
    "reason_CARD_EXPIRED": "Card validity period expired",
    "reason_PAYMENT_METHOD_FAILURE": "Payment instrument failure",
    "reason_CUSTOMER_ABANDONMENT": "Customer drop-off during authentication",
    "reason_TECHNICAL_ERROR": "Technical / upstream gateway error",
    "pm_card": "Card payment channel",
    "pm_upi": "UPI instant payment channel",
    "pm_netbanking": "NetBanking channel",
    "pm_wallet": "Digital wallet channel",
    "pm_emi": "EMI payment channel",
}

class RecoveryPredictor:
    def __init__(self):
        self.xgb_model = None
        self.lr_model = None
        self.scaler = None
        self.shap_explainer = None
        self.metadata = {}
        self.is_loaded = False
        self.load_models()

    def load_models(self):
        try:
            xgb_path = os.path.join(ARTIFACTS_DIR, "xgb_model.joblib")
            lr_path = os.path.join(ARTIFACTS_DIR, "lr_model.joblib")
            scaler_path = os.path.join(ARTIFACTS_DIR, "scaler.joblib")
            shap_path = os.path.join(ARTIFACTS_DIR, "shap_explainer.joblib")
            meta_path = os.path.join(ARTIFACTS_DIR, "model_metadata.json")

            if not os.path.exists(xgb_path):
                print("[Predictor] Artifacts not found. Triggering training pipeline...")
                from train import train_recovery_models
                train_recovery_models()

            self.xgb_model = joblib.load(xgb_path)
            self.lr_model = joblib.load(lr_path)
            self.scaler = joblib.load(scaler_path)
            self.shap_explainer = joblib.load(shap_path)

            if os.path.exists(meta_path):
                with open(meta_path, "r") as f:
                    self.metadata = json.load(f)

            self.is_loaded = True
            print("[Predictor] Models and SHAP explainer loaded successfully.")
        except Exception as e:
            print(f"[Predictor] Error loading models: {e}")
            self.is_loaded = False

    def _normalize_shap_values(self, shap_values: Any, expected_features: int) -> np.ndarray:
        # SHAP < 0.45 returns a list containing one matrix per class.  Select
        # the positive (recovery) class before coercing it to an array.  Calling
        # np.asarray on that list first loses the class/sample/feature meaning
        # and previously resulted in feature values from the wrong axis.
        if isinstance(shap_values, (list, tuple)):
            if not shap_values:
                return np.array([], dtype=float)
            shap_values = shap_values[1] if len(shap_values) > 1 else shap_values[0]

        arr = np.asarray(shap_values, dtype=float)
        if arr.size == 0:
            return np.array([], dtype=float)

        if arr.ndim == 3:
            # SHAP >= 0.45: (samples, features, outputs)
            if arr.shape[1] == expected_features:
                output_index = 1 if arr.shape[2] > 1 else 0
                arr = arr[0, :, output_index]
            # Legacy array form: (outputs, samples, features)
            elif arr.shape[2] == expected_features:
                output_index = 1 if arr.shape[0] > 1 else 0
                arr = arr[output_index, 0, :]
            else:
                raise ValueError(
                    f"Unsupported SHAP value shape {arr.shape}; expected "
                    f"{expected_features} feature contributions."
                )
        elif arr.ndim == 2:
            if arr.shape[0] == 1 and arr.shape[1] == expected_features:
                arr = arr[0]
            elif arr.shape[1] == 1 and arr.shape[0] == expected_features:
                arr = arr[:, 0]
            elif arr.shape[0] == expected_features and arr.shape[1] > 1:
                arr = arr[:, 1]
            elif arr.shape[1] == expected_features:
                arr = arr[0]

        arr = np.asarray(arr, dtype=float).ravel()
        if arr.size > expected_features:
            arr = arr[:expected_features]
        elif arr.size < expected_features:
            pad = expected_features - arr.size
            if pad > 0:
                arr = np.pad(arr, (0, pad), mode='constant', constant_values=0.0)
        return arr

    def _calculate_shap_values(self, features: np.ndarray) -> Any:
        """Support both the legacy and current SHAP explainer interfaces."""
        if self.shap_explainer is None:
            raise RuntimeError("SHAP explainer is not available")

        shap_values_method = getattr(self.shap_explainer, "shap_values", None)
        if callable(shap_values_method):
            return shap_values_method(features)

        explanation = self.shap_explainer(features)
        return getattr(explanation, "values", explanation)

    def predict(self, req_data: Dict[str, Any]) -> Tuple[float, float, List[SHAPFactor]]:
        if not self.is_loaded:
            self.load_models()

        df_feat = prepare_inference_features(req_data)
        X_vec = df_feat.values

        # XGBoost probability
        prob = float(self.xgb_model.predict_proba(X_vec)[0, 1])

        # Confidence score based on distance from decision boundary (0.5)
        # Closer to 0 or 1 means higher model certainty
        confidence = float(min(1.0, 0.5 + abs(prob - 0.5) * 1.0))

        # SHAP calculation
        shap_factors = []
        try:
            shap_values = self._calculate_shap_values(X_vec)
            feature_cols = get_feature_columns()
            s_vals = self._normalize_shap_values(shap_values, len(feature_cols))

            for col_name, val in zip(feature_cols, s_vals):
                val_float = float(val)
                if abs(val_float) > 0.015:
                    direction = "positive" if val_float > 0 else "negative"
                    desc = FEATURE_DESCRIPTIONS.get(col_name, col_name)
                    shap_factors.append(SHAPFactor(
                        feature=col_name,
                        impact=round(val_float, 4),
                        direction=direction,
                        description=desc
                    ))

            # Sort factors by magnitude
            shap_factors.sort(key=lambda x: abs(x.impact), reverse=True)
        except Exception as e:
            print(f"[Predictor] SHAP explanation calculation note: {e}")

        return round(prob, 4), round(confidence, 4), shap_factors[:6]

    def explain(self, req_data: Dict[str, Any]) -> ExplainabilityResponse:
        prob, conf, factors = self.predict(req_data)
        pos = [f for f in factors if f.direction == "positive"]
        neg = [f for f in factors if f.direction == "negative"]

        pos_text = ", ".join([f.description for f in pos[:2]]) if pos else "standard metrics"
        neg_text = ", ".join([f.description for f in neg[:2]]) if neg else "minimal risk factors"

        if prob >= 0.70:
            summary = f"High recovery probability ({prob:.0%}) driven by {pos_text}."
        elif prob >= 0.40:
            summary = f"Moderate recovery potential ({prob:.0%}). Positive influence from {pos_text}, balanced against {neg_text}."
        else:
            summary = f"Low recovery likelihood ({prob:.0%}) constrained primarily by {neg_text}."

        return ExplainabilityResponse(
            model_version=self.metadata.get("version", "v1.0.0"),
            base_value=0.50,
            positive_factors=pos,
            negative_factors=neg,
            summary=summary
        )

predictor = RecoveryPredictor()
