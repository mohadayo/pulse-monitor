"""Tests for ``app.config`` env-var validation helpers.

``Settings`` reads env vars at class definition time, so full-Settings behavior
is exercised via ``importlib.reload``. The finer-grained cases go through the
private helpers to keep tests fast and isolated.
"""
import importlib
import logging

import pytest

from app import config as config_module
from app.config import _resolve_log_level, _resolve_port


class TestResolveLogLevel:
    @pytest.mark.parametrize(
        "raw,expected",
        [
            ("DEBUG", "DEBUG"),
            ("INFO", "INFO"),
            ("WARNING", "WARNING"),
            ("ERROR", "ERROR"),
            ("CRITICAL", "CRITICAL"),
        ],
    )
    def test_accepts_canonical_upper(self, raw, expected):
        assert _resolve_log_level(raw) == expected

    @pytest.mark.parametrize("raw", ["debug", "Info", "warning", "  ERROR  "])
    def test_normalizes_case_and_whitespace(self, raw):
        # Any of these must survive the pass-through and remain valid.
        resolved = _resolve_log_level(raw)
        assert resolved in {"DEBUG", "INFO", "WARNING", "ERROR"}
        # And crucially, the resolved value must map to a real logging level.
        assert isinstance(getattr(logging, resolved), int)

    def test_warn_aliases_to_warning(self):
        # ``logging`` accepts WARN, but the canonical attribute is WARNING.
        assert _resolve_log_level("WARN") == "WARNING"
        assert _resolve_log_level("warn") == "WARNING"

    @pytest.mark.parametrize("raw", ["", None])
    def test_empty_falls_back_to_info(self, raw):
        assert _resolve_log_level(raw) == "INFO"

    @pytest.mark.parametrize("raw", ["TRACE", "verbose", "SHOUT", "1", "!!"])
    def test_invalid_falls_back_to_info_and_warns(self, raw, capsys):
        assert _resolve_log_level(raw) == "INFO"
        stderr = capsys.readouterr().err
        assert "invalid LOG_LEVEL" in stderr
        assert repr(raw) in stderr

    def test_resolved_value_is_always_a_valid_logging_attribute(self):
        # Guard for the actual bug: main.py does getattr(logging, LOG_LEVEL),
        # which must never AttributeError regardless of user input.
        for raw in ["debug", "INFO", "warn", "TRACE", "", None, "garbage"]:
            resolved = _resolve_log_level(raw)
            assert hasattr(logging, resolved), f"logging has no attr {resolved!r}"


class TestResolvePort:
    @pytest.mark.parametrize("raw,expected", [("8000", 8000), ("1", 1), ("65535", 65535), ("8080", 8080)])
    def test_accepts_valid_ports(self, raw, expected):
        assert _resolve_port(raw) == expected

    @pytest.mark.parametrize("raw", ["", None])
    def test_empty_falls_back_to_default(self, raw):
        assert _resolve_port(raw) == 8000

    @pytest.mark.parametrize("raw", ["abc", "8000abc", "8.5", "0x1F"])
    def test_non_numeric_falls_back_and_warns(self, raw, capsys):
        assert _resolve_port(raw) == 8000
        stderr = capsys.readouterr().err
        assert "invalid API_PORT" in stderr

    @pytest.mark.parametrize("raw", ["0", "-1", "65536", "999999"])
    def test_out_of_range_falls_back_and_warns(self, raw, capsys):
        assert _resolve_port(raw) == 8000
        stderr = capsys.readouterr().err
        assert "out of range" in stderr


class TestSettingsReload:
    """End-to-end: reload the module under env-var overrides and check attrs."""

    def _reload_with_env(self, monkeypatch, **env):
        for k, v in env.items():
            if v is None:
                monkeypatch.delenv(k, raising=False)
            else:
                monkeypatch.setenv(k, v)
        importlib.reload(config_module)
        return config_module.settings

    def test_defaults_when_env_absent(self, monkeypatch):
        s = self._reload_with_env(monkeypatch, LOG_LEVEL=None, API_PORT=None)
        assert s.PORT == 8000
        assert s.LOG_LEVEL == "INFO"

    def test_bad_env_still_produces_usable_settings(self, monkeypatch):
        # The whole point: importing main.py after a typo in LOG_LEVEL must
        # not raise, and getattr(logging, ...) must resolve.
        s = self._reload_with_env(monkeypatch, LOG_LEVEL="trace", API_PORT="not-a-port")
        assert s.PORT == 8000
        assert s.LOG_LEVEL == "INFO"
        assert isinstance(getattr(logging, s.LOG_LEVEL), int)

    def test_valid_lowercase_log_level_normalized(self, monkeypatch):
        s = self._reload_with_env(monkeypatch, LOG_LEVEL="debug", API_PORT="9001")
        assert s.LOG_LEVEL == "DEBUG"
        assert s.PORT == 9001
