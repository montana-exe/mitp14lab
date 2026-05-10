from __future__ import annotations

from pathlib import Path

import polars as pl

from pipeline.analyzer import add_indicators, aggregate_by_topic, clean_windows, run


def sample_frame() -> pl.DataFrame:
    return pl.DataFrame(
        [
            {
                "window_start": "2026-05-10T12:00:00Z",
                "window_end": "2026-05-10T12:01:00Z",
                "topic": "AI",
                "post_count": 3,
                "positive_count": 2,
                "negative_count": 1,
                "neutral_count": 0,
                "min_sentiment": -0.2,
                "max_sentiment": 0.7,
                "avg_sentiment": 0.3,
                "total_engagement": 120,
                "unique_authors": 3,
            },
            {
                "window_start": "2026-05-10T12:01:00Z",
                "window_end": "2026-05-10T12:02:00Z",
                "topic": "ai",
                "post_count": 2,
                "positive_count": 1,
                "negative_count": 0,
                "neutral_count": 1,
                "min_sentiment": 0.0,
                "max_sentiment": 0.5,
                "avg_sentiment": 0.25,
                "total_engagement": 80,
                "unique_authors": 2,
            },
        ]
    )


def test_clean_windows_normalizes_topic_and_types() -> None:
    cleaned = clean_windows(sample_frame())

    assert cleaned["topic"].to_list() == ["ai", "ai"]
    assert cleaned["window_start"].dtype.is_temporal()


def test_aggregate_by_topic() -> None:
    result = aggregate_by_topic(add_indicators(clean_windows(sample_frame())))

    assert result.to_dicts()[0]["posts"] == 5
    assert result.to_dicts()[0]["engagement"] == 200


def test_run_writes_reports(tmp_path: Path) -> None:
    input_path = tmp_path / "windows.jsonl"
    sample_frame().write_ndjson(input_path)

    result = run(input_path, tmp_path / "reports")

    assert result["rows"] == 2
    assert (tmp_path / "reports" / "social_windows.parquet").exists()
    assert (tmp_path / "reports" / "topic_summary.csv").exists()
