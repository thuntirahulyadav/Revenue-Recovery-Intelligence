import numpy as np
import pandas as pd
from typing import Tuple, Dict, Any, List

PAYMENT_METHODS = ["card", "upi", "netbanking", "wallet", "emi"]
FAILURE_REASONS = [
    "BANK_TIMEOUT",
    "NETWORK_ERROR",
    "INSUFFICIENT_FUNDS",
    "CARD_EXPIRED",
    "PAYMENT_METHOD_FAILURE",
    "CUSTOMER_ABANDONMENT",
    "TECHNICAL_ERROR",
]

NUMERICAL_FEATURES = [
    "transaction_amount",
    "attempt_count",
    "customer_success_rate",
    "customer_failure_rate",
    "customer_value",
    "hour_of_day",
    "day_of_week",
    "customer_risk_ratio",
    "value_per_attempt",
    "is_peak_hour",
    "is_weekend",
]

def engineer_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Applies deterministic feature transformations to the dataset.
    Works for both batch training and single-record real-time inference.
    """
    df = df.copy()

    # Derived interactions
    df["customer_risk_ratio"] = df["customer_failure_rate"] / (df["customer_success_rate"] + 0.001)
    df["value_per_attempt"] = df["transaction_amount"] / (df["attempt_count"].clip(lower=1))
    df["is_peak_hour"] = df["hour_of_day"].apply(lambda h: 1 if 9 <= h <= 21 else 0)
    df["is_weekend"] = df["day_of_week"].apply(lambda d: 1 if d in [5, 6] else 0)

    # One-Hot Encode Categorical variables with fixed categories
    for method in PAYMENT_METHODS:
        df[f"pm_{method}"] = (df["payment_method"] == method).astype(int)

    for reason in FAILURE_REASONS:
        df[f"reason_{reason}"] = (df["failure_reason"] == reason).astype(int)

    return df

def get_feature_columns() -> List[str]:
    """Returns the ordered list of final feature column names."""
    cols = list(NUMERICAL_FEATURES)
    for method in PAYMENT_METHODS:
        cols.append(f"pm_{method}")
    for reason in FAILURE_REASONS:
        cols.append(f"reason_{reason}")
    return cols

def prepare_inference_features(payload: Dict[str, Any]) -> pd.DataFrame:
    """Transforms a single payload dict into a model-ready feature vector."""
    df_raw = pd.DataFrame([payload])
    df_feat = engineer_features(df_raw)
    feature_cols = get_feature_columns()
    
    # Ensure all columns exist
    for col in feature_cols:
        if col not in df_feat.columns:
            df_feat[col] = 0

    return df_feat[feature_cols]
