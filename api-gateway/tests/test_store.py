"""Unit tests for ``app.store.ServiceStore``.

``store.py`` は現在 ``test_api.py`` から HTTP 経由で間接的に検証されているのみで、
純粋なストア層としてのユニットテストが無い。API を経由しないユニットテストを追加し、
ストア層の各メソッド（create/get/list_all/update_status/delete）の挙動を独立して固定する。
"""
from datetime import datetime, timezone

import pytest

from app.models import ServiceCreate, ServiceStatus
from app.store import ServiceStore


@pytest.fixture
def store() -> ServiceStore:
    """毎テストで独立したストアを提供し、テスト間の状態リークを防ぐ。"""
    return ServiceStore()


@pytest.fixture
def sample_data() -> ServiceCreate:
    return ServiceCreate(
        name="web-app",
        url="https://example.com/health",
        interval_seconds=60,
    )


def test_create_returns_service_with_generated_id(store, sample_data):
    before = datetime.now(timezone.utc)
    service = store.create(sample_data)
    after = datetime.now(timezone.utc)

    assert service.id
    assert isinstance(service.id, str)
    assert service.name == sample_data.name
    assert service.url == sample_data.url
    assert service.interval_seconds == sample_data.interval_seconds
    assert service.status == ServiceStatus.UNKNOWN
    assert service.last_checked is None
    # created_at は UTC の aware datetime で、生成前後の範囲に収まる
    assert service.created_at.tzinfo is not None
    assert before <= service.created_at <= after


def test_create_uses_default_interval_when_not_specified(store):
    service = store.create(ServiceCreate(name="svc", url="http://a.com"))
    assert service.interval_seconds == 30


def test_create_generates_unique_ids(store, sample_data):
    a = store.create(sample_data)
    b = store.create(sample_data)
    assert a.id != b.id


def test_get_returns_created_service(store, sample_data):
    created = store.create(sample_data)
    fetched = store.get(created.id)
    assert fetched is not None
    assert fetched.id == created.id
    assert fetched.name == created.name


def test_get_returns_none_for_unknown_id(store):
    assert store.get("does-not-exist") is None


def test_list_all_returns_empty_when_no_services(store):
    assert store.list_all() == []


def test_list_all_returns_all_created_services(store):
    a = store.create(ServiceCreate(name="svc1", url="http://a.com"))
    b = store.create(ServiceCreate(name="svc2", url="http://b.com"))
    ids = {s.id for s in store.list_all()}
    assert ids == {a.id, b.id}


def test_list_all_returns_independent_snapshot(store, sample_data):
    """list_all() の返り値は内部辞書の list であり、外部からの mutation が
    内部状態を破壊してはならない。"""
    store.create(sample_data)
    snapshot = store.list_all()
    snapshot.clear()
    assert len(store.list_all()) == 1


def test_update_status_sets_status_and_last_checked(store, sample_data):
    created = store.create(sample_data)
    assert created.last_checked is None

    before = datetime.now(timezone.utc)
    updated = store.update_status(created.id, ServiceStatus.HEALTHY)
    after = datetime.now(timezone.utc)

    assert updated is not None
    assert updated.id == created.id
    assert updated.status == ServiceStatus.HEALTHY
    assert updated.last_checked is not None
    assert before <= updated.last_checked <= after
    # ストア内部の値も更新されている
    assert store.get(created.id).status == ServiceStatus.HEALTHY


def test_update_status_returns_none_for_unknown_id(store):
    assert store.update_status("nope", ServiceStatus.HEALTHY) is None


def test_update_status_can_transition_between_states(store, sample_data):
    created = store.create(sample_data)
    store.update_status(created.id, ServiceStatus.HEALTHY)
    updated = store.update_status(created.id, ServiceStatus.UNHEALTHY)
    assert updated is not None
    assert updated.status == ServiceStatus.UNHEALTHY


def test_delete_removes_existing_service(store, sample_data):
    created = store.create(sample_data)
    assert store.delete(created.id) is True
    assert store.get(created.id) is None
    assert store.list_all() == []


def test_delete_returns_false_for_unknown_id(store):
    assert store.delete("nope") is False


def test_delete_only_removes_specified_service(store):
    a = store.create(ServiceCreate(name="svc1", url="http://a.com"))
    b = store.create(ServiceCreate(name="svc2", url="http://b.com"))
    assert store.delete(a.id) is True
    assert store.get(a.id) is None
    assert store.get(b.id) is not None
