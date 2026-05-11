FROM python:3.12-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY pipeline ./pipeline
CMD ["python", "-m", "pipeline.stream_consumer", "--nats-url", "nats://nats:4222", "--subject", "social.windows", "--output", "/app/data/stream_windows.jsonl", "--reports", "/app/reports", "--flush-every", "5"]
