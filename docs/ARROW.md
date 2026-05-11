# Apache Arrow в ЛР14

Arrow используется как бинарный формат обмена между Go-collector и Python/Streamlit dashboard.

## Зачем это нужно

- JSONL остаётся долговечным audit-log для batch и stream replay.
- Arrow stream отдаёт агрегированные окна без лишней сериализации в CSV/JSON.
- Python-клиент сразу превращает Arrow IPC stream в Polars DataFrame.

## Контракт

Collector публикует HTTP endpoint:

```text
GET /arrow?topic=ai
Content-Type: application/vnd.apache.arrow.stream
```

Клиентская функция `pipeline.arrow_client.read_arrow_stream()` проверяет пустой ответ и читает IPC stream через `pyarrow.ipc.open_stream`.

## Проверка

```bash
python -m pytest tests/test_arrow_client.py -q
```
