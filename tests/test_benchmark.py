from __future__ import annotations

from pathlib import Path

from pipeline import benchmark


def test_benchmark_rust_validator_reports_unavailable_without_cargo(
    monkeypatch, tmp_path: Path
) -> None:
    jsonl_path = tmp_path / "posts.jsonl"
    jsonl_path.write_text("{}\n", encoding="utf-8")
    monkeypatch.setattr(benchmark.shutil, "which", lambda _: None)

    result = benchmark.benchmark_rust_validator(tmp_path, jsonl_path)

    assert result["status"] == "unavailable"
