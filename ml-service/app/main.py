import os
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from api.endpoints import router as api_router
from services.predictor import predictor

app = FastAPI(
    title="Razorpay Recovery Intelligence - ML Service",
    version="1.0.0",
    description="Machine Learning Service for Payment Recovery Probability & SHAP Explainability"
)

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.on_event("startup")
def startup_event():
    print("[ML Service] Initializing predictor and ensuring model artifacts...")
    if not predictor.is_loaded:
        predictor.load_models()

@app.get("/health", tags=["Health"])
def health_check():
    return {
        "status": "healthy",
        "service": "rri-ml-service",
        "models_loaded": predictor.is_loaded,
        "version": predictor.metadata.get("version", "v1.0.0")
    }

# Include routers
app.include_router(api_router, prefix="", tags=["Recovery Intelligence"])

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
