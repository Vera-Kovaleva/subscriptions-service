.PHONY: debug
debug: docker-check
	@docker compose --profile dev down && docker system prune --volumes --force && docker compose --profile dev up -d

.PHONY: docker-check
docker-check:
	@if ! command -v docker &> /dev/null; then \
		echo "Error: Docker is not installed. Please install Docker first."; \
		exit 1; \
	fi

.PHONY: lint
lint:
	@go tool golangci-lint run

.PHONY: test
test:
	@go test -count=1 -covermode=atomic ./... | grep -v "/generated/"

.PHONY: db-cli
db-cli:
	@PGPASSWORD=password pgcli --host 127.0.0.1 --port 5432 --username postgres

.PHONY: format
format:
	@go tool gofumpt -l -w . && go tool golines -w . && go tool goimports -w -local "effective_mobile_project/" .

.PHONY: codegen
codegen:
	@go tool oapi-codegen --config=.oapi-codegen.yaml assignment/swagger.yaml \
	&& mockery --log-level="" && rm -rf internal/generated/mocks && mkdir internal/generated/mocks && mockery --log-level=""
