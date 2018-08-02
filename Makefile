

.PHONY: test
test:
	go test -timeout 15s ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: deps
deps:
	dep ensure

.PHONY: install
install:
	go install github.com/windmilleng/pets
