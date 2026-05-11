from __future__ import annotations

from io import BytesIO

import polars as pl
import pyarrow as pa
import pyarrow.ipc as ipc
import pytest

from pipeline.arrow_client import read_arrow_stream


def test_read_arrow_stream_roundtrip() -> None:
    table = pa.table({"topic": ["ai"], "post_count": [3]})
    sink = BytesIO()
    with ipc.new_stream(sink, table.schema) as writer:
        writer.write_table(table)

    result = read_arrow_stream(sink.getvalue())

    assert result.to_dicts() == [{"topic": "ai", "post_count": 3}]


def test_read_arrow_stream_rejects_empty_response() -> None:
    with pytest.raises(ValueError, match="empty"):
        read_arrow_stream(b"")
