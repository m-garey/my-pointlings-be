.PHONY: run test lint docker-build

run:
	go run ./cmd/server

test:
	go test -v -race -cover ./...

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

docker-build:
	docker build -t pointlings-backend .

docs:
	swag init

mock:
	brew install mockery
	mockery	--all
