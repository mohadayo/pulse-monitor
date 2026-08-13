import os
import sys


# ``logging`` モジュールが解釈できるレベル名。``WARN`` は ``WARNING`` の
# 別名 (`logging` モジュールが認めている) として受理する。
_ALLOWED_LOG_LEVELS: frozenset[str] = frozenset(
    {"DEBUG", "INFO", "WARNING", "WARN", "ERROR", "CRITICAL"}
)
_DEFAULT_LOG_LEVEL = "INFO"

_DEFAULT_PORT = 8000
_MIN_PORT = 1
_MAX_PORT = 65535


def _resolve_log_level(raw: str) -> str:
    """許可済みの LOG_LEVEL に正規化する。

    許可されていない値が来た場合はデフォルト (``INFO``) にフォールバックし、
    運用者が原因を追えるよう stderr に警告を書き出す。プロセスは落とさない。
    """
    if raw is None or raw == "":
        return _DEFAULT_LOG_LEVEL
    normalized = raw.strip().upper()
    if normalized in _ALLOWED_LOG_LEVELS:
        # ``logging`` は ``WARN`` も受け付けるが正式名は ``WARNING`` なので
        # 呼び出し側の getattr(...) が確実に引けるよう寄せておく。
        return "WARNING" if normalized == "WARN" else normalized
    print(
        f"[config] WARNING: invalid LOG_LEVEL={raw!r}; "
        f"falling back to {_DEFAULT_LOG_LEVEL}. "
        f"Allowed values: {sorted(_ALLOWED_LOG_LEVELS)}",
        file=sys.stderr,
    )
    return _DEFAULT_LOG_LEVEL


def _resolve_port(raw: str) -> int:
    """API_PORT を 1〜65535 の整数として解決する。

    数値化に失敗した場合や範囲外の場合はデフォルト (``8000``) を返し、
    stderr に警告を書き出す。プロセスは落とさない。
    """
    if raw is None or raw == "":
        return _DEFAULT_PORT
    try:
        port = int(raw)
    except (TypeError, ValueError):
        print(
            f"[config] WARNING: invalid API_PORT={raw!r}; "
            f"falling back to {_DEFAULT_PORT}.",
            file=sys.stderr,
        )
        return _DEFAULT_PORT
    if port < _MIN_PORT or port > _MAX_PORT:
        print(
            f"[config] WARNING: API_PORT={port} is out of range "
            f"[{_MIN_PORT}, {_MAX_PORT}]; falling back to {_DEFAULT_PORT}.",
            file=sys.stderr,
        )
        return _DEFAULT_PORT
    return port


class Settings:
    APP_NAME: str = os.getenv("APP_NAME", "Pulse API Gateway")
    APP_VERSION: str = os.getenv("APP_VERSION", "1.0.0")
    HOST: str = os.getenv("API_HOST", "0.0.0.0")
    PORT: int = _resolve_port(os.getenv("API_PORT", str(_DEFAULT_PORT)))
    LOG_LEVEL: str = _resolve_log_level(os.getenv("LOG_LEVEL", _DEFAULT_LOG_LEVEL))
    HEALTH_CHECKER_URL: str = os.getenv("HEALTH_CHECKER_URL", "http://health-checker:8001")
    ALERT_SERVICE_URL: str = os.getenv("ALERT_SERVICE_URL", "http://alert-service:8002")


settings = Settings()
