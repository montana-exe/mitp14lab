# Prompt Log

## Task 1. Go collector for social network monitoring

### Prompt 1
**Tool:** Codex GPT-5  
**Prompt:** "Implement lab 14, variant 12: an ETL pipeline for Twitter/X-like social network monitoring. Use Go for collection and Python for analytics."  
**Result:** Created the first design: synthetic social stream instead of real Twitter/X API because public API keys are not available in the reviewer environment.

### Prompt 2
**Tool:** Codex GPT-5  
**Prompt:** "Make the Go collector production-like: buffered writes, graceful shutdown, testable package and no tracked generated data."  
**Result:** Added deterministic simulator, generic JSONL batch writer, signal handling and Go unit tests.

### Prompt 3
**Tool:** Codex GPT-5  
**Prompt:** "Add tumbling-window aggregation on the Go side before Python receives data."  
**Result:** Added per-topic window metrics: post counts, positive/negative/neutral counts, min/max/avg sentiment, engagement and unique authors.

### Summary
- Prompt count: 3
- Manual fixes: split core logic into `collector/internal/social` so tests do not execute the CLI.
- Time: about 55 minutes

---

## Task 2. Apache Arrow transfer

### Prompt 1
**Tool:** Codex GPT-5  
**Prompt:** "Expose aggregated window metrics through an Apache Arrow endpoint from Go."  
**Result:** Added `github.com/apache/arrow/go/v17`, Arrow schema and `/arrow` HTTP endpoint.

### Prompt 2
**Tool:** Codex GPT-5  
**Prompt:** "Keep Arrow schema simple and stable for the Python client."  
**Result:** Used string timestamps and primitive numeric columns to avoid timezone incompatibilities between Go Arrow and Polars.

### Summary
- Prompt count: 2
- Manual fixes: ran `go mod tidy` and verified `go test ./...`.
- Time: about 30 minutes

---

## Task 3. Python Polars and DuckDB analysis

### Prompt 1
**Tool:** Codex GPT-5  
**Prompt:** "Build a Python analyzer: load JSONL, clean data with Polars, save Parquet, run DuckDB SQL and create plots."  
**Result:** Added `pipeline/analyzer.py`, requirements and pytest coverage.

### Prompt 2
**Tool:** Codex GPT-5  
**Prompt:** "Tests fail on Polars datetime parsing with ISO strings ending in Z. Fix for current Polars."  
**Result:** Added explicit datetime format `%Y-%m-%dT%H:%M:%S%#z`.

### Prompt 3
**Tool:** Codex GPT-5  
**Prompt:** "Add social indicators similar to SMA/RSI for dashboard use."  
**Result:** Added rolling sentiment SMA and `social_rsi` based on positive/negative post counts.

### Summary
- Prompt count: 3
- Manual fixes: adjusted tests to assert behavior instead of exact internal dtypes.
- Time: about 45 minutes

---

## Task 4. Dashboard and Docker

### Prompt 1
**Tool:** Codex GPT-5  
**Prompt:** "Create Streamlit dashboard for topic selection, sentiment trend, engagement chart and current table."  
**Result:** Added dashboard reading either Parquet or Arrow endpoint.

### Prompt 2
**Tool:** Codex GPT-5  
**Prompt:** "Package collector and dashboard in Docker Compose with healthchecks."  
**Result:** Added collector and dashboard Dockerfiles, Compose stack and `/health` endpoint.

### Summary
- Prompt count: 2
- Manual fixes: dashboard catches Arrow request failures and falls back to Parquet when configured.
- Time: about 35 minutes

---

## Task 5. Performance comparison and CI

### Prompt 1
**Tool:** Codex GPT-5  
**Prompt:** "Add performance comparison: Go collector vs Python asyncio collector under the same synthetic load."  
**Result:** Added `pipeline/python_collector.py` and `pipeline/benchmark.py`.

### Prompt 2
**Tool:** Codex GPT-5  
**Prompt:** "Add CI that runs Go tests, Python tests and a smoke ETL pipeline."  
**Result:** Added `.github/workflows/ci.yml`.

### Summary
- Prompt count: 2
- Manual fixes: generated benchmark outputs are ignored by git.
- Time: about 30 minutes

---

## Final reviewer pass

### Prompt 1
**Tool:** Codex GPT-5  
**Prompt:** "Review lab14 like a university AI reviewer: check artifacts, go.mod, tests, prompt log, README, Docker and production-like engineering."  
**Result:** Ensured `.gitignore`, `go.mod`, pyproject/requirements, Go tests, Python tests, Docker Compose and detailed docs are present.

### Summary
- Prompt count: 1
- Manual fixes: preserved all generated data under ignored `data/` and `reports/`.
- Time: about 20 minutes
