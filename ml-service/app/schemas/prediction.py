from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any

class PredictionRequest(BaseModel):
    transaction_amount: float = Field(..., description="Transaction amount in INR currency units", ge=0)
    payment_method: str = Field(..., description="Payment method: card, upi, netbanking, wallet, emi")
    failure_reason: str = Field(..., description="Failure category e.g., BANK_TIMEOUT, NETWORK_ERROR, INSUFFICIENT_FUNDS")
    attempt_count: int = Field(default=1, ge=1, description="Number of attempts so far")
    customer_success_rate: float = Field(default=0.85, ge=0.0, le=1.0, description="Customer lifetime payment success rate")
    customer_failure_rate: float = Field(default=0.15, ge=0.0, le=1.0, description="Customer lifetime payment failure rate")
    customer_value: float = Field(default=10000.0, ge=0.0, description="Customer lifetime spend value")
    hour_of_day: int = Field(default=14, ge=0, le=23, description="Hour of day (0-23)")
    day_of_week: int = Field(default=2, ge=0, le=6, description="Day of week (0=Mon, 6=Sun)")

class SHAPFactor(BaseModel):
    feature: str
    impact: float
    direction: str # 'positive' | 'negative'
    description: str

class PredictionResponse(BaseModel):
    recovery_probability: float
    confidence: float
    model_version: str
    shap_factors: List[SHAPFactor] = []

class ExplainabilityResponse(BaseModel):
    model_version: str
    base_value: float
    positive_factors: List[SHAPFactor]
    negative_factors: List[SHAPFactor]
    summary: str

class ModelMetadataResponse(BaseModel):
    version: str
    training_date: str
    num_records: int
    primary_model: Dict[str, Any]
    baseline_model: Dict[str, Any]
    feature_importance: Dict[str, float]
