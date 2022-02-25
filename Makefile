BUILD_DIR ?= ./build
VERSION = 0.0.0
SERVICE_NAME = authservice
MAIN_FILE= apps/main/main.go

# env
include .env
export $(shell sed 's/=.*//' .env)

# clean
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# build
.PHONY: build
build: clean
	@echo "Building version ${VERSION}"
	@mkdir -p $(BUILD_DIR)
	@go build \
		-ldflags="-X main.Version=${VERSION}" \
		-o "${BUILD_DIR}/${SERVICE_NAME}" \
		${MAIN_FILE}
	@echo "Build placed -> ${BUILD_DIR}/${SERVICE_NAME}"

# run
run:
	@go run \
		-ldflags="-X main.Version=${VERSION}" \
		${MAIN_FILE}