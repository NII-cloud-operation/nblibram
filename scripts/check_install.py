"""Check that the wheel installs a working CLI without a Go toolchain."""

import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import venv


wheels = Path(sys.argv[1]).resolve()
with tempfile.TemporaryDirectory(prefix="nblibram-install-") as directory:
    venv.create(directory, with_pip=True)
    scripts = Path(directory) / ("Scripts" if os.name == "nt" else "bin")
    python = scripts / ("python.exe" if os.name == "nt" else "python")
    env = dict(os.environ, PATH=str(scripts))
    subprocess.run(
        [str(python), "-m", "pip", "install", "--no-index", "--only-binary=:all:",
         "--find-links", str(wheels), "nblibram"],
        env=env, check=True,
    )
    subprocess.run([str(python), "-m", "pip", "check"], env=env, check=True)
    binary = shutil.which("nblibram", path=env["PATH"])
    assert binary is not None, "Installed nblibram command is missing"
    assert not shutil.which("go", path=env["PATH"])
    version = subprocess.check_output(
        [str(python), "-c", "from importlib.metadata import version; print(version('nblibram'))"],
        env=env, encoding="utf-8",
    ).strip()
    output = subprocess.check_output([binary, "version"], env=env, encoding="utf-8")
    assert output.strip() == f"nblibram {version}", output

    heading = "# Installation test 日本語\n"
    notebook = json.dumps({
        "nbformat": 4, "nbformat_minor": 5, "metadata": {},
        "cells": [{"cell_type": "markdown", "id": "heading", "metadata": {},
                   "source": [heading]}],
    }, ensure_ascii=False)
    for args in (["toc"], ["cells", "--query", "start:0", "--count", "1"]):
        result = subprocess.run(
            [binary, *args, "--format", "json"], input=notebook,
            stdout=subprocess.PIPE, encoding="utf-8", env=env, check=True,
        )
        cells = json.loads(result.stdout)["cells"]
        assert "".join(cells[0]["source"]) == heading, result.stdout

    result = subprocess.run(
        [binary, "invalid-command"], capture_output=True, encoding="utf-8", env=env,
    )
    assert result.returncode == 1, result
    assert "unknown command" in result.stderr, result.stderr
    print("Installed wheel: version, notebook queries, Unicode stdin and exit status OK")
