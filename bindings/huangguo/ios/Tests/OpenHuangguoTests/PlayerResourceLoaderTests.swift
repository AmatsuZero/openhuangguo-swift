import AVFoundation
import Foundation
import XCTest
@testable import OpenHuangguo

final class PlayerResourceLoaderTests: XCTestCase {
    private struct MediaEngine: Engine {
        let respond: @Sendable (String, String) -> String
        func start(_ method: String, _ args: String, _ done: @escaping @Sendable (String) -> Void) -> @Sendable () -> Void {
            done(respond(method, args))
            return {}
        }
        func setTokenJSON(_ json: String) throws {}
        func getTokenJSON() throws -> String { "{}" }
    }

    func testSignedResourcesUseHeadersOnlyOnPlaybackOrigin() async throws {
        let engine = MediaEngine { method, raw in
            XCTAssertEqual(method, "fetchSegment")
            let args = try! JSONSerialization.jsonObject(with: Data(raw.utf8)) as! [Any]
            let url = args[0] as! String
            let headers = args[1] as! [String: String]
            XCTAssertEqual(headers["Range"], "bytes=10-12")
            if url.hasPrefix("https://other.example") {
                XCTAssertNil(headers["X-Ticket"])
            } else {
                XCTAssertEqual(headers["X-Ticket"], "signed")
            }
            return #"{"result":{"status":206,"headers":{"Content-Range":"bytes 10-12/100","Content-Type":"video/mp2t"},"body":"YWJj"}}"#
        }
        let play = PlayInfo(videoUrl: "https://CDN.example:443/master.m3u8", mediaHeaders: ["X-Ticket": "signed"])
        let loader = PlayerResourceLoader(client: OpenHuangguoClient(engine: engine, tokenStore: nil), playback: play)
        for url in ["https://cdn.example/seg.ts", "https://cdn.example/key.bin", "https://other.example/seg.ts"] {
            let result = try await loader.load(URL(string: url)!, offset: 10, length: 3)
            XCTAssertEqual(result.body, Data("abc".utf8))
            XCTAssertEqual(result.contentLength, 100)
        }
    }

    func testIgnoredRangeSlicesFullBodyOnce() async throws {
        let engine = MediaEngine { _, _ in
            #"{"result":{"status":200,"headers":{"Content-Type":"video/mp4","Content-Length":"6"},"body":"YWJjZGVm"}}"#
        }
        let loader = PlayerResourceLoader(client: OpenHuangguoClient(engine: engine, tokenStore: nil))
        let result = try await loader.load(URL(string: "https://cdn.example/video.mp4")!, offset: 2, length: 3)
        XCTAssertEqual(result.body, Data("cde".utf8))
        XCTAssertEqual(result.contentLength, 6)
        XCTAssertEqual(result.contentType, "public.mpeg-4")
    }

    func testPlaylistUsesFinalURLAndRewrittenLength() async throws {
        let engine = MediaEngine { method, raw in
            let args = try! JSONSerialization.jsonObject(with: Data(raw.utf8)) as! [Any]
            if method == "fetchSegment" {
                let headers = args[1] as! [String: String]
                XCTAssertNil(headers["Range"])
                XCTAssertEqual(headers["X-Ticket"], "signed")
                return #"{"result":{"status":200,"url":"https://cdn.example/moved/master.m3u8","headers":{"Content-Length":"8","Content-Type":"application/vnd.apple.mpegurl"},"body":"I0VYVE0zVQo="}}"#
            }
            XCTAssertEqual(method, "rewritePlaylist")
            XCTAssertEqual(args[0] as? String, "https://cdn.example/moved/master.m3u8")
            return ##"{"result":"#EXTM3U\nhg://localhost/seg?u=child\n"}"##
        }
        let play = PlayInfo(videoUrl: "https://cdn.example/master.m3u8", mediaHeaders: ["X-Ticket": "signed"])
        let loader = PlayerResourceLoader(client: OpenHuangguoClient(engine: engine, tokenStore: nil), playback: play)
        let result = try await loader.load(URL(string: play.videoUrl)!, offset: 0, length: 8)
        XCTAssertEqual(result.body, Data("#EXTM3U\n".utf8))
        XCTAssertEqual(result.contentLength, 35)
        XCTAssertEqual(result.contentType, "com.apple.m3u-playlist")
    }

    func testCancelledMediaLoadCancelsCore() async throws {
        let started = expectation(description: "started")
        let cancelled = expectation(description: "cancelled")
        struct PendingEngine: Engine {
            let started: XCTestExpectation
            let cancelled: XCTestExpectation
            func start(_ method: String, _ args: String, _ done: @escaping @Sendable (String) -> Void) -> @Sendable () -> Void {
                started.fulfill()
                return {
                    cancelled.fulfill()
                    done(#"{"error":{"code":"CANCELLED","message":"cancelled","retryable":false,"retryAfterMs":0}}"#)
                }
            }
            func setTokenJSON(_ json: String) throws {}
            func getTokenJSON() throws -> String { "{}" }
        }
        let loader = PlayerResourceLoader(client: OpenHuangguoClient(engine: PendingEngine(started: started, cancelled: cancelled), tokenStore: nil))
        let task = Task { try await loader.load(URL(string: "https://cdn.example/video.mp4")!, offset: 0, length: nil) }
        await fulfillment(of: [started], timeout: 2)
        task.cancel()
        await fulfillment(of: [cancelled], timeout: 2)
        do { _ = try await task.value; XCTFail("cancelled load succeeded") }
        catch { XCTAssertTrue(error is CancellationError) }
    }

    func testLocalURLIsDecodedExactlyOnceAndDelegateIsInstalled() throws {
        let client = OpenHuangguoClient(engine: MediaEngine { _, _ in "{}" }, tokenStore: nil)
        let play = PlayInfo(videoUrl: "https://cdn.example/video%2Fpart.m3u8?sig=a%2Bb")
        let loader = PlayerResourceLoader(client: client, playback: play)
        let asset = try loader.asset()
        XCTAssertTrue(asset.resourceLoader.delegate === loader)
        XCTAssertEqual(PlayerResourceLoader.upstreamURL(from: asset.url, scheme: "hg", host: "localhost")?.absoluteString, play.videoUrl)
    }

    func testUnsupportedManifestMethodFailsBeforeNetwork() async throws {
        let loader = PlayerResourceLoader(
            client: OpenHuangguoClient(engine: MediaEngine { _, _ in XCTFail("unexpected network request"); return "{}" }, tokenStore: nil),
            playback: PlayInfo(videoUrl: "https://cdn.example/master.m3u8", manifestMethod: "POST"))
        do {
            _ = try await loader.load(URL(string: "https://cdn.example/master.m3u8")!, offset: 0, length: nil)
            XCTFail("unsupported method silently accepted")
        } catch { XCTAssertTrue(error.localizedDescription.contains("POST")) }
    }
}
