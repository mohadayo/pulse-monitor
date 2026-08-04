import pytest
from fastapi.testclient import TestClient

from app.main import app
from app.store import store


@pytest.fixture(autouse=True)
def clear_store():
    store._services.clear()
    yield
    store._services.clear()


@pytest.fixture
def client():
    return TestClient(app)


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "ok"
    assert data["service"] == "api-gateway"
    assert "version" in data
    assert "timestamp" in data


def test_create_service(client):
    resp = client.post("/services", json={
        "name": "web-app",
        "url": "http://example.com/health",
        "interval_seconds": 60,
    })
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == "web-app"
    assert data["url"] == "http://example.com/health"
    assert data["interval_seconds"] == 60
    assert data["status"] == "unknown"
    assert "id" in data


def test_create_service_validation(client):
    resp = client.post("/services", json={"name": "", "url": "http://x.com"})
    assert resp.status_code == 422


@pytest.mark.parametrize(
    "bad_url",
    [
        "foobar",              # no scheme
        "http://",             # no host
        "https://",            # no host
        "ftp://example.com",   # unsupported scheme
        "://example.com",      # empty scheme
        "example.com",         # scheme missing
        "http:///path",        # empty host with path
    ],
)
def test_create_service_rejects_invalid_url(client, bad_url):
    resp = client.post(
        "/services",
        json={"name": "svc", "url": bad_url},
    )
    assert resp.status_code == 422, f"expected 422 for {bad_url!r}, got {resp.status_code}"


@pytest.mark.parametrize(
    "good_url",
    [
        "http://example.com",
        "https://example.com/health",
        "http://localhost:8080/status",
        "https://sub.example.co.jp:8443/api/v1/health",
    ],
)
def test_create_service_accepts_valid_url(client, good_url):
    resp = client.post(
        "/services",
        json={"name": "svc", "url": good_url},
    )
    assert resp.status_code == 201, f"expected 201 for {good_url!r}, got {resp.status_code}"
    assert resp.json()["url"] == good_url


def test_create_service_interval_out_of_range(client):
    resp = client.post(
        "/services",
        json={"name": "svc", "url": "http://a.com", "interval_seconds": 4},
    )
    assert resp.status_code == 422
    resp = client.post(
        "/services",
        json={"name": "svc", "url": "http://a.com", "interval_seconds": 3601},
    )
    assert resp.status_code == 422


def test_list_services(client):
    client.post("/services", json={"name": "svc1", "url": "http://a.com"})
    client.post("/services", json={"name": "svc2", "url": "http://b.com"})
    resp = client.get("/services")
    assert resp.status_code == 200
    assert len(resp.json()) == 2


def test_get_service(client):
    create_resp = client.post("/services", json={"name": "svc", "url": "http://a.com"})
    svc_id = create_resp.json()["id"]
    resp = client.get(f"/services/{svc_id}")
    assert resp.status_code == 200
    assert resp.json()["name"] == "svc"


def test_get_service_not_found(client):
    resp = client.get("/services/nonexistent")
    assert resp.status_code == 404


def test_update_service_status(client):
    create_resp = client.post("/services", json={"name": "svc", "url": "http://a.com"})
    svc_id = create_resp.json()["id"]
    resp = client.put(f"/services/{svc_id}/status?status=healthy")
    assert resp.status_code == 200
    assert resp.json()["status"] == "healthy"
    assert resp.json()["last_checked"] is not None


def test_delete_service(client):
    create_resp = client.post("/services", json={"name": "svc", "url": "http://a.com"})
    svc_id = create_resp.json()["id"]
    resp = client.delete(f"/services/{svc_id}")
    assert resp.status_code == 204
    resp = client.get(f"/services/{svc_id}")
    assert resp.status_code == 404


def test_delete_service_not_found(client):
    resp = client.delete("/services/nonexistent")
    assert resp.status_code == 404
