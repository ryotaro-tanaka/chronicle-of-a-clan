#!/usr/bin/env python3
import json
import os
import re
import subprocess
import sys
import urllib.parse
import urllib.request
from pathlib import Path


def run_git(args):
    return subprocess.check_output(["git", *args], text=True).strip()


def git_show(ref, path):
    try:
        return run_git(["show", f"{ref}:{path}"])
    except subprocess.CalledProcessError:
        return ""


def parse_ready(markdown):
    m = re.search(r"^Ready:\s*(true|false)\s*$", markdown, flags=re.IGNORECASE | re.MULTILINE)
    return bool(m and m.group(1).lower() == "true")


def extract_social_text(markdown):
    section = re.search(r"^##\s+Social\s*$([\s\S]*?)(?=^##\s+|\Z)", markdown, flags=re.MULTILINE)
    if not section:
        return ""

    body = section.group(1).strip("\n")
    lines = body.splitlines()

    # Skip Ready line and leading blank lines.
    while lines and (not lines[0].strip() or re.match(r"^Ready:\s*(true|false)\s*$", lines[0], flags=re.IGNORECASE)):
        lines.pop(0)

    # Trim surrounding empty lines.
    while lines and not lines[0].strip():
        lines.pop(0)
    while lines and not lines[-1].strip():
        lines.pop()

    return "\n".join(lines).strip()


def list_changed_mvp_files(before, after):
    if not before or set(before) == {"0"}:
        before = run_git(["hash-object", "-t", "tree", "/dev/null"])

    output = run_git(["diff", "--name-status", before, after, "--", "docs/mvp/*.md"])
    changed = []
    if not output:
        return changed

    for line in output.splitlines():
        parts = line.split("\t")
        if not parts:
            continue
        status = parts[0]
        if status.startswith("R") and len(parts) >= 3:
            path = parts[2]
        elif len(parts) >= 2:
            path = parts[1]
        else:
            continue
        changed.append((status, path))
    return changed


def http_get_json(url):
    with urllib.request.urlopen(url) as res:
        return json.loads(res.read().decode("utf-8"))


def http_post_form(url, form_data):
    encoded = urllib.parse.urlencode(form_data).encode("utf-8")
    req = urllib.request.Request(url, data=encoded, method="POST")
    with urllib.request.urlopen(req) as res:
        return json.loads(res.read().decode("utf-8"))


def main():
    before = os.getenv("GITHUB_BEFORE", "")
    after = os.getenv("GITHUB_SHA", "")
    token = os.getenv("THREADS_LONG_LIVED_TOKEN", "")

    if not token:
        print("THREADS_LONG_LIVED_TOKEN is not set; skipping Threads posting.")
        return 0

    if not after:
        print("GITHUB_SHA is missing.", file=sys.stderr)
        return 1

    changed = list_changed_mvp_files(before, after)
    if not changed:
        print("No docs/mvp/*.md changes found.")
        return 0

    posts = []
    for status, path in changed:
        curr = Path(path).read_text(encoding="utf-8") if Path(path).exists() else ""
        if not curr:
            continue

        prev = git_show(before, path) if before and set(before) != {"0"} else ""

        ready_before = parse_ready(prev)
        ready_after = parse_ready(curr)
        is_new = status.startswith("A")
        turned_ready = (not ready_before) and ready_after

        if not ((is_new and ready_after) or turned_ready):
            continue

        text = extract_social_text(curr)
        if not text:
            print(f"Skipping {path}: ## Social content is empty.")
            continue

        posts.append((path, text))

    if not posts:
        print("No eligible posts found.")
        return 0

    me_url = f"https://graph.threads.net/v1.0/me?fields=id,username&access_token={urllib.parse.quote(token)}"
    me = http_get_json(me_url)
    user_id = me.get("id")
    if not user_id:
        print(f"Could not resolve Threads user id: {me}", file=sys.stderr)
        return 1

    for path, text in posts:
        create_url = f"https://graph.threads.net/v1.0/{user_id}/threads"
        created = http_post_form(create_url, {
            "media_type": "TEXT",
            "text": text,
            "access_token": token,
        })
        creation_id = created.get("id")
        if not creation_id:
            print(f"Failed to create post for {path}: {created}", file=sys.stderr)
            return 1

        publish_url = f"https://graph.threads.net/v1.0/{user_id}/threads_publish"
        published = http_post_form(publish_url, {
            "creation_id": creation_id,
            "access_token": token,
        })

        if not published.get("id"):
            print(f"Failed to publish post for {path}: {published}", file=sys.stderr)
            return 1

        print(f"Posted to Threads for {path}: {published.get('id')}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
