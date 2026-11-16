#!/usr/bin/env python3
"""只改写根 Package.swift 中 name == "OpenHuangguoCore" 的 binaryTarget，
不碰其他 binaryTarget，也不做全局 checksum 替换。

Usage: python3 scripts/fix_huangguo_binarytarget.py <release-tag> <sha256>
"""
import os
import re
import sys

BLOCK_RE = re.compile(
    r'(\.binaryTarget\(\s*name: "OpenHuangguoCore".*?)(\n\s*\.binaryTarget\(|\n\s*\])',
    re.S,
)


def rewrite(text: str, tag: str, checksum: str) -> str:
    """返回改写后的文本；目标块内 url/checksum 必须各命中恰好一次。"""
    new_url = (
        "https://github.com/AmatsuZero/openhuangguo-swift/releases/download/%s/"
        "OpenHuangguoCore.xcframework.zip" % tag
    )
    m = BLOCK_RE.search(text)
    if not m:
        raise ValueError("OpenHuangguoCore binaryTarget block not found")
    block = m.group(1)
    new_block, n_url = re.subn(r'url: "[^"]*"', 'url: "%s"' % new_url, block)
    new_block, n_ck = re.subn(r'checksum: "[0-9a-fA-F]*"', 'checksum: "%s"' % checksum, new_block)
    if n_url != 1 or n_ck != 1:
        raise ValueError("expected exactly one url and checksum in target block, got %d/%d" % (n_url, n_ck))
    return text[:m.start(1)] + new_block + text[m.end(1):]


def main() -> None:
    if len(sys.argv) != 3:
        raise SystemExit("usage: fix_huangguo_binarytarget.py <release-tag> <sha256>")
    tag, checksum = sys.argv[1], sys.argv[2]
    path = os.path.join(os.path.dirname(__file__), "..", "Package.swift")
    with open(path) as f:
        text = f.read()
    text = rewrite(text, tag, checksum)
    with open(path, "w") as f:
        f.write(text)
    print("--- updated OpenHuangguoCore binaryTarget ---")
    for line in text.splitlines():
        if "url:" in line or "checksum:" in line:
            print(line.strip())


if __name__ == "__main__":
    main()
