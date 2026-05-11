from __future__ import annotations

import subprocess
from pathlib import Path

import pytest

from pipeline import rust_validator


def test_validate_jsonl_reports_missing_cargo(monkeypatch: pytest.MonkeyPatch, tmp_path: Path) -> None:
    path = tmp_path / "windows.jsonl"
    path.write_text('{"window_start":"2026-05-10T12:00:00Z"}\n', encoding="utf-8")
    monkeypatch.setattr(rust_validator.shutil, "which", lambda _: None)

    with pytest.raises(rust_validator.RustValidatorUnavailable):
        rust_validator.validate_jsonl(path)


def test_validate_jsonl_raises_on_rust_error(monkeypatch: pytest.MonkeyPatch, tmp_path: Path) -> None:
    path = tmp_path / "windows.jsonl"
    path.write_text('{"topic":"ai"}\n', encoding="utf-8")
    monkeypatch.setattr(rust_validator.shutil, "which", lambda _: "cargo")

    def fake_run(*_args: object, **_kwargs: object) -> subprocess.CompletedProcess[str]:
        return subprocess.CompletedProcess(args=["cargo"], returncode=1, stdout="", stderr="line 1: invalid")

    monkeypatch.setattr(rust_validator.subprocess, "run", fake_run)

    with pytest.raises(ValueError, match="line 1"):
        rust_validator.validate_jsonl(path)
