from __future__ import annotations

from io import BytesIO

import polars as pl
import pyarrow.ipc as ipc
import requests


def fetch_arrow(url: str, topic: str | None = None, timeout: float = 5.0) -> pl.DataFrame:
    params = {"topic": topic} if topic else None
    response = requests.get(url, params=params, timeout=timeout)
    response.raise_for_status()
    reader = ipc.open_stream(BytesIO(response.content))
    table = reader.read_all()
    return pl.from_arrow(table)
