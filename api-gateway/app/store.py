from datetime import datetime, timezone
from typing import Optional
import uuid

from app.models import Service, ServiceCreate, ServiceStatus


class ServiceStore:
    def __init__(self):
        self._services: dict[str, Service] = {}

    def create(self, data: ServiceCreate) -> Service:
        service_id = str(uuid.uuid4())
        service = Service(
            id=service_id,
            name=data.name,
            url=data.url,
            interval_seconds=data.interval_seconds,
            status=ServiceStatus.UNKNOWN,
            last_checked=None,
            created_at=datetime.now(timezone.utc),
        )
        self._services[service_id] = service
        return service

    def get(self, service_id: str) -> Optional[Service]:
        return self._services.get(service_id)

    def list_all(self) -> list[Service]:
        return list(self._services.values())

    def update_status(self, service_id: str, status: ServiceStatus) -> Optional[Service]:
        service = self._services.get(service_id)
        if service is None:
            return None
        service.status = status
        service.last_checked = datetime.now(timezone.utc)
        self._services[service_id] = service
        return service

    def delete(self, service_id: str) -> bool:
        if service_id in self._services:
            del self._services[service_id]
            return True
        return False


store = ServiceStore()
