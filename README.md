# Lab 14 - Social Network ETL Pipeline

## Author

- Full name: Savenkov Denis Yurievich
- Group: 221331
- Variant: 12
- Domain: social network monitoring, Twitter/X API emulation
- Level: advanced

## Architecture

```text
Go collector
  -> synthetic Twitter/X-like post stream
  -> batch JSONL writer
  -> tumbling-window aggregation
  -> Apache Arrow HTTP endpoint

Python analytics
  -> Polars cleanup and validation
  -> Parquet storage
  -> DuckDB SQL report
  -> Plotly HTML charts
  -> Streamlit dashboard
```

## Implemented Requirements

| Area | Implementation |
| --- | --- |
| Go data collection | Deterministic Twitter/X emulator with topics, sentiment, engagement and author ids |
| Buffered writes | Generic batch JSONL writer with size/time flush |
| Graceful shutdown | `signal.NotifyContext`, final buffer flush and HTTP server shutdown |
| Window aggregation | Tumbling windows by topic with count, min/max/avg sentiment, engagement and unique authors |
| Apache Arrow | Go HTTP endpoint `/arrow` returns Arrow IPC stream |
| Python analysis | Polars cleanup, indicators, Parquet output and DuckDB report |
| Visualization | Plotly HTML reports and Streamlit dashboard |
| Performance | `pipeline.benchmark` compares Go collector and Python asyncio collector |
| Tests | Go unit tests and Python pytest tests |

## Local Setup

```powershell
python -m venv .venv
.\.venv\Scripts\python.exe -m pip install -U pip
.\.venv\Scripts\python.exe -m pip install -r requirements.txt
```

## Generate Data

```powershell
cd collector
go run ./cmd/collector -count 360 -out ../data -batch 60 -window 1m
```

Generated files:

- `data/posts.jsonl`
- `data/windows.jsonl`

## Analyze Data

```powershell
.\.venv\Scripts\python.exe -m pipeline.analyzer --input data/windows.jsonl --out reports
```

Generated reports:

- `reports/social_windows.parquet`
- `reports/topic_summary.csv`
- `reports/sentiment_trend.html`
- `reports/engagement.html`

## Apache Arrow Server

```powershell
cd collector
go run ./cmd/collector -count 360 -out ../data -batch 60 -serve -addr :8080
```

Endpoints:

- `GET /health`
- `GET /arrow`
- `GET /arrow?topic=ai`

## Dashboard

```powershell
$env:PARQUET_PATH="reports/social_windows.parquet"
.\.venv\Scripts\streamlit.exe run dashboard/app.py
```

Or through Compose:

```powershell
docker compose up --build
```

Dashboard URL: `http://127.0.0.1:8501`.

## Tests

```powershell
cd collector
go test ./...
```

```powershell
.\.venv\Scripts\python.exe -m pytest -q
```

## Performance Check

```powershell
.\.venv\Scripts\python.exe -m pipeline.benchmark --count 1000 --out reports/performance.json
```

The benchmark measures elapsed time and throughput for the Go collector and the Python asyncio emulator. The output file is ignored by git because it is generated evidence, not source code.

## Project Structure

```text
collector/          Go collector, aggregator, Arrow endpoint and Go tests
pipeline/           Python analytics, Arrow client and benchmark
dashboard/          Streamlit monitoring UI
tests/              Python tests
.github/workflows/  CI for Go, Python and smoke pipeline
docker-compose.yml  Collector + dashboard stack
PROMPT_LOG.md       Realistic AI usage log
```
