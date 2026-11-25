// swift-tools-version:5.9
// OpenHuangguo 的 SwiftPM 包（独立分发仓库）。
// 消费者：` .package(url: "https://github.com/AmatsuZero/openhuangguo-swift.git", from: "0.1.0") `
// 二进制 OpenHuangguoCore.xcframework 由 hg-xcf.yml 发布到本仓库 release 并回填 URL/checksum。
import PackageDescription
import Foundation

let publishedCore: Target = .binaryTarget(
    name: "OpenHuangguoCore",
    url: "https://github.com/AmatsuZero/openhuangguo-swift/releases/download/hg-xcf-v0.1.5/OpenHuangguoCore.xcframework.zip",
    checksum: "45dfffb712f0043e041fa4c2ec5a02d7633f453c1e11a1c49e2934e67542999c"
)

let core: Target = ProcessInfo.processInfo.environment["OPENHUANGGUO_CORE_PATH"].map {
    .binaryTarget(name: "OpenHuangguoCore", path: $0)
} ?? publishedCore

let package = Package(
    name: "OpenHuangguo",
    platforms: [.iOS(.v15), .macOS(.v12)],
    products: [
        .library(name: "OpenHuangguo", targets: ["OpenHuangguo"]),
    ],
    targets: [
        core,
        .target(
            name: "OpenHuangguo",
            dependencies: ["OpenHuangguoCore"],
            path: "Sources/OpenHuangguo"
        ),
    ]
)
