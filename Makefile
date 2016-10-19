# Testing
#########

test: lint vet protos run-tests

ci-test: vet run-tests

run-tests:
	go test --race -v .
lint:
	golint .
	test -z "$$(golint .)"
vet:
	go vet .

# Protobuf compilation
######################

GO_PROTOS := internal/wire/*.pb.go
protos: $(GO_PROTOS)
$(GO_PROTOS): protos/*.proto
	protoeasy --go ./protos --out ./internal/wire

# Dependencies
##############

install-protoeasy: protoc
	go get go.pedge.io/protoeasy/cmd/protoeasy
	go get github.com/golang/protobuf/protoc-gen-go
install-golint:
	go get github.com/golang/lint/golint
install-protoc:
	./scripts/install-protoc.sh
