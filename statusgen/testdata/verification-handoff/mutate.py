#!/usr/bin/env python3
"""Run bounded, offline production mutations from the accompanying manifest."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[3]
MANIFEST = json.loads((Path(__file__).with_name("01-mutations.json")).read_text())
SNAPSHOT = ["deskcore", "statusgen", "tools/desk", "comms", "forgeread", "gitquiet", "perfrec", "topology.yaml", ".assay-surfaces", "spec/intake-exit-v1.md"]


def run(case):
    # A bounded tracked snapshot plus explicitly enumerated not-yet-committed inputs.
    # Never copy a directory tree containing build output or module caches.
    raw = subprocess.check_output(["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", *SNAPSHOT], cwd=ROOT)
    files = sorted(set(p for p in raw.decode().split("\0") if p))
    if len(files) > 16000 or sum((ROOT / p).stat().st_size for p in files) > 96 * 1024 * 1024:
        raise RuntimeError("snapshot exceeds declared 16000-file/96-MiB budget")
    with tempfile.TemporaryDirectory(prefix="handoff-mutation-") as temp:
        target = Path(temp)
        for name in files:
            source = ROOT / name
            if source.is_symlink():
                raise RuntimeError("snapshot refuses symlinks: " + name)
            dest = target / name
            dest.parent.mkdir(parents=True, exist_ok=True)
            dest.write_bytes(source.read_bytes())
        changed = target / case["file"]
        if "plant" in case:
            changed.write_text(case["plant"])
        else:
            original = changed.read_text()
            if original.count(case["old"]) != 1:
                raise RuntimeError("mutation must match once: " + case["id"])
            changed.write_text(original.replace(case["old"], case["new"], 1))
        env = dict(os.environ, GOWORK="off", GOPROXY="off", GOSUMDB="off", KUBECONFIG=os.devnull)
        result = subprocess.run(["go", "test", "-count=1", "-v", "-timeout=4m", "-run", "^" + case["test"] + "$", "./issueflow"], cwd=target / "deskcore", env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=270)
        if result.returncode == 0 or case["expect"] not in result.stdout or "--- FAIL: " + case["test"] not in result.stdout:
            print(result.stdout)
            raise RuntimeError("mutation survived or failed for the wrong reason: " + case["id"])
        print("checked-failed as expected: " + case["id"])
        for line in result.stdout.splitlines():
            if case["expect"] in line or "--- FAIL: " + case["test"] in line:
                print(line)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--case", choices=[c["id"] for c in MANIFEST["mutations"]])
    args = parser.parse_args()
    for case in MANIFEST["mutations"]:
        if args.case is None or case["id"] == args.case:
            run(case)
