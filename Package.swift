// swift-tools-version:5.9
// OpenHuangguo 的 SwiftPM 包（独立分发仓库）。
// 消费者：` .package(url: "https://github.com/AmatsuZero/openhuangguo-swift.git", from: "0.1.0") `
// 二进制 OpenHuangguoCore.xcframework 由 hg-xcf.yml 发布到本仓库 release 并回填 URL/checksum。
import PackageDescription

let package = Package(
    name: "OpenHuangguo",
    platforms: [.iOS(.v15), .macOS(.v12)],
    products: [
        .library(name: "OpenHuangguo", targets: ["OpenHuangguo"]),
    ],
    targets: [
        .binaryTarget(
            name: "OpenHuangguoCore",
            url: "https://github.com/AmatsuZero/openhuangguo-swift/releases/download/hg-xcf-v0.0.0/OpenHuangguoCore.xcframework.zip",
            checksum: "0000000000000000000000000000000000000000000000000000000000000000"
        ),
        .target(
            name: "OpenHuangguo",
            dependencies: ["OpenHuangguoCore"],
            path: "Sources/OpenHuangguo"
        ),
    ]
)
