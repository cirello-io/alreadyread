all: darwin

darwin:
	go build -o alreadyread ./cmd/alreadyread

linux:
	docker run -ti --rm -v $(PWD)/../:/go/src/cirello.io/ \
		-w /go/src/cirello.io/alreadyread golang \
		/bin/bash -c 'go build -o alreadyread.linux ./cmd/alreadyread'

test:
	go test -v ./...
