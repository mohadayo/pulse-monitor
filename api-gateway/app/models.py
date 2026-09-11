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

    @field_validator("name")
    @classmethod
    def validate_name(cls, value: str) -> str:
        """Reject whitespace-only names and trim surrounding whitespace.

        `Field(min_length=1)` only checks raw string length, so a value like
        ``"   "`` slips through. A name that renders as visually blank is
        unusable in dashboards, logs and alert messages, so we require at
        least one non-whitespace character and normalize the stored value by
        stripping leading/trailing whitespace.
        """
        stripped = value.strip()
        if not stripped:
            raise ValueError("name must contain at least one non-whitespace character")
        return stripped

    @field_validator("url")
    @classmethod
    def validate_url(cls, value: str) -> str:
        """Normalize whitespace and ensure the URL uses http/https with a host.

        Trims leading/trailing whitespace before validation because
        ``urlparse("http://example.com ")`` returns ``netloc='example.com '``
        (with the trailing space) while ``hostname`` normalizes it, so a
        naive validator would accept the value and persist it with the
        embedded whitespace. The monitored URL is propagated to health-checker
        and alert-service and appears in structured logs; a trailing space
        breaks downstream HTTP clients and reads as a silent typo.

        Also rejects URLs that embed credentials (``http://user:pass@host``).
        Monitored URLs are propagated to the health-checker and alert-service
        and appear in structured logs, so accepting userinfo would risk
        leaking secrets through logs. Health check endpoints are expected to
        be reachable without in-URL credentials.
        """
        stripped = value.strip()
        if not stripped:
            raise ValueError("url must contain at least one non-whitespace character")
        try:
            parsed = urlparse(stripped)
        except (ValueError, TypeError) as exc:
            raise ValueError("url must be a valid URL") from exc
        if parsed.scheme not in ("http", "https"):
            raise ValueError("url must start with http:// or https://")
        if not parsed.netloc or not parsed.hostname:
            raise ValueError("url must include a host name")
        if parsed.username is not None or parsed.password is not None:
            raise ValueError("url must not contain embedded credentials")
        return stripped


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
