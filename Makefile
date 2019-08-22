GO111MODULES := on

all: test coverage lint fmt vet

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -coverprofile=/tmp/coverage.out
	go tool cover -func=/tmp/coverage.out

.PHONY: lint
lint:
	scripts/lint.sh

.PHONY: fmt
fmt:
	find . -type f -name '*.go' | xargs gofmt -w -s

.PHONY: vet
vet:
	go vet

.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 scripts/build.sh

.PHONY: darwin
darwin:
	GOOS=darwin GOARCH=amd64 scripts/build.sh

.PHONY: windows
windows:
	GOOS=windows GOARCH=amd64 scripts/build.sh
