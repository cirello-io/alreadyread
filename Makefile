all: darwin

darwin:
	go build -o alreadyread

linux:
	GOOS=linux go build -o alreadyread.linux

linux-docker:
	docker run --pull=always --rm \
		-v alreadyread-build-cache:/root/.cache/go-build \
		-v alreadyread-pkg-mod:/go/pkg/mod \
		-v $(PWD)/:/go/src/cirello.io/alreadyread/ \
		-w /go/src/cirello.io/alreadyread/ golang:latest \
		/bin/bash -c 'go build -o alreadyread.linux'

test:
	go test -coverprofile=coverage.out -v ./...
	go tool cover -html=coverage.out -o coverage.html

linters:
	go run -mod=readonly github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --disable-all \
		-E "errcheck" \
		-E "errname" \
		-E "errorlint" \
		-E "exhaustive" \
		-E "gci" \
		-E "gocritic" \
		-E "godot" \
		-E "gofmt" \
		-E "goimports" \
		-E "govet" \
		-E "grouper" \
		-E "ineffassign" \
		-E "ireturn" \
		-E "misspell" \
		-E "prealloc" \
		-E "predeclared" \
		-E "revive" \
		-E "staticcheck" \
		-E "thelper" \
		-E "unparam" \
		-E "unused" \
		./...
