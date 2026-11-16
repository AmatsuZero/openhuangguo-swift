import Foundation
import XCTest
@testable import OpenHuangguo

/// 用假 Engine 检查参数编码、结果解码、错误映射与取消，不访问网络。
final class EngineTests: XCTestCase {
    private struct FakeEngine: Engine {
        let reply: @Sendable (String, String) -> String?
        let onCancel: @Sendable () -> Void
        func start(_ method: String, _ args: String, _ done: @escaping @Sendable (String) -> Void) -> @Sendable () -> Void {
            if let out = reply(method, args) { done(out) }
            return {
                self.onCancel()
                done(#"{"error":{"code":"CANCELLED","message":"cancelled"}}"#)
            }
        }
        func setTokenJSON(_ json: String) throws {}
        func getTokenJSON() throws -> String { #"{"origin":"https://huangguoai.com","token":"","generation":0}"# }
    }

    func testFeedEncodesArgsAndDecodesPage() async throws {
        let engine = FakeEngine(reply: { method, args in
            XCTAssertEqual(method, "feed")
            XCTAssertEqual(args, "[1,20]")
            return #"{"result":{"items":[],"pagination":{"page":1,"pages":0,"size":20,"total":100}}}"#
        }, onCancel: {})
        let page = try await OpenHuangguoClient(engine: engine, tokenStore: nil).feed(1, 20)
        XCTAssertEqual(page.pagination.total, 100)
    }

    func testErrorMapping() async throws {
        let engine = FakeEngine(reply: { _, _ in
            #"{"error":{"code":"AUTHENTICATION_REQUIRED","message":"login","retryable":false,"retryAfterMs":0}}"#
        }, onCancel: {})
        do {
            _ = try await OpenHuangguoClient(engine: engine, tokenStore: nil).me()
            XCTFail("no error")
        } catch let e as HuangguoError {
            XCTAssertEqual(e.code, .authenticationRequired)
            XCTAssertEqual(e.message, "login")
        }
    }

    func testCancellingTaskCancelsCall() async throws {
        let cancelled = expectation(description: "cancelled")
        let engine = FakeEngine(reply: { _, _ in nil }, onCancel: { cancelled.fulfill() })
        let task = Task { try await OpenHuangguoClient(engine: engine, tokenStore: nil).rank("hot") }
        try await Task.sleep(nanoseconds: 50_000_000)
        task.cancel()
        await fulfillment(of: [cancelled], timeout: 5)
    }

    func testSegmentResultDecodesBase64Body() throws {
        let json = #"{"status":206,"headers":{"Content-Range":"bytes 0-4/100"},"body":"AAE="}"#
        let r = try JSONDecoder().decode(SegmentResult.self, from: Data(json.utf8))
        XCTAssertEqual(r.status, 206)
        XCTAssertEqual(r.headers["Content-Range"], "bytes 0-4/100")
        XCTAssertEqual(r.body, "AAE=")
    }
}
