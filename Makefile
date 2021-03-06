all: examples gen lint test

BINDIR := $(CURDIR)/bin

tmp/mods/go.mod:
	mkdir -p tmp/mods
	cd tmp/mods && \
	GO111MODULE=on go mod init mods

bin/goimports: tmp/mods/go.mod
	cd tmp/mods && \
	GO111MODULE=on GOBIN=$(BINDIR) go get golang.org/x/tools/cmd/goimports

bin/golangci-lint: tmp/mods/go.mod
	cd tmp/mods && \
	GO111MODULE=on GOBIN=$(BINDIR) go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.14.0

bin/godoc2md: $(shell find ./ -name \*.go)
	go build -o bin/godoc2md ./cmd/godoc2md/main.go

bin/goreadme: $(shell find ./ -name \*.go)
	go build -o bin/goreadme ./cmd/goreadme/main.go

bin/goreadme-lint: $(shell find ./ -name \*.go)
	go build -o bin/goreadme-lint ./cmd/goreadme-lint/main.go

gen: bin/goreadme
	go generate ./...

test:
	go test -race ./...

update-testdata:
	go test -tags testdata ./...

fmt: bin/goimports
	bin/goimports -w .

lint: bin/golangci-lint bin/goreadme-lint
	bin/golangci-lint run ./...
	bin/goreadme-lint .

examples: bin/godoc2md
	bin/godoc2md github.com/kr/fs > examples/fs/README.md
	bin/godoc2md github.com/codegangsta/martini > examples/martini/README.md
	bin/godoc2md github.com/gorilla/sessions > examples/sessions/README.md
	bin/godoc2md go/build > examples/build/README.md

.PHONY: examples all lint fmt gen
