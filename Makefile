.PHONY: app
app:
	@docker compose up --build -d

.PHONY: down
app-down:
	@docker compose down

.PHONY: db-cli
db-cli:
	@PGPASSWORD=password pgcli --host 127.0.0.1 --port 5432 --username postgres

.PHONY: format
format:
	@echo "Formatting code..."
	@go run mvdan.cc/gofumpt@latest -l -w .
	@go run github.com/segmentio/golines@latest -w .
	@go run golang.org/x/tools/cmd/goimports@latest -w -local "tz_effective_mobile/" .
	@echo "Formatting complete!"

.PHONY: codegen
codegen:
	@go tool oapi-codegen --config=.oapi-codegen.yaml assignment/swagger.yaml
