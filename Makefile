test-e2e:
	go test -tags=e2e -v -count=1 ./test/e2e
.PHONY: run