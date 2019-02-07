#This makefile is used by ci-operator

CGO_ENABLED=0
GOOS=linux

install:
	go install ./cmd/perfrepoclient
.PHONY: install

run:
	$(GOPATH)/bin/perfrepoclient
.PHONY: run

test-e2e:
	go test -tags=e2e -v -count=1 ./test/e2e
.PHONY: run