module github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/core

go 1.26.0

require github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk v0.0.0-00010101000000-000000000000

require (
	golang.org/x/mobile v0.0.0-20261011034528-197770444867 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
)

replace github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk => ../sdk

tool golang.org/x/mobile/cmd/gobind
