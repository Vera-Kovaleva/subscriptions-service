.PHONY: lint
lint:
	@go tool golangci-lint run

.PHONY: db-cli
db-cli:
	@PGPASSWORD=password pgcli --host 127.0.0.1 --port 5432 --username postgres

.PHONY: format
format:
	@go tool gofumpt -l -w . && go tool golines -w . && go tool goimports -w -local "effective_mobile_project/" .

.PHONY: codegen
codegen:
	@go tool oapi-codegen --config=.oapi-codegen.yaml assignment/swagger.yaml
