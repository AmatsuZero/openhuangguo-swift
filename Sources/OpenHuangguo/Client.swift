import Foundation
import OpenHuangguoCore

/// 错误码，对应 Go dispatch 的错误码。
public enum ErrorCode: String, Sendable, Codable {
    case cancelled = "CANCELLED"
    case authenticationRequired = "AUTHENTICATION_REQUIRED"
    case apiError = "API_ERROR"
    case invalidBase = "INVALID_BASE"
    case retryable = "RETRYABLE"
    case transport = "TRANSPORT"
    case unknown = "UNKNOWN"

    public init(from decoder: Decoder) throws {
        let raw = try decoder.singleValueContainer().decode(String.self)
        self = ErrorCode(rawValue: raw) ?? .unknown
    }
}

/// 黄果 SDK 返回的错误。
public struct HuangguoError: Error, LocalizedError, Sendable, Codable {
    public var code: ErrorCode
    public var message: String
    public var retryable: Bool
    public var retryAfterMs: Int

    public init(code: ErrorCode, message: String, retryable: Bool = false, retryAfterMs: Int = 0) {
        self.code = code
        self.message = message
        self.retryable = retryable
        self.retryAfterMs = retryAfterMs
    }

    public var errorDescription: String? { message }

    static func parse(_ message: String) -> HuangguoError {
        (try? JSONDecoder().decode(HuangguoError.self, from: Data(message.utf8)))
            ?? HuangguoError(code: .unknown, message: message)
    }
}

/// 原生 token 持久化协议。宿主实现必须线程安全（任意 Go goroutine 可调用）。
public protocol TokenStoring: Sendable {
    func load() throws -> String?
    func save(_ json: String) throws
    func clear() throws
}

/// 客户端设置。
public struct ClientOptions: Sendable {
    public var base: String?
    public var proxy: String?
    public var credTrustOrigins: [String: Bool]?
    public var tokenStore: (any TokenStoring)?
    public var logger: (@Sendable (String) -> Void)?

    public init(base: String? = nil, proxy: String? = nil, credTrustOrigins: [String: Bool]? = nil,
                tokenStore: (any TokenStoring)? = nil, logger: (@Sendable (String) -> Void)? = nil) {
        self.base = base
        self.proxy = proxy
        self.credTrustOrigins = credTrustOrigins
        self.tokenStore = tokenStore
        self.logger = logger
    }

    func json() throws -> String {
        var obj: [String: Any] = [:]
        if let base { obj["base"] = base }
        if let proxy { obj["proxy"] = proxy }
        if let trust = credTrustOrigins { obj["credTrustOrigins"] = trust }
        let data = try JSONSerialization.data(withJSONObject: obj)
        return String(decoding: data, as: UTF8.self)
    }
}

private final class Callback: NSObject, CoreCallbackProtocol, @unchecked Sendable {
    let handler: @Sendable (String) -> Void
    init(_ handler: @escaping @Sendable (String) -> Void) { self.handler = handler }
    func done(_ resultJSON: String?) { handler(resultJSON ?? "") }
}

private final class Logger: NSObject, CoreLoggerProtocol, @unchecked Sendable {
    let sink: @Sendable (String) -> Void
    init(_ sink: @escaping @Sendable (String) -> Void) { self.sink = sink }
    func log(_ line: String?) { sink(line ?? "") }
}

// 登录/注册结果：token 已写入原生 store，仅回传 user（原始 JSON）。
public struct AuthResult: Codable, Sendable {
    public var user: Data
    public init(user: Data = Data()) { self.user = user }
    enum CodingKeys: String, CodingKey { case user = "user" }
}

internal protocol Engine: Sendable {
    func start(_ method: String, _ args: String, _ done: @escaping @Sendable (String) -> Void) -> @Sendable () -> Void
    func setTokenJSON(_ json: String) throws
    func getTokenJSON() throws -> String
}

internal struct GoEngine: Engine, @unchecked Sendable {
    let client: CoreClient
    private struct CallRef: @unchecked Sendable { let call: CoreCall? }

    func start(_ method: String, _ args: String, _ done: @escaping @Sendable (String) -> Void) -> @Sendable () -> Void {
        let ref = CallRef(call: client.start(method, argsJSON: args, callback: Callback(done)))
        return { ref.call?.cancel() }
    }
    func setTokenJSON(_ json: String) throws {
        // Go 的 `error` 返回值被 gomobile 映射为 Swift 的 throws。
        try client.setTokenJSON(json)
    }
    func getTokenJSON() throws -> String {
        // gomobile 对「仅返回 error」用 throws，但对「返回值 + error」保留 NSError** 参数。
        var err: NSError?
        let s = client.getTokenJSON(&err)
        if let err { throw HuangguoError.parse(err.localizedDescription) }
        return s
    }
}

private final class CancelBox: @unchecked Sendable {
    private let lock = NSLock()
    private var cancel: (@Sendable () -> Void)?
    private var cancelled = false
    func set(_ f: @escaping @Sendable () -> Void) {
        lock.lock(); let now = cancelled
        if !now { cancel = f }
        lock.unlock()
        if now { f() }
    }
    func fire() {
        lock.lock(); cancelled = true; let f = cancel; cancel = nil
        lock.unlock(); f?()
    }
}

private struct Envelope<T: Decodable>: Decodable {
    var result: T?
    var error: HuangguoError?
}

private struct Nothing: Decodable {}

/// 黄果客户端。所有方法 async；失败抛 HuangguoError；Task 取消抛 CancellationError。
public final class OpenHuangguoClient: Sendable {
    private let engine: any Engine
    private let tokenStore: (any TokenStoring)?

    init(engine: any Engine, tokenStore: (any TokenStoring)?) {
        self.engine = engine
        self.tokenStore = tokenStore
    }

    /// 创建客户端。启动时若有持久化 token，先经 SetTokenJSON 恢复。
    public static func create(options: ClientOptions = ClientOptions()) throws -> OpenHuangguoClient {
        let logger = options.logger.map { Logger($0) }
        var err: NSError?
        guard let client = CoreNewClient(try options.json(), logger, &err) else {
            throw HuangguoError.parse(err?.localizedDescription ?? "unknown error")
        }
        let engine = GoEngine(client: client)
        if let store = options.tokenStore, let json = try store.load() {
            try engine.setTokenJSON(json)
        }
        return OpenHuangguoClient(engine: engine, tokenStore: options.tokenStore)
    }

    // MARK: - 账号

    public func login(_ username: String, _ password: String) async throws -> AuthResult {
        try await callPersist("login", username, password)
    }
    public func register(_ username: String, _ password: String, _ email: String = "") async throws -> AuthResult {
        try await callPersist("register", username, password, email)
    }
    public func logout() async throws {
        try await callVoidPersist("logout")
    }
    public func me() async throws -> User { try await call("me") }

    // MARK: - 列表 / 详情

    public func feed(_ page: Int = 1, _ size: Int = 20) async throws -> VideoCardPage {
        try await call("feed", page, size)
    }
    public func category(_ cate: String, _ page: Int = 1, _ size: Int = 20) async throws -> VideoCardPage {
        try await call("category", cate, page, size)
    }
    public func rank(_ kind: String = "hot") async throws -> RankPage { try await call("rank", kind) }
    public func search(_ q: String, _ page: Int = 1, _ size: Int = 20) async throws -> SearchResult {
        try await call("search", q, page, size)
    }
    public func posts(_ page: Int = 1, _ size: Int = 20) async throws -> PostPage {
        try await call("posts", page, size)
    }
    public func post(_ id: Int64) async throws -> Post { try await call("post", id) }
    public func video(_ id: Int64) async throws -> Detail { try await call("video", id) }
    public func play(_ id: Int64, _ ep: Int) async throws -> PlayInfo { try await call("play", id, ep) }

    // MARK: - 评论 / 互动

    public func comments(_ id: Int64, _ ep: Int, _ page: Int = 1, _ size: Int = 20) async throws -> CommentPage {
        try await call("comments", id, ep, page, size)
    }
    public func commentSend(_ id: Int64, _ ep: Int, _ text: String, _ replyID: Int64 = 0) async throws -> String {
        try await call("commentSend", id, ep, text, replyID)
    }
    public func like(_ id: Int64, _ ep: Int, _ on: Bool) async throws { try await callVoid("like", id, ep, on) }
    public func collect(_ id: Int64, _ on: Bool) async throws { try await callVoid("collect", id, on) }
    public func follow(_ uid: Int64, _ on: Bool) async throws { try await callVoid("follow", uid, on) }
    public func postLike(_ id: Int64, _ on: Bool) async throws { try await callVoid("postLike", id, on) }
    public func postCollect(_ id: Int64, _ on: Bool) async throws { try await callVoid("postCollect", id, on) }
    public func postComments(_ id: Int64, _ page: Int = 1, _ size: Int = 20) async throws -> CommentPage {
        try await call("postComments", id, page, size)
    }

    /// 切换线路：成功后清除 token 并持久化（generation++）。
    public func setBase(_ base: String) async throws {
        try await callVoidPersist("setBase", base)
    }

    // MARK: - 媒体纯函数

    /// 解密封面；返回图片数据（base64 由 Core JSON 通道还原）。
    public func decryptCover(_ data: Data) async throws -> Data {
        let b64 = try await call("decryptCover", data.base64EncodedString()) as String
        return Data(base64Encoded: b64) ?? Data()
    }
    public func rewritePlaylist(_ upstream: String, _ scheme: String, _ host: String, _ text: String) async throws -> String {
        try await call("rewritePlaylist", upstream, scheme, host, text)
    }
    public func fetchSegment(_ url: String, _ headers: [String: String]) async throws -> SegmentResult {
        try await call("fetchSegment", url, headers)
    }

    // MARK: - 内部

    private func call<T: Decodable>(_ method: String, _ args: any Encodable...) async throws -> T {
        try await invoke(method, try encodeArgs(args))
    }

    private func callVoid(_ method: String, _ args: any Encodable...) async throws {
        let _: Nothing = try await invoke(method, try encodeArgs(args))
    }

    /// 登录/注册/登出/SetBase：callback 成功后先写持久化，写入成功才 resolve。
    private func callPersist<T: Decodable>(_ method: String, _ args: any Encodable...) async throws -> T {
        try await invokePersist(method, try encodeArgs(args))
    }

    private func callVoidPersist(_ method: String, _ args: any Encodable...) async throws {
        let _: Nothing = try await invokePersist(method, try encodeArgs(args))
    }

    // 单一核心：callback 成功时先写 token 持久化，写入成功才 resolve（硬错误不静默）。
    private func invokePersist<T: Decodable>(_ method: String, _ raw: String) async throws -> T {
        try await withCheckedThrowingContinuation { cont in
            let box = CancelBox()
            let ref = engine.start(method, raw) { [weak self] json in
                let result: Result<T, Error> = decodeEnvelope(json)
                if case .success = result, let store = self?.tokenStore {
                    do {
                        guard let self else { throw HuangguoError(code: .unknown, message: "client deallocated") }
                        try store.save(self.engine.getTokenJSON())
                    } catch {
                        cont.resume(throwing: error)
                        box.set({})
                        return
                    }
                }
                box.set({})
                cont.resume(with: result)
            }
            box.set(ref)
        }
    }

    private func invoke<T: Decodable>(_ method: String, _ raw: String) async throws -> T {
        try await withCheckedThrowingContinuation { cont in
            let box = CancelBox()
            let ref = engine.start(method, raw) { json in
                box.set({})
                cont.resume(with: decodeEnvelope(json))
            }
            box.set(ref)
        }
    }

    private func encodeArgs(_ args: [any Encodable]) throws -> String {
        var array: [Data] = []
        for a in args { array.append(try JSONEncoder().encode(anyEncodable(a))) }
        let joined = array.map { String(decoding: $0, as: UTF8.self) }.joined(separator: ",")
        return "[\(joined)]"
    }
}

private struct anyEncodable: Encodable {
    let value: any Encodable
    init(_ v: any Encodable) { value = v }
    func encode(to encoder: Encoder) throws { try value.encode(to: encoder) }
}

private func decodeEnvelope<T: Decodable>(_ json: String) -> Result<T, Error> {
    guard let data = json.data(using: .utf8),
          let env = try? JSONDecoder().decode(Envelope<T>.self, from: data) else {
        return .failure(HuangguoError(code: .unknown, message: json))
    }
    if let e = env.error { return .failure(e) }
    guard let result = env.result else {
        return .failure(HuangguoError(code: .unknown, message: "missing result"))
    }
    return .success(result)
}

/// 媒体分片结果（与 Go SegmentResult 对齐；body 为 base64 字符串）。
public struct SegmentResult: Decodable, Sendable {
    public var status: Int
    public var headers: [String: String]
    public var body: String
}
