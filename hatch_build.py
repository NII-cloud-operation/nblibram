"""Wrap GoReleaser binaries in wheels without rebuilding them."""

import json
import os
from pathlib import Path
import subprocess
import sys

from hatchling.builders.hooks.plugin.interface import BuildHookInterface
from hatchling.metadata.plugin.interface import MetadataHookInterface


PLATFORMS = {
    ("linux", "amd64"): "manylinux_2_17_x86_64",
    ("linux", "arm64"): "manylinux_2_17_aarch64",
    ("darwin", "amd64"): "macosx_13_0_x86_64",
    ("darwin", "arm64"): "macosx_13_0_arm64",
    ("windows", "amd64"): "win_amd64",
    ("windows", "arm64"): "win_arm64",
}


class ReleaseMetadataHook(MetadataHookInterface):
    def update(self, metadata):
        release = json.loads((Path(self.root) / "dist/metadata.json").read_text())
        metadata["version"] = release["version"]


class ReleaseBuildHook(BuildHookInterface):
    def initialize(self, version, build_data):
        artifacts = json.loads((Path(self.root) / "dist/artifacts.json").read_text())
        artifact = artifacts[int(os.environ["NBLIBRAM_ARTIFACT"])]
        binary = Path(self.root) / artifact["path"]
        platform = PLATFORMS[artifact["goos"], artifact["goarch"]]
        build_data["pure_python"] = False
        build_data["tag"] = f"py3-none-{platform}"
        build_data["shared_scripts"][str(binary)] = binary.name


if __name__ == "__main__":
    root = Path(__file__).resolve().parent
    artifacts = json.loads((root / "dist/artifacts.json").read_text())
    for index, artifact in enumerate(artifacts):
        if artifact["type"] == "Binary":
            subprocess.run(
                [sys.executable, "-m", "build", "--wheel", "--no-isolation"],
                cwd=root, env=dict(os.environ, NBLIBRAM_ARTIFACT=str(index)),
                check=True,
            )
