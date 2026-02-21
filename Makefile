.PHONY: docker
.PHONY: docker-down
.PHONY: test-kafka
.PHONY: coverage

# Detect the operating system (use go env - reliable when Go is installed)
DETECTED_GOOS := $(shell go env GOOS)
ifeq ($(DETECTED_GOOS),windows)
	DETECTED_OS := Windows
else
ifeq ($(OS),Windows_NT)
	DETECTED_OS := Windows
else
	DETECTED_OS := $(shell uname -s)
endif
endif
ifeq ($(DETECTED_OS),)
	DETECTED_OS := Windows
endif

# Define the output binary names based on the OS
ifeq ($(DETECTED_OS),Windows)
	BINARY_REST_API := match-making-api-http-service.exe
	BINARY_CONSUMER_MATCHMAKING := consumer-matchmaking-commands.exe
	BINARY_CONSUMER_SERVER_ALLOCATED := consumer-server-allocated.exe
	BINARY_CONSUMER_MATCH_STARTED := consumer-match-started.exe
	BINARY_WORKER_QUEUE_STATUS := worker-queue-status.exe
	BINARY_WORKER_SERVER_ALLOCATION_TIMEOUT := worker-server-allocation-timeout.exe
else
	BINARY_REST_API := match-making-api-http-service
	BINARY_CONSUMER_MATCHMAKING := consumer-matchmaking-commands
	BINARY_CONSUMER_SERVER_ALLOCATED := consumer-server-allocated
	BINARY_CONSUMER_MATCH_STARTED := consumer-match-started
	BINARY_WORKER_QUEUE_STATUS := worker-queue-status
	BINARY_WORKER_SERVER_ALLOCATION_TIMEOUT := worker-server-allocation-timeout
endif

build-rest-api:
	@echo "Building REST API for $(DETECTED_OS)"
ifeq ($(DETECTED_OS),Windows)
	@go build -o $(BINARY_REST_API) ./cmd/rest-api/main.go
else
	CGO_ENABLED=0 go build -o $(BINARY_REST_API) ./cmd/rest-api/main.go
endif

build-consumer-matchmaking:
	@echo "Building Matchmaking Commands Consumer for $(DETECTED_OS)"
ifeq ($(DETECTED_OS),Windows)
	@go build -o $(BINARY_CONSUMER_MATCHMAKING) ./cmd/consumers/matchmaking-commands/main.go
else
	CGO_ENABLED=0 go build -o $(BINARY_CONSUMER_MATCHMAKING) ./cmd/consumers/matchmaking-commands/main.go
endif

build-consumer-server-allocated:
	@echo "Building Server Allocated Consumer for $(DETECTED_OS)"
ifeq ($(DETECTED_OS),Windows)
	@go build -o $(BINARY_CONSUMER_SERVER_ALLOCATED) ./cmd/consumers/server-allocated/main.go
else
	CGO_ENABLED=0 go build -o $(BINARY_CONSUMER_SERVER_ALLOCATED) ./cmd/consumers/server-allocated/main.go
endif

build-consumer-match-started:
	@echo "Building Match Started Consumer for $(DETECTED_OS)"
ifeq ($(DETECTED_OS),Windows)
	@go build -o $(BINARY_CONSUMER_MATCH_STARTED) ./cmd/consumers/match-started/main.go
else
	CGO_ENABLED=0 go build -o $(BINARY_CONSUMER_MATCH_STARTED) ./cmd/consumers/match-started/main.go
endif

build-worker-queue-status:
	@echo "Building Queue Status Worker for $(DETECTED_OS)"
ifeq ($(DETECTED_OS),Windows)
	@go build -o $(BINARY_WORKER_QUEUE_STATUS) ./cmd/workers/queue-status/main.go
else
	CGO_ENABLED=0 go build -o $(BINARY_WORKER_QUEUE_STATUS) ./cmd/workers/queue-status/main.go
endif

build-worker-server-allocation-timeout:
	@echo "Building Server Allocation Timeout Worker for $(DETECTED_OS)"
ifeq ($(DETECTED_OS),Windows)
	@go build -o $(BINARY_WORKER_SERVER_ALLOCATION_TIMEOUT) ./cmd/workers/server-allocation-timeout/main.go
else
	CGO_ENABLED=0 go build -o $(BINARY_WORKER_SERVER_ALLOCATION_TIMEOUT) ./cmd/workers/server-allocation-timeout/main.go
endif

build-all: build-rest-api build-consumer-matchmaking build-consumer-server-allocated build-consumer-match-started build-worker-queue-status build-worker-server-allocation-timeout
	@echo "All binaries built successfully"

start-rest-api:
	@echo "Running REST API"
	@export DEV_ENV="true"
	@./$(BINARY_REST_API)

start-consumer-matchmaking:
	@echo "Running Matchmaking Commands Consumer"
	@export DEV_ENV="true"
	@./$(BINARY_CONSUMER_MATCHMAKING)

start-worker-queue-status:
	@echo "Running Queue Status Worker"
	@export DEV_ENV="true"
	@./$(BINARY_WORKER_QUEUE_STATUS)

test-docker:
	@echo "Running tests"
	@docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

docker:
	@clear
	@printf "$(NEW_BUFFER)"
	@echo $(LOGO)
	@echo "♻️ $(CG)Removing$(CEND) containers and volumes"
	@docker-compose -f docker-compose.dev.yml down -v
	@echo "🔨 $(CC)Building$(CEND) new containers"
	@docker-compose -f docker-compose.dev.yml build
	@echo "🚀 $(CR)⦿ Running$(CEND) containers"
	@docker-compose -f docker-compose.dev.yml up -d

docker-down:
	@clear
	@printf "$(NEW_BUFFER)"
	@echo $(LOGO)
	@echo "♻️ $(CG)Removing$(CEND) containers and volumes"
	@docker-compose -f docker-compose.dev.yml down -v

test-coverage:
	@go test -covermode=atomic -coverprofile=coverage.out ./...
	@mkdir -p ./.coverage  
	@go tool cover -html=coverage.out -o ./.coverage/coverage.html 

test-kafka-produce:
	@go run pkg/infra/events/pub_kafka_poc.go

test-kafka-consume:
	@go run cmd/async-api/main.go

CG = \033[0;32m
CR = \033[0;31m
CEND = \033[0m
CC = \033[0;36m
B = \033[1m
NEW_BUFFER = \033[H\033[2J
LOGO = "\n\t$(CR)⦿ Match Making$(CEND)API\n\n"

# --- Project Configuration ---

PROJECT_NAME	 := match-making-api
LICENSE_FILE	 := LICENSE
IGNORE_DIRS	  := vendor test		 # Directories to exclude from license checks
ALLOWED_LICENSES := Apache-2.0 MIT	  # Specify allowed licenses (comma-separated)

# --- Go Tools ---

GO ?= go
GO_LICENSES ?= go-licenses

# --- Makefile Targets ---

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $1, $2}' $(MAKEFILE_LIST)

.PHONY: licenses-check
licenses-check: ## Check dependencies for allowed licenses
	@$(GO_LICENSES) check $(PROJECT_NAME) --ignore "$(IGNORE_DIRS)" --allow "$(ALLOWED_LICENSES)"

.PHONY: licenses-csv
licenses-csv: ## Generate a CSV report of dependency licenses
	@$(GO_LICENSES) csv $(PROJECT_NAME) --ignore "$(IGNORE_DIRS)" > licenses.csv

.PHONY: licenses-save
licenses-save: ## Save dependency licenses to a directory
	@$(GO_LICENSES) save $(PROJECT_NAME) --ignore "$(IGNORE_DIRS)" --save_path licenses

.PHONY: install-tools
install-tools: ## Install required Go tools (if not already installed)
	@if ! hash $(GO_LICENSES) 2>/dev/null; then \
		$(GO) install github.com/google/go-licenses@latest; \
	fi

# --- Additional Targets (Customize as needed) ---

.PHONY: licenses-report # Example: Generate a custom report 
licenses-report: licenses-csv
	# Process the licenses.csv file (e.g., using a script or another tool) to create a custom report


.PHONY: mocks
mocks: mocks-clean mocks-generate

.PHONY: mocks-clean
mocks-clean:
ifeq ($(DETECTED_OS),Windows)
	@if exist test\mocks rmdir /s /q test\mocks
	@if not exist test mkdir test
	@mkdir test\mocks
else
	@rm -rf test/mocks
	@mkdir -p test/mocks
endif


.PHONY: mocks-generate
mocks-generate:
	@echo "Generating mocks for MongoDB repositories..."
	@mockgen -source="pkg/infra/db/mongodb/game_mode_mongodb.go" -destination="test/mocks/game_mode_repository_mock.go" -package=mocks
	@mockgen -source="pkg/infra/db/mongodb/game_mongodb.go" -destination="test/mocks/game_repository_mock.go" -package=mocks
	@mockgen -source="pkg/infra/db/mongodb/region_mongodb.go" -destination="test/mocks/region_repository_mock.go" -package=mocks
	@echo "Note: Port interface mocks (test/mocks/port_interfaces_testify_mock.go) are manually maintained"
	@echo "      using testify/mock for compatibility with existing tests."
	@echo "      Update them manually when port interfaces change."
	@echo "Mocks generated successfully!"

.PHONY: mocks-test
mocks-test:
	@echo "Testing mock generation..."
	@mockgen -source="./pkg/infra/db/mongodb/game_mongodb.go" -destination="./test/mocks/test_mock.go" -package=mocks
	@echo "Test mock generated successfully!"
ifeq ($(DETECTED_OS),Windows)
	@if exist test\mocks\test_mock.go del test\mocks\test_mock.go
else
	@rm -f test/mocks/test_mock.go
endif

# --- Default Target ---

.DEFAULT_GOAL := help

#--------- server

# Set your image name and container name (replace with your actual values)
IMAGE_NAME := cs2-server
CONTAINER_NAME := my-cs2-server

# Build the Docker image
build:
	docker build -t $(IMAGE_NAME) .

# Run the Docker container
run:
	docker run -d --name $(CONTAINER_NAME) -p 27015:27015/udp -p 27015:27015 $(IMAGE_NAME)

# Stop the Docker container
stop:
	docker stop $(CONTAINER_NAME)

# Remove the Docker container
rm:
	docker rm $(CONTAINER_NAME)

# Rebuild the Docker image and run the container (cleans up previous container if it exists)
rebuild-run: stop rm build run

# Push the Docker image to a registry (e.g., Docker Hub)
push:
	docker push $(IMAGE_NAME)
