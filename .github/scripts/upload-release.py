import hashlib
import json
import os
import subprocess
import sys
from pathlib import Path


def gh(*args):
    return subprocess.check_output(["gh", *args], text=True)


def list_all(endpoint):
    return [
        item
        for page in json.loads(gh("api", "--paginate", "--slurp", f"{endpoint}?per_page=100"))
        for item in page
    ]


def find_release(endpoint, tag):
    matches = [release for release in list_all(endpoint) if release["tag_name"] == tag]
    if len(matches) > 1:
        raise SystemExit(f"Multiple releases for {tag}; resolve the conflict before retrying.")
    return matches[0] if matches else None


def digest(path):
    with path.open("rb") as asset:
        return "sha256:" + hashlib.file_digest(asset, "sha256").hexdigest()


def check_assets(endpoint, expected):
    uploaded = set()
    seen = set()
    unfinished = []
    for asset in list_all(endpoint):
        name = asset["name"]
        if name not in expected or name in seen:
            raise SystemExit(f"Unexpected or duplicate release asset: {name}")
        seen.add(name)
        if asset["state"] == "starter" and asset["size"] == 0:
            unfinished.append(asset["id"])
        elif asset["state"] == "uploaded" and asset.get("digest") == expected[name]:
            uploaded.add(name)
        else:
            raise SystemExit(f"Existing release asset differs from this run: {name}")
    return sorted(expected.keys() - uploaded), unfinished


def main(tag, directory):
    endpoint = f"repos/{os.environ['GH_REPO']}/releases"
    notes = directory / "RELEASE_NOTES.md"
    expected = {path.name: digest(path) for path in directory.iterdir()}
    release = find_release(endpoint, tag)
    if release is None:
        gh("release", "create", tag, "--verify-tag", "--draft", "--title", tag, "--notes-file", str(notes))
        release = find_release(endpoint, tag)
    if release is None or not release["draft"] or release["prerelease"]:
        raise SystemExit(f"Expected a stable draft release for {tag}.")
    if release["name"] != tag or (release.get("body") or "").rstrip() != notes.read_text(encoding="utf-8").rstrip():
        raise SystemExit(f"The existing draft's title or notes differ from the prepared release for {tag}.")
    assets = f"{endpoint}/{release['id']}/assets"
    missing, unfinished = check_assets(assets, expected)
    for asset_id in unfinished:
        gh("api", "--method", "DELETE", f"{endpoint}/assets/{asset_id}")
    if missing:
        gh("release", "upload", tag, *(str(directory / name) for name in missing))
    if check_assets(assets, expected) != ([], []):
        raise SystemExit(f"Release asset verification is incomplete for {tag}.")


if __name__ == "__main__":
    main(sys.argv[1], Path(sys.argv[2]))
