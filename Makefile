.PHONY: lint
lint:
	go fmt ./...
	go vet ./...

.PHONY: build
build: lint
	go build -v -o ${GOPATH}/bin/md-link-lint

.PHONY: run
run:
	go run ./main.go
