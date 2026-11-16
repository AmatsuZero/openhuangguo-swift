#!/bin/sh
# 用 gomobile 编译 OpenHuangguoCore，并放进 iOS / Android 绑定工程。
#
#   ./build.sh android   -> android/build/outputs/aar/huangguo-release.aar
#   ./build.sh ios       -> ios/OpenHuangguoCore.xcframework
#   ./build.sh           -> 以上全部
#
# 需要 gomobile（go install golang.org/x/mobile/cmd/gobind 已由 core/go.mod 的 tool 锁定）。
# Android 还需要 ANDROID_HOME / NDK / JDK 17；iOS 需要 Xcode（macOS 目标 v12）。
set -eu

here=$(cd "$(dirname "$0")" && pwd)
core="$here/core"
out="$here/build"

core_android() {
	[ -f "$out/android/core.jar" ] && return
	: "${ANDROID_HOME:?set ANDROID_HOME to the Android SDK}"
	if [ -z "${ANDROID_NDK_HOME:-}" ]; then
		ANDROID_NDK_HOME=$(ls -d "$ANDROID_HOME"/ndk/* | sort -V | tail -1)
		export ANDROID_NDK_HOME
	fi
	rm -rf "$out/android"
	mkdir -p "$out/android"
	(cd "$core" && GOWORK=off gomobile bind -target android -androidapi 21 -javapkg dev.openhuangguo \
		-trimpath -ldflags="-s -w" -o "$out/android/huangguo-release.aar" .)
	(cd "$out/android" && unzip -q huangguo-release.aar classes.jar 'jni/*' && mv classes.jar core.jar)
}

core_ios() {
	[ -d "$out/OpenHuangguoCore.xcframework" ] && return
	mkdir -p "$out"
	(cd "$core" && GOWORK=off gomobile bind -target ios,iossimulator,macos -iosversion 15 -macosversion 12.0 \
		-trimpath -ldflags="-s -w" -o "$out/OpenHuangguoCore.xcframework" .)
}

put_android() {
	rm -rf "$1/libs/core.jar" "$1/src/main/jniLibs"
	mkdir -p "$1/libs" "$1/src/main/jniLibs"
	cp "$out/android/core.jar" "$1/libs/core.jar"
	cp -R "$out/android/jni/." "$1/src/main/jniLibs/"
}

put_ios() {
	rm -rf "$1/OpenHuangguoCore.xcframework"
	cp -R "$out/OpenHuangguoCore.xcframework" "$1/"
}

android() {
	core_android
	put_android "$here/android"
	(cd "$here/android" && ./gradlew --quiet test assembleRelease)
	mv "$here/android/build/outputs/aar/huangguo-release.aar" \
		"$here/android/build/outputs/aar/openhuangguo-release.aar" 2>/dev/null || true
	echo "built $here/android/build/outputs/aar/openhuangguo-release.aar"
}

ios() {
	core_ios
	put_ios "$here/ios"
	echo "built $here/ios/OpenHuangguoCore.xcframework"
}

rm -rf "$out"
case "${1:-all}" in
android) android ;;
ios) ios ;;
all) android && ios ;;
*)
	echo "usage: $0 [android|ios|all]" >&2
	exit 2
	;;
esac
