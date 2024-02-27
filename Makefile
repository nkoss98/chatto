.PHONY: generate
generate:
	go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.14.0 -generate types,chi-server,spec  -package api api/api.yaml > api/api.gen.go

.PHONY: rundb
rundb:
	docker run --name scratch -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -d postgres

.PHONY: add-migration
add-migration:
	goose -dir internal/storage/migrations create $(name) sql

.PHONY: gen-sqlc
gen-sqlc:
	sqlc generate

.PHONY: install-tools
install-tools:
	go install github.com/golang/mock/mockgen@v1.6.0

.PHONY: gen-mocks
gen-mocks:
	mockgen -source internal/storage/database/querier.go -destination internal/storage/database/mock/querier.go -package db

.PHONY: test-coverage
test-coverage:
	go test ./... -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html