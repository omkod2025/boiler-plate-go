.PHONY: check run generate
check:
	go vet ./...
	go test ./...
run:
	go run .
generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config oapi-codegen.yaml ../../spec/contracts/openapi/__SERVICE__.yaml
