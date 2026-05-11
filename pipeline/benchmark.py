from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import time
import tracemalloc
from pathlib import Path

from pipeline.python_collector import collect_posts


def benchmark_python(count: int) -> dict[str, float]:
    import asyncio

    tracemalloc.start()
    started = time.perf_counter()
    posts = asyncio.run(collect_posts(count=count))
    elapsed = time.perf_counter() - started
    _, peak = tracemalloc.get_traced_memory()
    tracemalloc.stop()
    return {
        "records": float(len(posts)),
        "seconds": round(elapsed, 4),
        "records_per_second": round(len(posts) / elapsed, 2),
        "peak_mb": round(peak / 1024 / 1024, 3),
    }


def benchmark_go(count: int, workdir: Path) -> dict[str, float]:
    out_dir = workdir / "data" / "bench-go"
    if out_dir.exists():
        shutil.rmtree(out_dir)
    started = time.perf_counter()
    subprocess.run(
        ["go", "run", "./cmd/collector", "-count", str(count), "-out", str(out_dir), "-batch", "100"],
        cwd=workdir / "collector",
        check=True,
        capture_output=True,
        text=True,
    )
    elapsed = time.perf_counter() - started
    rows = sum(1 for _ in (out_dir / "posts.jsonl").open(encoding="utf-8"))
    size_mb = (out_dir / "posts.jsonl").stat().st_size / 1024 / 1024
    return {
        "records": float(rows),
        "seconds": round(elapsed, 4),
        "records_per_second": round(rows / elapsed, 2),
        "output_mb": round(size_mb, 3),
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Compare Go and Python social collectors")
    parser.add_argument("--count", type=int, default=1000)
    parser.add_argument("--out", type=Path, default=Path("reports/performance.json"))
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    result = {"go": benchmark_go(args.count, root), "python": benchmark_python(args.count)}
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(result, indent=2), encoding="utf-8")
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
