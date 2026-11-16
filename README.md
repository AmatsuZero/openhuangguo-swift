# OpenHuangguoCore

黄果（HGhub）Go SDK + gomobile Core（OpenHuangguoCore）+ Swift/Kotlin 移动端绑定，独立仓库，与 hongguo / OpenHongguoCore 完全隔离。

## 工具链

- Go 1.26（core module 因 gomobile tool 用 1.26；sdk / gen 用 1.22 亦可）
- gomobile / gobind（core 构建期）
- Xcode（iOS xcframework）
- Android SDK + JDK（AAR）

## 本地开发

本地只做 Go 源码编译校验与测试；xcframework / AAR 编译、打包、发布一律走 CI（`.github/workflows/hg-xcf.yml`）。

```bash
make fmt            # 逐 module gofmt
make vet            # GOWORK=off 逐 module go vet ./...
make test           # GOWORK=off 逐 module go test ./...
make gen-check      # 重新生成 Types.swift / Types.kt 并校验无 drift
make all            # 以上全部
```

本地临时 `go.work` 仅供个人开发，不入库、不进 CI（见 `.gitignore`）。

## 模块

- `bindings/huangguo/sdk` — 业务接口、transport、edge、token、媒体纯函数
- `bindings/huangguo/core` — gomobile 入口（OpenHuangguoCore），按名调用，复杂数据走 `Start`/callback JSON
- `bindings/huangguo/gen-huangguo` — Swift/Kotlin 类型生成器（构建期 schema 依赖仅在此）

## 消费方式

Swift 通过 SwiftPM 指向本仓库根 `Package.swift` 的 `binaryTarget`（`OpenHuangguoCore`），从本仓库 release 下载 `OpenHuangguoCore.xcframework.zip`。Android 通过本仓库 release 的 `huangguo-release.aar`。

## Token 持久化边界

默认内存 store；移动端由宿主接 Keychain / Keystore（经 core 的 `SetTokenJSON` / `GetTokenJSON` 桥接）。**Clear / Save 失败一律视为硬错误上报，不静默成功**：若 Clear 持久化失败，磁盘仍存旧 token+generation，重启会一并加载，首版不保证跨重启的登出/切域名退出生效，宿主须重试/上报。登录/注册/登出/SetBase 的原生 async 流程必须"先写持久化成功，再 resolve Promise/Future"。

## 发版

推 main → 打 `hg-xcf-v*` 触发构建 + release（二进制资产）→ 回填后打标准 semver `vX.Y.Z` 供 SwiftPM 按版本消费。`hg-xcf-v*` 仅作二进制资产 tag，不是 SwiftPM 版本。
