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

Swift 通过 SwiftPM 指向本仓库根 `Package.swift` 的 `binaryTarget`（`OpenHuangguoCore`），从本仓库 release 下载 `OpenHuangguoCore.xcframework.zip`。Android 通过本仓库 release 的 `huangguo-release.aar`——这是完整壳 AAR（Kotlin `OpenHuangguoClient` + Go `core.jar` + jni），下载后用 `implementation(files("huangguo-release.aar"))` 集成；其传递依赖（kotlinx-coroutines、media3-datasource）需在 App 侧另行声明（见 `bindings/huangguo/android/build.gradle.kts`）。

## 签名播放资源

将完整 `PlayInfo` 交给加载器，才能把 `media_headers` 用于主清单、子清单、分片与密钥。播放 URL 的转义保留原样；相对资源按重定向后的最终清单 URL 解析。

```swift
let playback = try await client.play(videoID, episode)
let loader = PlayerResourceLoader(client: client, playback: playback)
let asset = try loader.asset()
let player = AVPlayer(playerItem: AVPlayerItem(asset: asset))
// 播放期间由宿主持有 loader 和 player；AVFoundation 不会替宿主持有 delegate。
```

```kotlin
val playback = client.play(videoId, episode)
val factory = PlayerDataSourceFactory(client, defaultDataSourceFactory, playback = playback)
val uri = requireNotNull(factory.playbackUri())
// 使用 factory 创建 MediaSource，并以 uri 创建 MediaItem。
```

选择 HEVC 地址时，同时传入 `playbackURL`（Swift）/ `playbackUrl`（Kotlin），签名头限定于所选 URL 的 origin（scheme、host、有效端口）。跨源子资源不继承签名头；媒体重定向一旦跨源，后续跳转也不会恢复凭据。`Range` 由播放器请求决定，不接受签名头覆盖。

当前仅支持空值或 `GET` 的 `manifest_method`；其他方法在取流前明确报错。旧的不带 `playback` 的加载器与 `fetchSegment(url, headers)` 调用继续可用。`SegmentResult.url` 是新增的最终资源 URL；旧 Core 缺少该字段时回退请求 URL，因此要正确处理重定向后的相对资源，需要包含本次 Go 修改的新 Core。

上述为现有站点 SDK 的播放兼容逻辑；尚未接入 HGhub 新版 HPP 插件运行时。

## Token 持久化边界

默认内存 store；移动端由宿主接 Keychain / Keystore（经 core 的 `SetTokenJSON` / `GetTokenJSON` 桥接）。**Clear / Save 失败一律视为硬错误上报，不静默成功**：若 Clear 持久化失败，磁盘仍存旧 token+generation，重启会一并加载，首版不保证跨重启的登出/切域名退出生效，宿主须重试/上报。登录/注册/登出/SetBase 的原生 async 流程必须"先写持久化成功，再 resolve Promise/Future"。

## 发版

推 main → 打 `hg-xcf-v*` 触发构建 + release（二进制资产）→ 回填后打标准 semver `vX.Y.Z` 供 SwiftPM 按版本消费。`hg-xcf-v*` 仅作二进制资产 tag，不是 SwiftPM 版本。
