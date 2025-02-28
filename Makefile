
.PHONY: build run clean test

BINARY_NAME=bin/url-shortener

build:
	@go build -o $(BINARY_NAME) cmd/main.go

run: build
	@./$(BINARY_NAME)

clean:
	@rm -f $(BINARY_NAME)

test:
	@go test -v ./...

migrate:
	migrate -path ./db/migrations -database "postgres://postgres:''@localhost:5432/url_shortner_go?sslmode=disable" up