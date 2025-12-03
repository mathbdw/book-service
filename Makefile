ifneq ($(wildcard .env),)
include .env
export
else
$(warning WARNING: .env file not found! Using .env.example)
include .env.example
export
endif

MODULE_PATH=github.com/mathbdw/book/proto
SERVICE_PATH=mathbdw/proto-book

.PHONY: run-app
run-app:
	go run cmd/app/main.go

.PHONY: run-publisher
run-publisher:
	go run cmd/publisher/main.go

.PHONY: run-bot
run-bot:
	go run cmd/bot/main.go

.PHONY: proto-go
proto-go:
	protoc \
		-I docs/proto \
 		--go_out=docs/pkg \
 		--go_opt=module=$(MODULE_PATH) \
 		--go-grpc_out=docs/pkg \
 		--go-grpc_opt=module=$(MODULE_PATH) \
 		--validate_out="lang=go,module=$(MODULE_PATH):docs/pkg" \
 		--grpc-gateway_opt=module=$(MODULE_PATH) \
 		--grpc-gateway_out=docs/pkg \
 		--openapiv2_out ./docs/swagger \
		--openapiv2_opt logtostderr=true \
		--openapiv2_opt allow_merge=true \
		--openapiv2_opt merge_file_name=api \
 		docs/proto/v1/*.proto

.PHONY: deps-go
deps-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest


EXCLUDE_DIRS = vendor|docs|mocks
TEST_PACKAGES = $(shell go list ./... | grep -v -E "($(EXCLUDE_DIRS))")

.PHONY: test
test:
	@echo "📦 Testing packages:"
	@echo "$(TEST_PACKAGES)" | tr ' ' '\n'
	@echo ""
	go test -v $(TEST_PACKAGES) -cover -coverprofile=./coverage.out

.PHONY: test-cover
test-cover: test
	go tool cover -html=./coverage.out
	@echo "✅ Coverage report generated: coverage.html"

.PHONY: test-short
test-short:
	go test -v $(TEST_PACKAGES) -short -cover

.PHONY: test-race
test-race:
	go test -v $(TEST_PACKAGES) -race -cover

.PHONY: build-go
build-go: .build

.build:
	go mod download && CGO_ENABLED=0  go build \
		-tags='no_mysql no_sqlite3' \
		-o ./bin/book-service$(shell go env GOEXE) ./cmd/app/main.go
