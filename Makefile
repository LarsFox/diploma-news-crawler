default:
	@go run -race cmd/main.go

lint:
	@golangci-lint run
