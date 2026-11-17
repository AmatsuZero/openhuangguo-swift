import AVFoundation
import Foundation
import UniformTypeIdentifiers

/// Loads playback resources through Core. Retain this loader while its asset is in use.
@available(iOS 15.0, macOS 12.0, *)
public final class PlayerResourceLoader: NSObject, AVAssetResourceLoaderDelegate, @unchecked Sendable {
    private let client: OpenHuangguoClient
    private let scheme: String
    private let host: String
    private let playback: PlayInfo?
    private let playbackURL: URL?
    private let queue = DispatchQueue(label: "dev.openhuangguo.player.loader", qos: .userInitiated)
    // Accessed only on queue, including completion and cancellation.
    private var tasks: [ObjectIdentifier: Task<Void, Never>] = [:]

    public init(client: OpenHuangguoClient, scheme: String = "hg", host: String = "localhost",
                playback: PlayInfo? = nil, playbackURL: URL? = nil) {
        self.client = client
        self.scheme = scheme
        self.host = host
        self.playback = playback
        self.playbackURL = playbackURL ?? playback.flatMap { URL(string: $0.videoUrl) }
    }

    /// Creates the signed playback asset; playbackURL may select the HEVC stream.
    public func asset() throws -> AVURLAsset {
        try validateManifestMethod()
        guard let url = playbackURL, ["http", "https"].contains(url.scheme?.lowercased() ?? ""), url.host != nil else {
            throw failure("invalid playback URL")
        }
        var local = URLComponents()
        local.scheme = scheme
        local.host = host
        local.path = "/seg"
        local.queryItems = [URLQueryItem(name: "u", value: url.absoluteString)]
        guard let localURL = local.url else { throw failure("invalid local media URL") }
        return asset(forLocalURL: localURL)
    }

    public func asset(forLocalURL local: URL) -> AVURLAsset {
        let asset = AVURLAsset(url: local)
        asset.resourceLoader.setDelegate(self, queue: queue)
        return asset
    }

    public func resourceLoader(_ resourceLoader: AVAssetResourceLoader,
                               shouldWaitForLoadingOfRequestedResource request: AVAssetResourceLoadingRequest) -> Bool {
        guard let url = request.request.url,
              let upstream = Self.upstreamURL(from: url, scheme: scheme, host: host) else {
            request.finishLoading(with: failure("invalid local media URL"))
            return true
        }
        queue.async { [self] in
            guard !request.isCancelled else { return }
            let id = ObjectIdentifier(request)
            let dataRequest = request.dataRequest
            let offset = dataRequest?.currentOffset ?? 0
            let length = dataRequest.flatMap { data -> Int? in
                if data.requestsAllDataToEndOfResource { return nil }
                let consumed = max(0, data.currentOffset - data.requestedOffset)
                return max(0, data.requestedLength - Int(consumed))
            }
            tasks[id] = Task { [self] in
                do {
                    let resource = try await self.load(upstream, offset: offset, length: length)
                    self.queue.async { [self] in
                        guard tasks.removeValue(forKey: id) != nil, !request.isCancelled else { return }
                        if let info = request.contentInformationRequest {
                            info.contentLength = resource.contentLength
                            info.contentType = resource.contentType
                            info.isByteRangeAccessSupported = true
                        }
                        request.dataRequest?.respond(with: resource.body)
                        request.finishLoading()
                    }
                } catch {
                    self.queue.async { [self] in
                        guard tasks.removeValue(forKey: id) != nil, !request.isCancelled else { return }
                        request.finishLoading(with: error)
                    }
                }
            }
        }
        return true
    }

    public func resourceLoader(_ resourceLoader: AVAssetResourceLoader,
                               didCancel request: AVAssetResourceLoadingRequest) {
        queue.async { [self] in tasks.removeValue(forKey: ObjectIdentifier(request))?.cancel() }
    }

    struct Resource {
        let body: Data
        let contentLength: Int64
        let contentType: String
    }

    // Shared by the delegate and tests: requests real Core resources, then adapts
    // HTTP byte offsets to the bytes AVFoundation actually requested.
    func load(_ upstream: URL, offset: Int64, length: Int?) async throws -> Resource {
        try validateManifestMethod()
        guard offset >= 0, length == nil || length! >= 0 else { throw failure("invalid media range") }
        var headers: [String: String] = [:]
        if let root = playbackURL, Self.sameOrigin(root, upstream) {
            headers = playback?.mediaHeaders ?? [:]
        }
        // The player owns Range. Signed headers cannot override the requested bytes.
        headers = headers.filter { $0.key.lowercased() != "range" }
        if upstream.pathExtension.lowercased() != "m3u8" {
            if let length, length > 0 {
                let (end, overflow) = offset.addingReportingOverflow(Int64(length - 1))
                guard !overflow else { throw failure("invalid media range") }
                headers["Range"] = "bytes=\(offset)-\(end)"
            } else if offset > 0 { headers["Range"] = "bytes=\(offset)-" }
        }
        var result = try await client.fetchSegment(upstream.absoluteString, headers)
        try Task.checkCancellation()
        var body = try decodedBody(result)
        if Self.looksLikePlaylist(url: upstream, headers: result.headers, body: body) {
            // A MIME-detected playlist may have been requested with Range. Rewrite
            // only complete manifests, never a truncated 206 response.
            if result.status == 206 {
                headers.removeValue(forKey: "Range")
                result = try await client.fetchSegment(upstream.absoluteString, headers)
                try Task.checkCancellation()
                body = try decodedBody(result)
            }
            guard result.status == 200 else { throw failure("incomplete media playlist") }
            let base = result.url.isEmpty ? upstream.absoluteString : result.url
            let rewritten = try await client.rewritePlaylist(base, scheme, host, String(decoding: body, as: UTF8.self))
            try Task.checkCancellation()
            body = Data(rewritten.utf8)
            return Resource(body: slice(body, offset: offset, length: length), contentLength: Int64(body.count), contentType: "com.apple.m3u-playlist")
        }
        var start: Int64 = 0
        var total = Int64(body.count)
        if result.status == 206 {
            guard let range = Self.header("Content-Range", in: result.headers), range.hasPrefix("bytes ") else {
                throw failure("missing media Content-Range")
            }
            let parts = range.dropFirst(6).split(separator: "/")
            let bounds = parts.first?.split(separator: "-") ?? []
            guard parts.count == 2, bounds.count == 2,
                  let first = Int64(bounds[0]), let last = Int64(bounds[1]), let size = Int64(parts[1]),
                  first >= 0, last >= first, last < size, last - first == Int64(body.count) - 1,
                  first <= offset, offset <= last + 1 else { throw failure("invalid media Content-Range") }
            start = first
            total = size
        }
        let mime = Self.header("Content-Type", in: result.headers)?.split(separator: ";").first.map(String.init)
        let type = mime.flatMap { UTType(mimeType: $0)?.identifier } ?? UTType.data.identifier
        return Resource(body: slice(body, offset: offset - start, length: length), contentLength: total, contentType: type)
    }

    private func decodedBody(_ result: SegmentResult) throws -> Data {
        guard result.status == 200 || result.status == 206 else { throw failure("media HTTP \(result.status)") }
        guard let body = Data(base64Encoded: result.body) else { throw failure("invalid media body") }
        return body
    }

    private func validateManifestMethod() throws {
        let method = (playback?.manifestMethod ?? "").trimmingCharacters(in: .whitespacesAndNewlines).uppercased()
        guard method.isEmpty || method == "GET" else { throw failure("unsupported manifest method: \(method)") }
    }

    private func slice(_ data: Data, offset: Int64, length: Int?) -> Data {
        guard offset >= 0, offset < data.count else { return Data() }
        let start = Int(offset)
        let count = min(length ?? data.count, data.count - start)
        return data.subdata(in: start..<(start + count))
    }

    private func failure(_ message: String) -> NSError {
        NSError(domain: "dev.openhuangguo.player", code: -1, userInfo: [NSLocalizedDescriptionKey: message])
    }

    static func upstreamURL(from local: URL, scheme: String, host: String) -> URL? {
        guard local.scheme?.lowercased() == scheme.lowercased(), local.host?.lowercased() == host.lowercased(),
              local.path == "/seg", let components = URLComponents(url: local, resolvingAgainstBaseURL: false),
              let value = components.queryItems?.first(where: { $0.name == "u" })?.value else { return nil }
        // URLComponents has already decoded the query value. A second decode
        // corrupts signed URLs containing escaped path separators or query values.
        return URL(string: value)
    }

    private static func sameOrigin(_ a: URL, _ b: URL) -> Bool {
        func port(_ url: URL) -> Int { url.port ?? (url.scheme?.lowercased() == "https" ? 443 : 80) }
        return a.scheme?.lowercased() == b.scheme?.lowercased() && a.host?.lowercased() == b.host?.lowercased() && port(a) == port(b)
    }

    private static func header(_ name: String, in headers: [String: String]) -> String? {
        headers.first { $0.key.caseInsensitiveCompare(name) == .orderedSame }?.value
    }

    private static func looksLikePlaylist(url: URL, headers: [String: String], body: Data) -> Bool {
        let mime = header("Content-Type", in: headers)?.lowercased() ?? ""
        return url.pathExtension.lowercased() == "m3u8" || mime.contains("mpegurl") || mime.contains("m3u8") || body.starts(with: Data("#EXTM3U".utf8))
    }
}
