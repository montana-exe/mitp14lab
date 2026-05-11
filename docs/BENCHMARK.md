# Benchmark

Benchmark сравнивает три слоя платформы:

- Go collector: генерация и JSONL batching.
- Python collector: asyncio-симулятор для baseline.
- Rust validator: проверка JSONL событий, если установлен `cargo`.

Команда:

```bash
python -m pipeline.benchmark --count 1000 --out reports/performance.json
```

Если Rust toolchain отсутствует, отчёт не подменяет результат фейковым числом, а пишет:

```json
{
  "rust_validator": {
    "status": "unavailable",
    "reason": "cargo is not installed"
  }
}
```

Такой формат выбран специально, чтобы benchmark оставался честным и воспроизводимым на машине проверяющего.
