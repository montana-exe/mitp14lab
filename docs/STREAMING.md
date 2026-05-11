# Streaming pipeline

В Docker Compose используется NATS с включённым JetStream monitoring port.

```text
collector -> NATS subject social.windows -> stream analyzer -> reports/
```

Collector публикует каждую `WindowMetric` как JSON message. Analyzer подписывается на subject, валидирует обязательные поля сообщения и append-only пишет `data/stream_windows.jsonl`. После каждых N сообщений он запускает Polars/DuckDB analysis.

## Ошибки и устойчивость

- Если NATS ещё не готов, analyzer повторяет подключение.
- Некорректные сообщения не ломают consumer, а логируются и пропускаются.
- JSONL остаётся replay-log: поток можно переиграть в batch analyzer.
