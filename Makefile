all: darwin

darwin:
	go build -o alreadyread

linux:
	GOOS=linux go build -o alreadyread.linux

linux-docker:
	docker run --rm -v $(PWD)/:/go/src/cirello.io/alreadyread/ \
		-w /go/src/cirello.io/alreadyread/ golang \
		/bin/bash -c 'go build -o alreadyread.linux'

test:
	GOEXPERIMENT=loopvar go test -coverprofile=coverage.out -v ./...
	go tool cover -html=coverage.out -o coverage.html

linters:
	which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
	golangci-lint run --timeout 5m --modules-download-mode=vendor --disable-all \
		-E "errcheck" \
		-E "errname" \
		-E "errorlint" \
		-E "exhaustive" \
		-E "exportloopref" \
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
