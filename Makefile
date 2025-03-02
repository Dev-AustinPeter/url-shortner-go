
.PHONY: build run clean test testCoverage testCoverageHtmlNoJson

BINARY_NAME=bin/url-shortener

current_dir:=$(shell pwd)
test_output:=$(current_dir)/tests/_output

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

testCoverage:
	go test "./..." -coverprofile="$(test_output)/coverage.out" -covermode=count -json > $(test_output)/report.json || true

testCoverageHtmlNoJson:
	mkdir -p $(test_output) # ✅ Ensure directory exists
	go test -p 1 -tags unit,integration `go list ./... | grep -v node_modules` -v 2>&1 | go-junit-report > $(test_output)/report-download.xml
	go test -p 1 -tags unit,integration -coverprofile=$(test_output)/cover.txt `go list ./... | grep -v node_modules` -v 2>&1 || exit 1
	sleep 2 
	test -f $(test_output)/cover.txt || { echo "Error: cover.txt not found!"; exit 1; } # ✅ Check if cover.txt exists
	go tool cover --html=tests/_output/cover.txt -o tests/_output/coverage/index.html
