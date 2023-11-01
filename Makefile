.PHONY: test coverage

test:
	@go test ./... -count=1

coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

clean:
	@rm -f coverage.out