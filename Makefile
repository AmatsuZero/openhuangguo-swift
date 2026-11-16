MODULES := bindings/huangguo/sdk bindings/huangguo/core bindings/huangguo/gen-huangguo

.PHONY: fmt vet test gen-check all
fmt:
	@for m in $(MODULES); do (cd $$m && gofmt -w $$(find . -name '*.go')) || exit 1; done
vet:
	@for m in $(MODULES); do (cd $$m && GOWORK=off go vet ./...) || exit 1; done
test:
	@for m in $(MODULES); do (cd $$m && GOWORK=off go test ./...) || exit 1; done
gen-check:
	cd bindings/huangguo/gen-huangguo && GOWORK=off go run . --repo-root ../../..
	git diff --exit-code -- bindings/huangguo/ios/Sources/OpenHuangguo/Types.swift bindings/huangguo/android/src/main/kotlin/dev/openhuangguo/Types.kt
all: fmt vet test gen-check
