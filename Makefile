.PHONY: change-module help test generate generate-check migrate-up migrate-create run

GOOSE_VERSION := v3.28.0

# Default target
help:
	@echo "make test                        unit + integration test (ต้องมี Docker สำหรับ PostgreSQL/RabbitMQ)"
	@echo "make generate                    generate type จาก api/openapi.yaml (oapi-codegen)"
	@echo "make generate-check              fail ถ้าไฟล์ generate ไม่ตรงกับ contract (CI)"
	@echo "make run ROLE=api                รันด้วย role: api | stream | worker | receiver | sandbox | migrate"
	@echo "make migrate-up                  รัน migration (role migrate) กับ DB ตาม .env"
	@echo "make migrate-create NAME=x       สร้างไฟล์ migration ใหม่ใน migrations/"
	@echo ""
	@echo "Usage: make change-module NEW_MODULE=<new-module-name>"
	@echo ""
	@echo "Example:"
	@echo "  make change-module NEW_MODULE=github.com/your-org/your-service"
	@echo ""
	@echo "This will:"
	@echo "  1. Update go.mod with the new module name"
	@echo "  2. Replace all import statements in .go files"
	@echo "  3. Run 'go mod tidy' to clean up dependencies"

# Change module name across the entire project
change-module:
	@if [ -z "$(NEW_MODULE)" ]; then \
		echo "Error: NEW_MODULE is required"; \
		echo "Usage: make change-module NEW_MODULE=<new-module-name>"; \
		exit 1; \
	fi
	@echo "Changing module from current to: $(NEW_MODULE)"
	@echo ""
	@# Get current module name from go.mod
	@CURRENT_MODULE=$$(grep '^module ' go.mod | awk '{print $$2}'); \
	if [ -z "$$CURRENT_MODULE" ]; then \
		echo "Error: Could not find current module name in go.mod"; \
		exit 1; \
	fi; \
	echo "Current module: $$CURRENT_MODULE"; \
	echo "New module: $(NEW_MODULE)"; \
	echo ""; \
	if [ "$$CURRENT_MODULE" = "$(NEW_MODULE)" ]; then \
		echo "Error: New module name is the same as current module name"; \
		exit 1; \
	fi; \
	echo "Step 1: Updating go.mod..."; \
	sed -i.bak "s|^module $$CURRENT_MODULE|module $(NEW_MODULE)|" go.mod; \
	rm -f go.mod.bak; \
	echo "✓ go.mod updated"; \
	echo ""; \
	echo "Step 2: Replacing import statements in .go files..."; \
	find . -type f -name "*.go" ! -path "./vendor/*" -exec sed -i.bak "s|\"$$CURRENT_MODULE|\"$(NEW_MODULE)|g" {} \; && \
	find . -type f -name "*.go.bak" -delete; \
	echo "✓ All import statements updated"; \
	echo ""; \
	echo "Step 3: Running 'go mod tidy'..."; \
	go mod tidy; \
	echo "✓ Dependencies cleaned up"; \
	echo ""; \
	echo "Successfully changed module from '$$CURRENT_MODULE' to '$(NEW_MODULE)'"


test:
	go test -race ./...

generate:
	go generate ./...

generate-check: generate
	@git diff --exit-code -- delivery/http/gen || { echo "generated code is stale: run make generate and commit"; exit 1; }

ROLE ?= api
run:
	go run . -role=$(ROLE)

migrate-up:
	go run . -role=migrate

migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<name>"; exit 1; fi
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir migrations -s create $(NAME) sql
