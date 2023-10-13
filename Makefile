.PHONY: test coverage

test:
	@go test ./...

coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

clean:
	@rm -f coverage.out