import AVFoundation
import Foundation

/// 把 Core 的纯函数媒体工具接到 AVPlayer 的资源加载层。
///
/// Core 的 RewritePlaylist 把 KEY / 分片 / 子清单 URI 改写成
/// `hg://localhost/seg?u=<pct>`（u 为上游完整 URL 的 percent 编码）。
/// 本 loader 还原上游 URL，经 Core 的 fetchSegment / rewritePlaylist 取回资源，
/// 并把 Range 与 206 / Content-Range 正确应答给播放器。
@available(iOS 15.0, macOS 12.0, *)
public final class PlayerResourceLoader: NSObject, AVAssetResourceLoaderDelegate, @unchecked Sendable {

    private let client: OpenHuangguoClient
    private let scheme: String
    private let host: String
    private let queue = DispatchQueue(label: "dev.openhuangguo.player.loader", qos: .userInitiated)

    public init(client: OpenHuangguoClient, scheme: String = "hg", host: String = "localhost") {
        self.client = client
        self.scheme = scheme
        self.host = host
    }

    /// 创建可交给 AVPlayer 的 AVAsset；scheme 必须与传给 RewritePlaylist 的一致。
    public func asset(forLocalURL local: URL) -> AVAsset {
        AVAsset(url: local)
    }

    public func resourceLoader(_ resourceLoader: AVAssetResourceLoader,
                               shouldWaitForLoadingOfRequestedResource loadingRequest: AVAssetResourceLoadingRequest) -> Bool {
        guard let url = loadingRequest.request.url,
              let upstream = Self.upstreamURL(from: url, scheme: scheme, host: host) else {
            loadingRequest.finishLoading(with: NSError(domain: "dev.openhuangguo", code: -1,
                                                        userInfo: [NSLocalizedDescriptionKey: "invalid local media URL"]))
            return true
        }
        queue.async { [weak self] in
            guard let self else { return }
            Task {
                do {
                    try await self.serve(loadingRequest, upstream: upstream)
                } catch {
                    loadingRequest.finishLoading(with: error)
                }
            }
        }
        return true
    }

    public func resourceLoader(_ resourceLoader: AVAssetResourceLoader,
                               didCancel loadingRequest: AVAssetResourceLoadingRequest) {
        // Core 的 Start/Call 在 Task 取消时由调用方取消；此处不持有长任务引用。
    }

    // MARK: - 处理

    private func serve(_ request: AVAssetResourceLoadingRequest, upstream: URL) async throws {
        let range = request.dataRequest?.requestsAllDataToEndOfResource == true
            ? nil
            : (request.dataRequest.map { "bytes=\($0.requestedOffset)-\($0.requestedOffset + Int64(max($0.requestedLength, 1)) - 1)" })
        var headers: [String: String] = [:]
        if let range { headers["Range"] = range }
        let result = try await client.fetchSegment(upstream.absoluteString, headers)

        guard let data = Data(base64Encoded: result.body) else {
            throw NSError(domain: "dev.openhuangguo", code: -2, userInfo: [NSLocalizedDescriptionKey: "bad segment body"])
        }

        // 子清单递归：扩展名 / MIME / #EXTM3U 内容任一命中即再次改写。
        if Self.looksLikePlaylist(url: upstream, headers: result.headers, body: data) {
            let text = String(decoding: data, as: UTF8.self)
            let rewritten = try await client.rewritePlaylist(upstream.absoluteString, scheme, host, text)
            let out = Data(rewritten.utf8)
            serveData(request, data: out, status: result.status, headers: result.headers, isPlaylist: true)
            return
        }

        serveData(request, data: data, status: result.status, headers: result.headers, isPlaylist: false)
    }

    private func serveData(_ request: AVAssetResourceLoadingRequest, data: Data, status: Int,
                           headers: [String: String], isPlaylist: Bool) {
        if let info = request.contentInformationRequest {
            if let len = headers["Content-Length"].flatMap({ Int($0) }) { info.contentLength = Int64(len) }
            info.contentType = isPlaylist ? "application/vnd.apple.mpegurl" : (headers["Content-Type"] ?? "video/mp2t")
            info.isByteRangeAccessSupported = true
        }
        if let dataRequest = request.dataRequest {
            let offset = Int(dataRequest.currentOffset)
            if offset < data.count {
                dataRequest.respond(with: data.subdata(in: offset..<data.count))
            } else {
                dataRequest.respond(with: Data())
            }
        }
        let response = HTTPURLResponse(url: request.request.url!, statusCode: status,
                                       httpVersion: "HTTP/1.1", headerFields: headers)!
        request.response = response
        request.finishLoading()
    }

    // MARK: - 工具

    /// 从 `hg://localhost/seg?u=<pct>` 还原上游 URL（scheme+host 显式传入，避免重复拼接）。
    static func upstreamURL(from local: URL, scheme: String, host: String) -> URL? {
        guard local.scheme?.lowercased() == scheme.lowercased(),
              local.host?.lowercased() == host.lowercased(),
              let components = URLComponents(url: local, resolvingAgainstBaseURL: false),
              let encoded = components.queryItems?.first(where: { $0.name == "u" })?.value else {
            return nil
        }
        return URL(string: encoded.removingPercentEncoding ?? encoded)
    }

    private static func looksLikePlaylist(url: URL, headers: [String: String], body: Data) -> Bool {
        let ext = (url.pathExtension as NSString).lowercased
        if ext == "m3u8" { return true }
        let cth = (headers["Content-Type"] ?? "").lowercased()
        if cth.contains("mpegurl") || cth.contains("m3u8") { return true }
        let prefix = "#EXTM3U".utf8
        return body.count >= prefix.count && Array(body.prefix(prefix.count)) == Array(prefix)
    }
}
