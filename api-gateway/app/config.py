import os


class Settings:
    APP_NAME: str = os.getenv("APP_NAME", "Pulse API Gateway")
    APP_VERSION: str = os.getenv("APP_VERSION", "1.0.0")
    HOST: str = os.getenv("API_HOST", "0.0.0.0")
    PORT: int = int(os.getenv("API_PORT", "8000"))
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO")
    HEALTH_CHECKER_URL: str = os.getenv("HEALTH_CHECKER_URL", "http://health-checker:8001")
    ALERT_SERVICE_URL: str = os.getenv("ALERT_SERVICE_URL", "http://alert-service:8002")


settings = Settings()
