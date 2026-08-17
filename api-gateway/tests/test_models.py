"""Pydantic モデル層の直接テスト。

`test_api.py` は FastAPI 経由で `ServiceCreate` の URL バリデーションを網羅するが、
`ServiceStatus` enum の値・`Service` のデフォルト値・`AlertConfig` の任意フィールド・
`ErrorResponse` の必須フィールドといった契約は、モデル単体では回帰していなかった。
これらを個別に固定することで、API を経由しないコード（例: バックグラウンドチェッカーが
`Service` を直接構築するケース）や、enum 値を参照する他サービス（alert-service）と
の互換破壊を検知する。
"""

from datetime import datetime, timezone

import pytest
from pydantic import ValidationError

from app.models import (
    AlertConfig,
    ErrorResponse,
    HealthResponse,
    Service,
    ServiceStatus,
)


class TestServiceStatus:
    """`ServiceStatus` enum の値と型の回帰。

    `PUT /services/{id}/status?status=<value>` の受理値・alert-service が参照する
    ステータス文字列の双方に影響するため、値そのものを固定する。
    """

    def test_healthy_value(self):
        assert ServiceStatus.HEALTHY.value == "healthy"

    def test_unhealthy_value(self):
        assert ServiceStatus.UNHEALTHY.value == "unhealthy"

    def test_unknown_value(self):
        assert ServiceStatus.UNKNOWN.value == "unknown"

    def test_is_str_subclass(self):
        # `class ServiceStatus(str, Enum)` の定義に依存する。
        # `str` サブクラスであることで JSON 直列化が値そのものになる契約を回帰する。
        assert isinstance(ServiceStatus.HEALTHY, str)
        assert ServiceStatus.HEALTHY == "healthy"

    def test_all_members(self):
        # メンバー数の増減があった場合に他所のテストと同期を取れるよう固定する。
        assert {s.value for s in ServiceStatus} == {"healthy", "unhealthy", "unknown"}


class TestService:
    """`Service` のデフォルト値と必須フィールドの回帰。"""

    def _base_kwargs(self) -> dict:
        return {
            "id": "svc-1",
            "name": "web",
            "url": "http://example.com",
            "interval_seconds": 30,
            "created_at": datetime(2026, 1, 1, tzinfo=timezone.utc),
        }

    def test_status_defaults_to_unknown(self):
        svc = Service(**self._base_kwargs())
        assert svc.status == ServiceStatus.UNKNOWN

    def test_last_checked_defaults_to_none(self):
        svc = Service(**self._base_kwargs())
        assert svc.last_checked is None

    def test_explicit_status_accepted(self):
        kwargs = self._base_kwargs()
        kwargs["status"] = ServiceStatus.HEALTHY
        svc = Service(**kwargs)
        assert svc.status == ServiceStatus.HEALTHY

    def test_status_accepts_string_value(self):
        # `str` サブクラス enum のため文字列でも受理される契約
        kwargs = self._base_kwargs()
        kwargs["status"] = "unhealthy"
        svc = Service(**kwargs)
        assert svc.status == ServiceStatus.UNHEALTHY

    def test_created_at_missing_raises(self):
        kwargs = self._base_kwargs()
        del kwargs["created_at"]
        with pytest.raises(ValidationError):
            Service(**kwargs)

    def test_id_missing_raises(self):
        kwargs = self._base_kwargs()
        del kwargs["id"]
        with pytest.raises(ValidationError):
            Service(**kwargs)


class TestAlertConfig:
    """`AlertConfig` の任意フィールド契約の回帰。

    現状 `webhook_url` / `email` はいずれも `Optional[str]`（デフォルト `None`）で、
    `service_id` のみ必須。将来「少なくとも 1 つの通知先が必須」等のルール強化を入れる
    場合は、この振る舞いテストを更新した上でバリデータを追加すること。
    """

    def test_only_service_id_required(self):
        config = AlertConfig(service_id="svc-1")
        assert config.service_id == "svc-1"
        assert config.webhook_url is None
        assert config.email is None

    def test_with_webhook_only(self):
        config = AlertConfig(service_id="svc-1", webhook_url="https://hooks.example.com/x")
        assert config.webhook_url == "https://hooks.example.com/x"
        assert config.email is None

    def test_with_email_only(self):
        config = AlertConfig(service_id="svc-1", email="ops@example.com")
        assert config.email == "ops@example.com"
        assert config.webhook_url is None

    def test_with_both_channels(self):
        config = AlertConfig(
            service_id="svc-1",
            webhook_url="https://hooks.example.com/x",
            email="ops@example.com",
        )
        assert config.webhook_url == "https://hooks.example.com/x"
        assert config.email == "ops@example.com"

    def test_missing_service_id_raises(self):
        with pytest.raises(ValidationError):
            AlertConfig()  # type: ignore[call-arg]


class TestErrorResponse:
    """`ErrorResponse` の必須フィールドの回帰。"""

    def test_valid(self):
        err = ErrorResponse(error="not_found", detail="service not found")
        assert err.error == "not_found"
        assert err.detail == "service not found"

    def test_error_missing_raises(self):
        with pytest.raises(ValidationError):
            ErrorResponse(detail="x")  # type: ignore[call-arg]

    def test_detail_missing_raises(self):
        with pytest.raises(ValidationError):
            ErrorResponse(error="x")  # type: ignore[call-arg]


class TestHealthResponse:
    """`HealthResponse` の必須フィールドの回帰。

    `test_api.py::test_health` は API レスポンスに `status`/`service`/`version`/
    `timestamp` が存在することは検査するが、モデルのフィールド必須性はテストしていない。
    """

    def test_valid(self):
        resp = HealthResponse(
            status="ok",
            service="api-gateway",
            version="1.0.0",
            timestamp=datetime(2026, 1, 1, tzinfo=timezone.utc),
        )
        assert resp.status == "ok"
        assert resp.service == "api-gateway"
        assert resp.version == "1.0.0"

    def test_missing_field_raises(self):
        with pytest.raises(ValidationError):
            HealthResponse(  # type: ignore[call-arg]
                status="ok",
                service="api-gateway",
                version="1.0.0",
            )
