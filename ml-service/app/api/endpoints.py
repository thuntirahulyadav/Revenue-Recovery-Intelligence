from fastapi import APIRouter, HTTPException, status
from schemas.prediction import (
    PredictionRequest,
    PredictionResponse,
    ExplainabilityResponse,
    ModelMetadataResponse
)
from services.predictor import predictor

router = APIRouter()

@router.post("/predict", response_model=PredictionResponse, status_code=status.HTTP_200_OK)
def predict_recovery(request: PredictionRequest):
    """
    Predicts probability of successful revenue recovery given payment context.
    """
    try:
        prob, conf, factors = predictor.predict(request.model_dump())
        version = predictor.metadata.get("version", "v1.0.0")
        return PredictionResponse(
            recovery_probability=prob,
            confidence=conf,
            model_version=version,
            shap_factors=factors
        )
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Inference failed: {str(e)}"
        )

@router.post("/explain", response_model=ExplainabilityResponse, status_code=status.HTTP_200_OK)
def explain_prediction(request: PredictionRequest):
    """
    Returns SHAP-based feature explanations (positive & negative factors).
    """
    try:
        return predictor.explain(request.model_dump())
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Explanation generation failed: {str(e)}"
        )

@router.get("/metrics", response_model=ModelMetadataResponse)
def get_model_metrics():
    """
    Returns model training metadata, comparison between Logistic Regression and XGBoost,
    and feature importance values.
    """
    if not predictor.metadata:
        predictor.load_models()
    return predictor.metadata
