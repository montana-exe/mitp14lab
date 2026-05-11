# Prompt Log

## Задание 1 - etcd collector

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь distributed coordination: collectors должны регистрироваться в etcd и хранить collector-id, shard-index, shard-total."  
**Результат:** Collector пишет ключ `/lab14/collectors/<collector-id>` через etcd v3 HTTP API.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Нужно shard assignment не только по post_id, но и по темам/авторам, чтобы было похоже на реальную распределённую систему."  
**Результат:** Добавлены стратегии `hash`, `topic`, `author-range`, режим `-shard-index -1` для Kubernetes pod names и Go unit tests.

### Итого
- Количество промптов: 2
- Проблемы: Docker networking для etcd требовал service DNS `http://etcd:2379`, а не localhost.
- Что исправлено вручную: добавлена сериализация metadata в etcd value и тесты shard assignment.

---

## Задание 2 - window aggregation

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай Go aggregation по tumbling windows для social monitoring: counts, sentiment, engagement, unique authors."  
**Результат:** Добавлен агрегатор окон по topic с метриками positive/negative/neutral, min/max/avg sentiment, engagement и unique authors.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Вынеси collector logic из CLI, чтобы можно было тестировать aggregator и graceful flush отдельно."  
**Результат:** Логика вынесена в `collector/internal/social`, добавлены Go unit tests.

### Итого
- Количество промптов: 2
- Проблемы: первоначальный CLI-вариант было сложно покрыть тестами.
- Что исправлено вручную: writer получил batch flush по размеру и времени.

---

## Задание 3 - Arrow

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Открой агрегированные окна через Apache Arrow endpoint из Go."  
**Результат:** Добавлен `/arrow`, Arrow IPC stream и Python client.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Arrow response иногда может быть пустым при upstream ошибке. Добавь тестируемую защиту."  
**Результат:** Выделена `read_arrow_stream()`, добавлены pytest tests.

### Итого
- Количество промптов: 2
- Проблемы: Arrow serialization чувствителен к timestamp/timezone, поэтому timestamp передаётся строкой.
- Что исправлено вручную: добавлен отдельный документ `docs/ARROW.md`.

---

## Задание 4 - Rust validation

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь Rust validation module для social records: topic, sentiment range, engagement, timestamp."  
**Результат:** Создан `rust-validator` crate с библиотекой, CLI `validate`, CLI `bench` и Rust unit tests.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Подключи Rust validator к Python analyzer, но не ломай окружение без cargo."  
**Результат:** Добавлена Python-обёртка `pipeline.rust_validator` и флаг `--rust-validate`.

### Итого
- Количество промптов: 2
- Проблемы: локально `cargo` не установлен, поэтому Python-тесты покрывают unavailable/error paths.
- Что исправлено вручную: benchmark честно пишет `rust_validator.status = unavailable`, а не имитирует Rust-результат.

---

## Задание 5 - Kubernetes/HPA

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь Kubernetes manifests: deployment.yaml, service.yaml, hpa.yaml, resource requests/limits."  
**Результат:** Созданы `k8s/deployment.yaml`, `k8s/service.yaml`, `k8s/hpa.yaml`, namespace, etcd и NATS manifests.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Проверь kustomize output и убери дублирующиеся deployment файлы."  
**Результат:** `kubectl kustomize k8s` строит единый валидный manifest set.

### Итого
- Количество промптов: 2
- Проблемы: HPA требует metrics-server в реальном кластере.
- Что исправлено вручную: split manifests свернуты в ожидаемые `deployment.yaml` и `service.yaml`.

---

## Задание 6 - Go vs Python benchmark

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь честный benchmark Go collector против Python asyncio collector."  
**Результат:** `pipeline.benchmark` запускает реальный `go run`, считает records/sec и сравнивает с Python collector.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь Rust validator benchmark, но не фейкуй его без cargo."  
**Результат:** Rust benchmark запускается только при доступном `cargo`, иначе помечается как unavailable.

### Итого
- Количество промптов: 2
- Проблемы: старая папка benchmark могла влиять на количество строк.
- Что исправлено вручную: `data/bench-go` очищается перед каждым запуском.

---

## Задание 7 - Kafka/NATS streaming

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Добавь настоящий stream broker: collector публикует события/агрегаты, Python analyzer читает поток. JSONL оставить fallback."  
**Результат:** Выбран NATS, collector публикует `social.windows`, analyzer подписывается через `pipeline.stream_consumer`.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "NATS может стартовать позже analyzer. Добавь reconnect и validation сообщений."  
**Результат:** Analyzer повторяет подключение, invalid messages логируются и пропускаются.

### Итого
- Количество промптов: 2
- Проблемы: `connect refused` при холодном старте Docker Compose.
- Что исправлено вручную: JSONL оставлен как replay/fallback, но основной compose path идёт через NATS.

---

## Задание 8 - realtime dashboard

### Промпт 1
**Инструмент:** Codex GPT-5  
**Промпт:** "Сделай realtime dashboard для social monitoring: topic filter, sentiment trend, engagement, table."  
**Результат:** Streamlit dashboard читает Arrow endpoint и показывает ключевые графики.

### Промпт 2
**Инструмент:** Codex GPT-5  
**Промпт:** "Dashboard не должен быть единственным сервисом в compose, добавь analyzer, healthchecks и depends_on."  
**Результат:** Compose содержит `etcd`, `nats`, `collector`, `analyzer`, `dashboard`; healthchecks и `condition: service_healthy`.

### Итого
- Количество промптов: 2
- Проблемы: Docker Desktop локально выключен, поэтому runtime compose проверялся через `docker compose config`.
- Что исправлено вручную: README описывает dashboard, Arrow path, streaming path и kubectl команды.
