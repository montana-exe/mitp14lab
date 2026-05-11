from __future__ import annotations

import shutil
import subprocess
from pathlib import Path


class RustValidatorUnavailable(RuntimeError):
    pass


def validate_jsonl(path: Path, manifest_path: Path | None = None) -> None:
    if not path.exists():
        raise FileNotFoundError(path)
    cargo = shutil.which("cargo")
    if cargo is None:
        raise RustValidatorUnavailable("cargo is not installed; Rust validation cannot run")

    repo_root = Path(__file__).resolve().parents[1]
    manifest = manifest_path or repo_root / "rust-validator" / "Cargo.toml"
    result = subprocess.run(
        [cargo, "run", "--quiet", "--manifest-path", str(manifest), "--", "validate", str(path)],
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise ValueError(result.stderr.strip() or result.stdout.strip())
