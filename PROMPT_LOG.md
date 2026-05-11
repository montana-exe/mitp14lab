# Prompt Log

## Advanced 1. Go collectors и предметная область

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай ЛР14, вариант 12: распределённая платформа мониторинга соцсетей, Go collector, Python analytics, без реальных API-ключей."  
**Результат:** Получил deterministic simulator Twitter/X-подобных постов: topics, authors, sentiment, engagement, language.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Вынеси core logic из CLI, чтобы были нормальные Go unit tests и graceful shutdown."  
**Результат:** Добавлены `collector/internal/social`, JSONL batch writer, aggregator tests.  
**Что исправлено вручную:** CLI не должен держать бизнес-логику, иначе тесты запускали бы весь collector.

### Итого
- Количество промптов: 2
- Проблемы: сначала код был слишком CLI-oriented.
- Время: ~45 минут

---

## Advanced 2. Distributed coordination через etcd

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь etcd registry для collectors: collector-id, shard-index, shard-total, HTTP v3 API без тяжёлой зависимости."  
**Результат:** Collector пишет ключ `/lab14/collectors/<id>` в etcd.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай shard filter, чтобы несколько collectors не обрабатывали один и тот же поток."  
**Результат:** Добавлен FNV-1a hash по `post_id`.  
**Что исправлено вручную:** Для Kubernetes добавлен режим `-shard-index -1`, чтобы индекс выводился из имени pod, иначе HPA-копии получали одинаковый shard.

### Итого
- Количество промптов: 2
- Проблемы: Docker networking для etcd требовал endpoint `http://etcd:2379`, а не localhost.
- Время: ~40 минут

---

## Advanced 3. NATS streaming pipeline

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь настоящий stream broker между collectors и analyzer. Kafka можно, но для учебного compose лучше NATS."  
**Результат:** Collector публикует `WindowMetric` в subject `social.windows`.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай Python analyzer service, который подписывается на NATS и пишет stream JSONL."  
**Результат:** Добавлен `pipeline.stream_consumer`, Dockerfile analyzer и tests для message validation.

### Промпт 3
**Инструмент:** Codex GPT-5  
**Промпт:** "NATS в compose стартует позже analyzer, добавь reconnect/retry."  
**Результат:** Analyzer повторяет подключение и не падает при раннем старте.  
**Что исправлено вручную:** Некорректные NATS messages логируются и пропускаются, чтобы один bad event не ломал consumer.

### Итого
- Количество промптов: 3
- Проблемы: connection refused при холодном старте compose.
- Время: ~55 минут

---

## Advanced 4. Apache Arrow exchange

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай Arrow endpoint в Go и Python client для Streamlit dashboard."  
**Результат:** Добавлен `/arrow`, `pipeline.arrow_client.fetch_arrow`, dashboard читает Arrow IPC stream.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Arrow response иногда пустой при ошибке upstream. Сделай тестируемую защиту."  
**Результат:** Добавлена `read_arrow_stream()` и tests.

### Итого
- Количество промптов: 2
- Проблемы: Arrow serialization чувствителен к типам timestamp, поэтому timestamps передаются строкой.
- Время: ~35 минут

---

## Advanced 5. Python analytics: Polars + DuckDB

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Реализуй analyzer: JSONL -> Polars cleanup -> indicators -> Parquet -> DuckDB report -> Plotly HTML."  
**Результат:** Добавлены cleanup, `sentiment_sma`, `social_rsi`, topic summary и HTML-графики.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Pytest падает на datetime parsing Polars для ISO Z. Исправь совместимо с текущей версией."  
**Результат:** Использован явный формат `%Y-%m-%dT%H:%M:%S%#z`.

### Промпт 3
**Инструмент:** Codex GPT-5  
**Промпт:** "DuckDB отчёт должен быть параметризованным, без string SQL путей."  
**Результат:** `read_parquet(?)` вызывается с параметром.  
**Что исправлено вручную:** Тесты проверяют outcome, а не хрупкое внутреннее представление Polars dtype.

### Итого
- Количество промптов: 3
- Проблемы: Polars/DuckDB различались в обработке datetime и nullable columns.
- Время: ~60 минут

---

## Advanced 6. Rust validation module

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь Rust-библиотеку для валидации social posts/window metrics JSONL."  
**Результат:** Создан `rust-validator` crate: lib, CLI `validate`, CLI `bench`, unit tests.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Подключи Rust validator к Python analyzer, но не ломай окружение без Rust."  
**Результат:** Добавлена `pipeline.rust_validator`; analyzer включает проверку через `--rust-validate`.

### Промпт 3
**Инструмент:** Codex GPT-5  
**Промпт:** "На Windows cargo не найден. Сделай честный benchmark вместо фейкового результата."  
**Результат:** Benchmark пишет `rust_validator.status = unavailable`, если Cargo отсутствует.  
**Что исправлено вручную:** Локально `cargo test` не запускался, потому что Rust toolchain не установлен.

### Итого
- Количество промптов: 3
- Проблемы: Rust build нельзя было проверить на текущей машине без установки toolchain.
- Время: ~50 минут

---

## Advanced 7. Docker Compose production stack

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Расширь docker-compose до distributed stack: etcd, NATS, collector, analyzer, dashboard."  
**Результат:** Добавлены сервисы, volumes, healthchecks и `depends_on: condition: service_healthy`.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Проверь compose config и исправь networking между services."  
**Результат:** `docker compose config` проходит; service names используются как DNS names.

### Итого
- Количество промптов: 2
- Проблемы: Docker Desktop на локальной машине был выключен, поэтому полный `compose up` нельзя было подтвердить.
- Время: ~30 минут

---

## Advanced 8. Kubernetes + HPA + финальная проверка

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь Kubernetes manifests: deployments, services, resource limits, HPA."  
**Результат:** Создан `k8s/` с namespace, etcd, NATS, collector, analyzer, dashboard и HPA.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "README должен выглядеть как distributed-system документация, а prompt log - как реальный процесс, не 1 идеальный промпт."  
**Результат:** README переписан на русском: architecture diagram, data flow, streaming, Arrow, Kubernetes, scaling, tests.

### Промпт 3
**Инструмент:** Codex GPT-5  
**Промпт:** "Проверь работу как AI reviewer: тесты, git artifacts, README, PROMPT_LOG, Docker config."  
**Результат:** Запущены Go tests, pytest, benchmark, docker compose config; проверено отсутствие tracked artifacts.

### Итого
- Количество промптов: 3
- Проблемы: Kubernetes HPA требует metrics-server в кластере, это описано в docs.
- Время: ~45 минут
