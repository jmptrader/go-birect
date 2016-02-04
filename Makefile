GOLINT := $(GOPATH)/bin/golint
GO_PROTOS := internal/wire/*.pb.go
PROTOEASY := $(GOPATH)/bin/protoeasy

test: protos test-without-proto-compilation
test-without-proto-compilation: lint vet
	go test --race -v .
lint:
	golint .
	test -z "$$(golint .)"
vet:
	go vet .

protos: $(GO_PROTOS)
$(GO_PROTOS): protos/*.proto $(PROTOEASY)
	PATH=$(PATH):$(GOPATH)/bin && $(PROTOEASY) --go ./protos --out ./internal/wire
$(PROTOEASY): protoc
	go get go.pedge.io/protoeasy/cmd/protoeasy
	go get github.com/golang/protobuf/protoc-gen-go
$(GOLINT):
	go get github.com/golang/lint/golint
protoc:
	./scripts/install-protoc.sh