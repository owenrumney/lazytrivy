.PHONY: quality
quality:
	which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.48.0
	golangci-lint run

.PHONY: test
test:
	go test -v ./...

.PHONY: run
run:
	go run .