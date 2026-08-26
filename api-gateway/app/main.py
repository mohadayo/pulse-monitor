import logging
from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware

from app.config import settings
from app.models import (
    ServiceCreate,
    Service,
    ServiceStatus,
    HealthResponse,
    ErrorResponse,
)
from app.store import store

logging.basicConfig(
    level=getattr(logging, settings.LOG_LEVEL),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("api-gateway")

app = FastAPI(
    title=settings.APP_NAME,
    version=settings.APP_VERSION,
)


# 全応答に付与するセキュリティレスポンスヘッダ。
# JSON API サーバとして外部依存 (secure-headers 等) を追加せずに、
# starlette 標準の BaseHTTPMiddleware で以下を付ける:
#
# - `X-Content-Type-Options: nosniff` … JSON エンドポイントを別 MIME として
#   解釈させる MIME sniffing 攻撃を抑止。
# - `X-Frame-Options: DENY` … API を `<iframe>` に埋め込ませない。
#   JSON API はフレーム表示を意図しないため常時拒否 (clickjacking 対策)。
# - `Referrer-Policy: no-referrer` … 内部 URL やクエリ文字列がリンク先の
#   Referrer ヘッダとして外部に漏れないよう抑止。
#
# `if key not in response.headers` により既存ヘッダは上書きしない
# (`setdefault` 相当) — 将来 per-route オーバーライドやテストを壊さない。
_SECURITY_HEADERS = {
    "X-Content-Type-Options": "nosniff",
    "X-Frame-Options": "DENY",
    "Referrer-Policy": "no-referrer",
}


class SecurityHeadersMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        response = await call_next(request)
        for key, value in _SECURITY_HEADERS.items():
            if key not in response.headers:
                response.headers[key] = value
        return response


app.add_middleware(SecurityHeadersMiddleware)


@app.get("/health", response_model=HealthResponse)
async def health():
    logger.info("Health check requested")
    return HealthResponse(
        status="ok",
        service="api-gateway",
        version=settings.APP_VERSION,
        timestamp=datetime.now(timezone.utc),
    )


@app.post("/services", response_model=Service, status_code=201)
async def create_service(data: ServiceCreate):
    logger.info("Creating service: %s", data.name)
    service = store.create(data)
    logger.info("Service created: %s (id=%s)", service.name, service.id)
    return service


@app.get("/services", response_model=list[Service])
async def list_services():
    logger.info("Listing all services")
    return store.list_all()


@app.get("/services/{service_id}", response_model=Service)
async def get_service(service_id: str):
    logger.info("Getting service: %s", service_id)
    service = store.get(service_id)
    if service is None:
        logger.warning("Service not found: %s", service_id)
        raise HTTPException(status_code=404, detail="Service not found")
    return service


@app.put("/services/{service_id}/status")
async def update_service_status(service_id: str, status: ServiceStatus):
    logger.info("Updating service %s status to %s", service_id, status)
    service = store.update_status(service_id, status)
    if service is None:
        raise HTTPException(status_code=404, detail="Service not found")
    return service


@app.delete("/services/{service_id}", status_code=204)
async def delete_service(service_id: str):
    logger.info("Deleting service: %s", service_id)
    if not store.delete(service_id):
        raise HTTPException(status_code=404, detail="Service not found")
    return None


@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    logger.error("Unhandled exception: %s", str(exc))
    return JSONResponse(
        status_code=500,
        content=ErrorResponse(
            error="internal_server_error",
            detail="An unexpected error occurred",
        ).model_dump(),
    )
