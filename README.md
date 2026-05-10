# Лабораторная работа №14 - конвейер ETL для мониторинга соцсетей

## Автор

- ФИО: Савенков Денис Юрьевич
- Группа: 221331
- Вариант: 12
- Предметная область: мониторинг социальных сетей, эмуляция Twitter/X API
- Уровень: повышенный

## Архитектура

```text
Go collector
  -> синтетический поток постов в стиле Twitter/X
  -> пакетная запись JSONL
  -> tumbling-window агрегация
  -> HTTP endpoint с Apache Arrow

Python analytics
  -> очистка и валидация через Polars
  -> сохранение в Parquet
  -> SQL-отчёт через DuckDB
  -> HTML-графики Plotly
  -> Streamlit dashboard
```

## Что реализовано

| Требование | Реализация |
| --- | --- |
| Сбор данных на Go | Детерминированный эмулятор Twitter/X с темами, sentiment, engagement и author id |
| Буферизация и пакетная запись | Generic JSONL writer с flush по размеру batch и времени |
| Graceful shutdown | `signal.NotifyContext`, финальный flush буферов и корректное завершение HTTP server |
| Оконная агрегация | Tumbling windows по topic: count, min/max/avg sentiment, engagement, unique authors |
| Apache Arrow | Go endpoint `/arrow` отдаёт Arrow IPC stream |
| Python-анализ | Polars cleanup, индикаторы, Parquet output и DuckDB report |
| Визуализация | Plotly HTML-отчёты и Streamlit dashboard |
| Производительность | `pipeline.benchmark` сравнивает Go collector и Python asyncio collector |
| Тесты | Go unit tests и Python pytest tests |

## Установка

```powershell
python -m venv .venv
.\.venv\Scripts\python.exe -m pip install -U pip
.\.venv\Scripts\python.exe -m pip install -r requirements.txt
```

## Генерация данных

```powershell
cd collector
go run ./cmd/collector -count 360 -out ../data -batch 60 -window 1m
```

Будут созданы файлы:

- `data/posts.jsonl`
- `data/windows.jsonl`

Эти файлы являются артефактами выполнения и не коммитятся в Git.

## Анализ данных

```powershell
.\.venv\Scripts\python.exe -m pipeline.analyzer --input data/windows.jsonl --out reports
```

Результаты:

- `reports/social_windows.parquet`
- `reports/topic_summary.csv`
- `reports/sentiment_trend.html`
- `reports/engagement.html`

Папка `reports/` тоже игнорируется Git, потому что содержит сгенерированные отчёты.

## Apache Arrow server

```powershell
cd collector
go run ./cmd/collector -count 360 -out ../data -batch 60 -serve -addr :8080
```

Endpoint:

- `GET /health`
- `GET /arrow`
- `GET /arrow?topic=ai`

## Dashboard

Локально по Parquet:

```powershell
$env:PARQUET_PATH="reports/social_windows.parquet"
.\.venv\Scripts\streamlit.exe run dashboard/app.py
```

Через Docker Compose:

```powershell
docker compose up --build
```

Dashboard: `http://127.0.0.1:8501`.

## Тесты

Go:

```powershell
cd collector
go test ./...
```

Python:

```powershell
.\.venv\Scripts\python.exe -m pytest -q
```

## Проверка производительности

```powershell
.\.venv\Scripts\python.exe -m pipeline.benchmark --count 1000 --out reports/performance.json
```

Benchmark измеряет время выполнения и throughput для Go collector и Python asyncio emulator. Файл `reports/performance.json` не хранится в Git, потому что это сгенерированное доказательство запуска, а не исходный код.

## Docker

```powershell
docker compose config
docker compose build
docker compose up -d
```

В Compose есть:

- `collector` - Go service, генерирует данные, агрегирует окна и отдаёт Arrow;
- `dashboard` - Streamlit UI, читает Arrow endpoint;
- healthcheck `collector` через `/health`.

## CI

`.github/workflows/ci.yml` запускает:

- `go test ./...`;
- `pytest -q`;
- smoke pipeline: Go collector генерирует данные, Python analyzer строит отчёт.

## Структура проекта

```text
collector/          Go collector, aggregator, Arrow endpoint и Go tests
pipeline/           Python analytics, Arrow client и benchmark
dashboard/          Streamlit monitoring UI
tests/              Python tests
.github/workflows/  CI для Go, Python и smoke pipeline
docker-compose.yml  Collector + dashboard stack
PROMPT_LOG.md       Реальный лог работы с AI
```
