// swift-tools-version:5.9
// 本地开发用：binaryTarget 指向 build.sh ios 生成的 xcframework。
// 发布用根 Package.swift（指向本仓库 release 资产），见仓库根。
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
            path: "OpenHuangguoCore.xcframework"
        ),
        .target(
            name: "OpenHuangguo",
            dependencies: ["OpenHuangguoCore"],
            path: "Sources/OpenHuangguo"
        ),
        .testTarget(name: "OpenHuangguoTests", dependencies: ["OpenHuangguo"]),
    ]
)
