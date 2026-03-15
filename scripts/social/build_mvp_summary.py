#!/usr/bin/env python3
"""Build a compact SNS post body from docs/mvp diffs."""

from __future__ import annotations

import argparse
import subprocess
from pathlib import Path


MAX_CHARS = 450


def run(cmd: list[str]) -> str:
    return subprocess.check_output(cmd, text=True).strip()


def changed_docs(base: str, head: str) -> list[str]:
    output = run([
        "git",
        "diff",
        "--name-only",
        f"{base}..{head}",
        "--",
        "docs/mvp",
    ])
    docs = [line for line in output.splitlines() if line.endswith(".md")]
    return sorted(dict.fromkeys(docs))


def read_highlights(base: str, head: str, paths: list[str], limit_per_file: int = 2) -> list[str]:
    highlights: list[str] = []
    for path in paths:
        diff = run([
            "git",
            "diff",
            "--unified=0",
            f"{base}..{head}",
            "--",
            path,
        ])
        additions = []
        for line in diff.splitlines():
            if not line.startswith("+"):
                continue
            if line.startswith("+++"):
                continue
            text = line[1:].strip()
            if not text:
                continue
            if text.startswith("#"):
                continue
            additions.append(text)
        for item in additions[:limit_per_file]:
            highlights.append(f"- {Path(path).name}: {item}")
    return highlights


def crop(text: str, max_chars: int = MAX_CHARS) -> str:
    if len(text) <= max_chars:
        return text
    return text[: max_chars - 1] + "…"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base", required=True)
    parser.add_argument("--head", required=True)
    parser.add_argument("--repo-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    docs = changed_docs(args.base, args.head)
    if not docs:
        Path(args.output).write_text("", encoding="utf-8")
        return 0

    title = "📌 MVP仕様の更新まとめ"
    file_list = "\n".join(f"・{Path(path).name}" for path in docs)
    highlights = read_highlights(args.base, args.head, docs)

    body_lines = [
        title,
        "",
        "更新ファイル:",
        file_list,
        "",
        "変更ポイント:",
    ]
    if highlights:
        body_lines.extend(highlights)
    else:
        body_lines.append("- 仕様の更新が入りました（詳細は差分を確認してください）")

    body_lines.extend(
        [
            "",
            f"詳細: {args.repo_url}",
            "#indiedev #gamedev #oss",
        ]
    )

    post = crop("\n".join(body_lines))
    Path(args.output).write_text(post, encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
