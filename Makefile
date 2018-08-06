
.PHONY: install
install:
	go install github.com/windmilleng/pets

.PHONY: test
test:
	go test -timeout 20s ./...

.PHONY: lint
lint:
	go vet ./...

.PHONY: deps
deps:
	dep ensure
