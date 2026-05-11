from __future__ import annotations

import argparse
import asyncio
import json
import logging
from pathlib import Path
from typing import Any

import nats

from pipeline.analyzer import run

LOGGER = logging.getLogger("lab14.stream-consumer")


def decode_window_message(payload: bytes) -> dict[str, Any]:
    message = json.loads(payload.decode("utf-8"))
    required = {"window_start", "window_end", "topic", "post_count", "avg_sentiment", "total_engagement"}
    missing = required - set(message)
    if missing:
        raise ValueError(f"NATS message is missing fields: {sorted(missing)}")
    return message


async def consume_windows(nats_url: str, subject: str, output: Path, report_dir: Path, flush_every: int) -> None:
    output.parent.mkdir(parents=True, exist_ok=True)
    report_dir.mkdir(parents=True, exist_ok=True)
    processed = 0

    async def handle_message(message: Any) -> None:
        nonlocal processed
        try:
            window = decode_window_message(message.data)
        except (json.JSONDecodeError, UnicodeDecodeError, ValueError) as exc:
            LOGGER.warning("skip invalid stream message: %s", exc)
            return

        with output.open("a", encoding="utf-8") as file:
            file.write(json.dumps(window, ensure_ascii=False) + "\n")
        processed += 1
        if processed % flush_every == 0:
            LOGGER.info("running incremental analysis after %s streamed windows", processed)
            run(output, report_dir)

    while True:
        try:
            client = await nats.connect(nats_url, name="lab14-stream-analyzer", connect_timeout=5)
            break
        except OSError as exc:
            LOGGER.warning("NATS is unavailable (%s), retrying in 3 seconds", exc)
            await asyncio.sleep(3)

    await client.subscribe(subject, cb=handle_message)
    LOGGER.info("listening on %s via %s", subject, nats_url)
    try:
        while True:
            await asyncio.sleep(3600)
    finally:
        await client.drain()


def main() -> None:
    parser = argparse.ArgumentParser(description="Consume social window metrics from NATS")
    parser.add_argument("--nats-url", default="nats://localhost:4222")
    parser.add_argument("--subject", default="social.windows")
    parser.add_argument("--output", type=Path, default=Path("data/stream_windows.jsonl"))
    parser.add_argument("--reports", type=Path, default=Path("reports"))
    parser.add_argument("--flush-every", type=int, default=10)
    args = parser.parse_args()

    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    asyncio.run(consume_windows(args.nats_url, args.subject, args.output, args.reports, args.flush_every))


if __name__ == "__main__":
    main()
