#!/usr/bin/env python3
"""验证回填脚本只改 OpenHuangguoCore 目标，其他 binaryTarget 不受影响。"""
import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(__file__))
from fix_huangguo_binarytarget import rewrite  # noqa: E402

FIXTURE = '''// swift-tools-version:5.9
import PackageDescription
let package = Package(
    name: "OpenHuangguo",
    targets: [
        .binaryTarget(
            name: "OpenHuangguoCore",
            url: "https://github.com/AmatsuZero/openhuangguo-swift/releases/download/hg-xcf-v0.0.0/OpenHuangguoCore.xcframework.zip",
            checksum: "0dd00000"
        ),
        .binaryTarget(
            name: "OtherCore",
            url: "https://example.com/other.zip",
            checksum: "keepme1111"
        ),
    ]
)
'''


class BackfillTest(unittest.TestCase):
    def test_only_target_block_changes(self):
        out = rewrite(FIXTURE, "hg-xcf-v0.1.0", "newchecksum2222")
        self.assertIn("hg-xcf-v0.1.0/OpenHuangguoCore.xcframework.zip", out)
        self.assertIn('checksum: "newchecksum2222"', out)
        # 其他目标原样保留。
        self.assertIn('name: "OtherCore"', out)
        self.assertIn("https://example.com/other.zip", out)
        self.assertIn('checksum: "keepme1111"', out)
        self.assertNotIn('checksum: "0dd00000"', out)

    def test_missing_block_raises(self):
        with self.assertRaises(ValueError):
            rewrite('targets: [ .binaryTarget(name: "X", url: "u", checksum: "c") ]', "t", "c")


if __name__ == "__main__":
    unittest.main()
