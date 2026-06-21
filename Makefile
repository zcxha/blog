lint:
	@files="$$(gofmt -l .)"; test -z "$$files" || (echo "The following Go files need gofmt:"; echo "$$files"; exit 1)
	go vet ./...

test:
	$(MAKE) lint
	go test ./...

.PHONY: lint test
