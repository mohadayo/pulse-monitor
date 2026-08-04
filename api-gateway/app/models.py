from pydantic import BaseModel, Field, field_validator
from typing import Optional
from datetime import datetime
from enum import Enum
from urllib.parse import urlparse


class ServiceStatus(str, Enum):
    HEALTHY = "healthy"
    UNHEALTHY = "unhealthy"
    UNKNOWN = "unknown"


class ServiceCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    url: str = Field(..., min_length=1)
    interval_seconds: int = Field(default=30, ge=5, le=3600)

    @field_validator("url")
    @classmethod
    def validate_url(cls, value: str) -> str:
        """Ensure the URL uses http/https and includes a host name."""
        try:
            parsed = urlparse(value)
        except (ValueError, TypeError) as exc:
            raise ValueError("url must be a valid URL") from exc
        if parsed.scheme not in ("http", "https"):
            raise ValueError("url must start with http:// or https://")
        if not parsed.netloc or not parsed.hostname:
            raise ValueError("url must include a host name")
        return value


class Service(BaseModel):
    id: str
    name: str
    url: str
    interval_seconds: int
    status: ServiceStatus = ServiceStatus.UNKNOWN
    last_checked: Optional[datetime] = None
    created_at: datetime


class HealthResponse(BaseModel):
    status: str
    service: str
    version: str
    timestamp: datetime


class AlertConfig(BaseModel):
    service_id: str
    webhook_url: Optional[str] = None
    email: Optional[str] = None


class ErrorResponse(BaseModel):
    error: str
    detail: str
