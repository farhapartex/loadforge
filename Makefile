BINARY_NAME=loadforge
WEB_BINARY_NAME=loadforge-web
BUILD_DIR=./bin
MAIN_PATH=./cmd/loadforge
WEB_MAIN_PATH=./cmd/web

.PHONY: build build-web run run-web clean tidy

build:
	@echo "Building $(BINARY_NAME)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary at $(BUILD_DIR)/$(BINARY_NAME)"

build-web:
	@echo "Building $(WEB_BINARY_NAME)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(WEB_BINARY_NAME) $(WEB_MAIN_PATH)
	@echo "Binary at $(BUILD_DIR)/$(WEB_BINARY_NAME)"

run:
	go run $(MAIN_PATH) $(ARGS)

run-web:
	go run $(WEB_MAIN_PATH) $(ARGS)

tidy:
	go mod tidy

clean:
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned."

