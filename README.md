# Лабораторная работа №14 - distributed streaming analytics platform

## Автор

- ФИО: Савенков Денис Юрьевич
- Группа: 221331
- Вариант: 12
- Предметная область: мониторинг социальных сетей, эмуляция Twitter/X-потока
- Уровень: повышенный, 8/8 advanced tasks

## Идея проекта

Проект моделирует распределённую платформу аналитики соцсетей: Go-collectors генерируют поток постов, делят нагрузку по шардам, регистрируются в etcd, публикуют оконные метрики в NATS, Python analyzer строит Parquet/DuckDB/Plotly-отчёты, Streamlit dashboard читает Apache Arrow endpoint, а Rust validator проверяет JSONL records перед аналитикой.

```text
                +----------------+
                |      etcd      |
                | collector registry
                +--------^-------+
                         |
+-------------+   shard ownership   +------------------+
| collector-1 |-------------------->|                  |
| collector-N |                     |   NATS broker    |
+------+------+                     |  social.windows  |
       |                            +---------+--------+
       | JSONL posts/windows                  |
       v                                      v
+-------------+                      +------------------+
| Arrow HTTP  |<---------------------| stream analyzer  |
| /arrow      |                      | Polars + DuckDB  |
+------+------+                      +---------+--------+
       |                                       |
       v                                       v
+-------------+                      +------------------+
| Streamlit   |                      | Parquet / CSV /  |
| dashboard   |                      | Plotly reports   |
+-------------+                      +------------------+

Rust validator validates JSONL posts/windows before strict analysis or benchmark.
Kubernetes manifests deploy collector/analyzer/dashboard plus etcd, NATS and HPA.
```

## Что закрыто по advanced-критериям

| Задание | Реализация |
| --- | --- |
| Distributed collectors | Go collector поддерживает `collector-id`, `shard-index`, `shard-total`, `shard-strategy`, детерминированный shard assignment и регистрацию в etcd через v3 HTTP API. |
| Streaming broker | NATS JetStream включён в Docker Compose; collector публикует `social.windows`, analyzer подписывается и пишет stream JSONL. |
| Fast analytics | Polars очищает окна, DuckDB строит SQL-отчёт, результат сохраняется в Parquet/CSV/HTML. |
| Apache Arrow | Go endpoint `/arrow` отдаёт Arrow IPC stream; Python/Streamlit читает его через `pyarrow`. |
| Rust validation | `rust-validator/` содержит Rust-библиотеку, CLI, unit-тесты и benchmark-команду для JSONL records. |
| Dashboard | Streamlit dashboard показывает sentiment trend, engagement и таблицу окон. |
| Kubernetes/HPA | `k8s/` содержит namespace, deployments, services, resource requests/limits и `autoscaling/v2` HPA. |
| Production hygiene | `.gitignore`, go.mod, tests, Docker healthchecks, PROMPT_LOG, docs, отсутствие артефактов в Git. |

## Запуск локально

```powershell
python -m venv .venv
.\.venv\Scripts\python.exe -m pip install -U pip
.\.venv\Scripts\python.exe -m pip install -r requirements.txt
```

Сгенерировать данные:

```powershell
cd collector
go run ./cmd/collector -count 360 -out ../data -batch 60 -window 1m
cd ..
```

Построить отчёты:

```powershell
.\.venv\Scripts\python.exe -m pipeline.analyzer --input data/windows.jsonl --out reports
```

При установленном Rust toolchain можно включить строгую Rust-валидацию:

```powershell
.\.venv\Scripts\python.exe -m pipeline.analyzer --input data/windows.jsonl --out reports --rust-validate
```

## Docker Compose

```powershell
docker compose config
docker compose up --build
```

Сервисы:

- `etcd` - registry collectors и coordination metadata.
- `nats` - streaming broker для `social.windows`.
- `collector` - Go service, пишет JSONL, отдаёт Arrow и публикует в NATS.
- `analyzer` - Python stream consumer, создаёт Parquet/CSV/HTML отчёты.
- `dashboard` - Streamlit UI, читает Arrow endpoint.

Healthchecks есть у `etcd`, `nats` и `collector`; зависимости в Compose используют `condition: service_healthy`.

## Kubernetes

Перед применением нужно собрать локальные образы:

```powershell
docker build -t lab14-social-collector:latest ./collector
docker build -t lab14-stream-analyzer:latest -f analyzer/Dockerfile .
docker build -t lab14-social-dashboard:latest -f dashboard/Dockerfile .
kubectl apply -k k8s
kubectl -n lab14-social get pods
kubectl -n lab14-social get hpa
```

HPA масштабирует `collector` от 2 до 6 реплик по CPU. Для демонстрации shard ownership в Kubernetes используется `-shard-index -1`: collector сам выводит индекс шарда из имени pod. Поддерживаются стратегии `hash`, `topic` и `author-range`, поэтому нагрузку можно делить по источникам, темам или диапазонам авторов.

## Apache Arrow

```powershell
cd collector
go run ./cmd/collector -count 360 -out ../data -batch 60 -serve -addr :8080
```

Endpoints:

- `GET /health`
- `GET /arrow`
- `GET /arrow?topic=ai`

Arrow нужен для быстрой передачи агрегированных окон из Go в Python/Streamlit без промежуточного CSV. Подробнее: [docs/ARROW.md](docs/ARROW.md).

## Benchmark

```powershell
.\.venv\Scripts\python.exe -m pipeline.benchmark --count 1000 --out reports/performance.json
```

Benchmark сравнивает Go collector, Python asyncio collector и Rust validator. Если `cargo` не установлен, Rust-часть честно помечается как `unavailable`, без фейковых результатов. Подробнее: [docs/BENCHMARK.md](docs/BENCHMARK.md).

## Тесты

```powershell
cd collector
go test ./...
cd ..
.\.venv\Scripts\python.exe -m pytest -q
```

Если установлен Rust:

```powershell
cd rust-validator
cargo test
```

## Структура

```text
collector/          Go collector, sharding, etcd registration, NATS publishing, Arrow endpoint
pipeline/           Python analyzer, NATS consumer, Arrow client, benchmark, Rust wrapper
rust-validator/     Rust JSONL validation library and CLI
dashboard/          Streamlit dashboard
analyzer/           Dockerfile for stream analyzer service
k8s/                Kubernetes deployment.yaml, service.yaml, hpa.yaml and infra manifests
docs/               Architecture notes for Arrow, streaming, coordination and benchmark
tests/              Python unit tests
docker-compose.yml  Full distributed local stack
PROMPT_LOG.md       Реалистичный лог работы с AI по 8 advanced-задачам
```
