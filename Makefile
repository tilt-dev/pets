

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	go vet ./...
