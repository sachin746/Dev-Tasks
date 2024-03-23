build:
	@go build -o bin/bankmanage

run: build
	@./bin/bankmanage

test:
	@go test -v ./...
