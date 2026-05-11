from __future__ import annotations

import json

import pytest

from pipeline.stream_consumer import decode_window_message


def test_decode_window_message_accepts_valid_window() -> None:
    payload = {
        "window_start": "2026-05-10T12:00:00Z",
        "window_end": "2026-05-10T12:01:00Z",
        "topic": "ai",
        "post_count": 5,
        "avg_sentiment": 0.25,
        "total_engagement": 100,
    }

    assert decode_window_message(json.dumps(payload).encode("utf-8")) == payload


def test_decode_window_message_rejects_missing_fields() -> None:
    with pytest.raises(ValueError, match="missing fields"):
        decode_window_message(b'{"topic":"ai"}')
