BINARY_NAME=loadforge
BUILD_DIR=./bin
MAIN_PATH=./cmd/loadforge

.PHONY: build run clean tidy

build:
	@echo "Building $(BINARY_NAME)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary at $(BUILD_DIR)/$(BIANRY_NAME)"

run:
	go run $(MAIN_PATH) $(ARGS)

tidy:
	go mod tidy

clean:
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned."

