from __future__ import annotations

import argparse
from pathlib import Path
from typing import Any

import duckdb
import plotly.express as px
import polars as pl

from pipeline.rust_validator import validate_jsonl

WINDOW_SCHEMA = {
    "window_start": pl.String,
    "window_end": pl.String,
    "topic": pl.String,
    "post_count": pl.Int64,
    "positive_count": pl.Int64,
    "negative_count": pl.Int64,
    "neutral_count": pl.Int64,
    "min_sentiment": pl.Float64,
    "max_sentiment": pl.Float64,
    "avg_sentiment": pl.Float64,
    "total_engagement": pl.Int64,
    "unique_authors": pl.Int64,
}


def load_windows(path: Path) -> pl.DataFrame:
    if not path.exists():
        raise FileNotFoundError(f"window file not found: {path}")
    return pl.read_ndjson(path, schema=WINDOW_SCHEMA)


def clean_windows(df: pl.DataFrame) -> pl.DataFrame:
    required = set(WINDOW_SCHEMA)
    missing = required - set(df.columns)
    if missing:
        raise ValueError(f"missing columns: {sorted(missing)}")

    cleaned = (
        df.unique(subset=["window_start", "topic"], keep="last")
        .with_columns(
            pl.col("window_start").str.to_datetime("%Y-%m-%dT%H:%M:%S%#z", strict=False).alias("window_start"),
            pl.col("window_end").str.to_datetime("%Y-%m-%dT%H:%M:%S%#z", strict=False).alias("window_end"),
            pl.col("topic").str.to_lowercase().str.strip_chars().alias("topic"),
            pl.col("post_count").fill_null(0).clip(0),
            pl.col("positive_count").fill_null(0).clip(0),
            pl.col("negative_count").fill_null(0).clip(0),
            pl.col("neutral_count").fill_null(0).clip(0),
            pl.col("avg_sentiment").fill_null(0).clip(-1, 1),
            pl.col("min_sentiment").fill_null(0).clip(-1, 1),
            pl.col("max_sentiment").fill_null(0).clip(-1, 1),
            pl.col("total_engagement").fill_null(0).clip(0),
            pl.col("unique_authors").fill_null(0).clip(0),
        )
        .filter(pl.col("topic").is_not_null() & (pl.col("topic") != ""))
        .sort(["window_start", "topic"])
    )
    return cleaned


def add_indicators(df: pl.DataFrame, sma_window: int = 3) -> pl.DataFrame:
    return (
        df.with_columns(
            pl.col("avg_sentiment")
            .rolling_mean(window_size=sma_window, min_samples=1)
            .over("topic")
            .alias("sentiment_sma"),
            (pl.col("positive_count") - pl.col("negative_count")).alias("sentiment_delta"),
        )
        .with_columns(
            (
                100
                - 100
                / (
                    1
                    + (
                        pl.col("positive_count") + 1
                    )
                    / (pl.col("negative_count") + 1)
                )
            ).alias("social_rsi")
        )
    )


def aggregate_by_topic(df: pl.DataFrame) -> pl.DataFrame:
    return (
        df.group_by("topic")
        .agg(
            pl.col("post_count").sum().alias("posts"),
            pl.col("total_engagement").sum().alias("engagement"),
            pl.col("avg_sentiment").mean().round(3).alias("avg_sentiment"),
            pl.col("min_sentiment").min().alias("min_sentiment"),
            pl.col("max_sentiment").max().alias("max_sentiment"),
            pl.col("unique_authors").max().alias("max_unique_authors"),
        )
        .sort("engagement", descending=True)
    )


def run_duckdb_report(parquet_path: Path) -> list[dict[str, Any]]:
    with duckdb.connect(database=":memory:") as conn:
        rows = conn.execute(
            """
            SELECT
                topic,
                SUM(post_count) AS posts,
                SUM(total_engagement) AS engagement,
                ROUND(AVG(avg_sentiment), 3) AS avg_sentiment,
                MAX(unique_authors) AS max_unique_authors
            FROM read_parquet(?)
            GROUP BY topic
            ORDER BY engagement DESC, topic
            """,
            [str(parquet_path)],
        ).fetchall()
    return [
        {
            "topic": row[0],
            "posts": int(row[1]),
            "engagement": int(row[2]),
            "avg_sentiment": float(row[3]),
            "max_unique_authors": int(row[4]),
        }
        for row in rows
    ]


def write_plots(df: pl.DataFrame, report_dir: Path) -> None:
    report_dir.mkdir(parents=True, exist_ok=True)
    pdf = df.to_pandas()
    trend = px.line(
        pdf,
        x="window_start",
        y="avg_sentiment",
        color="topic",
        markers=True,
        title="Average sentiment by topic",
    )
    trend.write_html(report_dir / "sentiment_trend.html")

    engagement = px.bar(
        pdf,
        x="topic",
        y="total_engagement",
        color="topic",
        title="Window engagement by topic",
    )
    engagement.write_html(report_dir / "engagement.html")


def run(input_path: Path, output_dir: Path, validate_with_rust: bool = False) -> dict[str, Any]:
    if validate_with_rust:
        validate_jsonl(input_path)
    output_dir.mkdir(parents=True, exist_ok=True)
    cleaned = add_indicators(clean_windows(load_windows(input_path)))
    parquet_path = output_dir / "social_windows.parquet"
    cleaned.write_parquet(parquet_path)
    summary = aggregate_by_topic(cleaned)
    summary.write_csv(output_dir / "topic_summary.csv")
    duckdb_rows = run_duckdb_report(parquet_path)
    write_plots(cleaned, output_dir)
    return {
        "rows": cleaned.height,
        "topics": cleaned.select("topic").unique().height,
        "parquet": str(parquet_path),
        "duckdb_report": duckdb_rows,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Analyze lab14 social media windows")
    parser.add_argument("--input", type=Path, default=Path("data/windows.jsonl"))
    parser.add_argument("--out", type=Path, default=Path("reports"))
    parser.add_argument("--rust-validate", action="store_true", help="validate input JSONL with Rust before analysis")
    args = parser.parse_args()
    result = run(args.input, args.out, validate_with_rust=args.rust_validate)
    print(result)


if __name__ == "__main__":
    main()
