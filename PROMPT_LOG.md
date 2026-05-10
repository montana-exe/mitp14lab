# Prompt Log

## Задание 1. Go collector для мониторинга соцсетей

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Реализуй лабораторную 14, вариант 12: ETL-конвейер для мониторинга Twitter/X-подобных соцсетей. Go использовать для сбора, Python для аналитики."  
**Результат:** Сформирована архитектура: синтетический social stream вместо реального Twitter/X API, потому что публичные API-ключи недоступны в окружении проверяющего.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай Go collector production-like: buffered writes, graceful shutdown, testable package и никаких сгенерированных данных в Git."  
**Результат:** Добавлены deterministic simulator, generic JSONL batch writer, signal handling и Go unit tests.

### Промпт 3
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь tumbling-window aggregation на стороне Go до передачи данных в Python."  
**Результат:** Добавлены оконные метрики по topic: post count, positive/negative/neutral count, min/max/avg sentiment, engagement, unique authors.

### Итого
- Количество промптов: 3
- Что исправлено вручную: core logic вынесена в `collector/internal/social`, чтобы тесты не запускали CLI.
- Время: примерно 55 минут

---

## Задание 2. Передача данных через Apache Arrow

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Открой агрегированные оконные метрики через Apache Arrow endpoint из Go."  
**Результат:** Добавлена зависимость `github.com/apache/arrow/go/v17`, Arrow schema и endpoint `/arrow`.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай Arrow schema простой и стабильной для Python client."  
**Результат:** Timestamp передаётся строкой, остальные поля примитивные numeric/string. Это снижает риск timezone несовместимости между Go Arrow и Polars.

### Итого
- Количество промптов: 2
- Что исправлено вручную: выполнен `go mod tidy`, затем проверено `go test ./...`.
- Время: примерно 30 минут

---

## Задание 3. Python-анализ через Polars и DuckDB

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Собери Python analyzer: загрузка JSONL, очистка через Polars, сохранение Parquet, DuckDB SQL и графики."  
**Результат:** Добавлены `pipeline/analyzer.py`, зависимости и pytest tests.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Тесты падают на Polars datetime parsing для ISO строк с Z. Исправь под текущую версию Polars."  
**Результат:** Добавлен явный формат datetime `%Y-%m-%dT%H:%M:%S%#z`.

### Промпт 3
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь social indicators, похожие на SMA/RSI, чтобы dashboard был полезнее."  
**Результат:** Добавлены rolling sentiment SMA и `social_rsi` на основе positive/negative counts.

### Итого
- Количество промптов: 3
- Что исправлено вручную: тесты проверяют поведение, а не хрупкие внутренние dtype-представления.
- Время: примерно 45 минут

---

## Задание 4. Dashboard и Docker

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Создай Streamlit dashboard: выбор topic, sentiment trend, engagement chart и таблица текущих окон."  
**Результат:** Добавлен dashboard, который читает либо Parquet, либо Arrow endpoint.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Упакуй collector и dashboard в Docker Compose с healthchecks."  
**Результат:** Добавлены Dockerfile для collector и dashboard, Compose stack и endpoint `/health`.

### Итого
- Количество промптов: 2
- Что исправлено вручную: dashboard ловит ошибки Arrow request и может работать через Parquet при локальном запуске.
- Время: примерно 35 минут

---

## Задание 5. Производительность и CI

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь сравнение производительности: Go collector против Python asyncio collector при одинаковой синтетической нагрузке."  
**Результат:** Добавлены `pipeline/python_collector.py` и `pipeline/benchmark.py`.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь CI, который запускает Go tests, Python tests и smoke ETL pipeline."  
**Результат:** Добавлен `.github/workflows/ci.yml`.

### Промпт 3
**Инструмент:** Codex GPT-5  
**Промпт:** "Benchmark повторно считает старые Go records, если data/bench-go уже существует."  
**Результат:** Перед Go benchmark папка `data/bench-go` очищается, чтобы результат был воспроизводимым.

### Итого
- Количество промптов: 3
- Что исправлено вручную: generated benchmark outputs остаются в ignored `reports/`.
- Время: примерно 35 минут

---

## Финальная ревизия

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Проверь lab14 как университетский AI reviewer: artifacts, go.mod, tests, prompt log, README, Docker и production-like engineering."  
**Результат:** Проверено наличие `.gitignore`, `go.mod`, pyproject/requirements, Go tests, Python tests, Docker Compose и подробной документации.

### Итого
- Количество промптов: 1
- Что исправлено вручную: все сгенерированные данные находятся в ignored `data/` и `reports/`; README и PROMPT_LOG переведены на русский.
- Время: примерно 20 минут
