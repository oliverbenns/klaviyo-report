openapi:
	mkdir -p ./generated/klaviyo
	rm -rf ./generated/klaviyo/*
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
	oapi-codegen ./api/klaviyo.json > ./generated/klaviyo/klaviyo.go
	make fmt

fmt:
	gofmt -w `find . -name '*.go'`
