.PHONY: help frontend datasystem functionsystem runtime_launcher yuanrong dashboard pkg aio all

# Number of parallel jobs for build (can be overridden via environment variable)
# functionsystem uses half of CPU cores because it requires large memory for compilation
JOBS ?= $(shell nproc)
FUNCTIONSYSTEM_JOBS ?= $(shell echo $$(( $$(nproc) / 2 )))
# Remote cache URL for bazel build (leave empty to disable remote cache)
REMOTE_CACHE_URL ?=

help:
	@echo "Available targets:"
	@echo "  make frontend        - Build frontend (auto-fixes go.mod path)"
	@echo "  make datasystem     - Build datasystem"
	@echo "  make functionsystem - Build functionsystem"
	@echo "  make runtime_launcher - Build runtime-launcher"
	@echo "  make yuanrong       - Build runtime"
	@echo "  make dashboard      - Build dashboard"
	@echo "  make pkg           - Copy packages to example/aio/pkg/"
	@echo "  make aio           - Build (cd example/aio && docker build)"
	@echo "  make all           - Build all targets and copy outputs to output/"
	@echo "  make clean         - Clean all build outputs"
	@echo ""
	@echo "Environment variables:"
	@echo "  JOBS                - Parallel jobs (default: nproc)"
	@echo "  FUNCTIONSYSTEM_JOBS - Functionsystem jobs (default: nproc/2)"
	@echo "  REMOTE_CACHE_URL    - Bazel remote cache URL (default: empty)"
	@echo "Parameters (optional):"
	@echo "  REMOTE_CACHE       - Remote cache server address"
	@echo "                      Example: make yuanrong REMOTE_CACHE=grpc://192.168.3.45:9092"
	@echo "                      If not provided, build will proceed without remote cache"
	@echo "  JOBS               - Number of parallel jobs for compilation (default: auto/2)"
	@echo "                      Example: make functionsystem JOBS=8"

# Default values
REMOTE_CACHE ?=
NPROCS := $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
JOBS ?= $(shell echo $$(($(NPROCS) / 2)))

frontend:
	@if grep -q 'yuanrong.org/kernel/runtime.*=>.*\.\./yuanrong/api/go' "frontend/go.mod"; then \
		sed -i 's|yuanrong.org/kernel/runtime.*=>.*\.\./yuanrong/api/go|yuanrong.org/kernel/runtime => ../api/go|g' "frontend/go.mod"; \
		echo "Updated frontend/go.mod: yuanrong.org/kernel/runtime => ../api/go"; \
	else \
		echo "frontend/go.mod already correct"; \
	fi
	bash frontend/build.sh
	@mkdir -p output
	@cp frontend/output/yr-frontend*.tar.gz output/

datasystem:
	@echo "Building datasystem..."
	bash datasystem/build.sh -X off -G on -i on
	@mkdir -p output
	@cp datasystem/output/yr-datasystem*.tar.gz output/
	mkdir -p functionsystem/vendor/src
	rm -rf functionsystem/vendor/src/datasystem
	rm -rf functionsystem/vendor/output/Install/datasystem
	cp datasystem/output/yr-datasystem-*.tar.gz functionsystem/vendor/src/yr-datasystem.tar.gz
	[ -d datasystem/output/sdk ] || tar --no-same-owner -zxf datasystem/output/yr-datasystem-*.tar.gz --strip-components=1 -C datasystem/output

runtime_launcher:
	@echo "Building runtime-launcher..."
	@export PATH=/usr/local/go/bin:~/bin:~/go/bin:$$PATH; \
	if ! command -v go >/dev/null 2>&1; then \
		echo "Error: Go not found. Please install Go and add to PATH."; \
		exit 1; \
	fi
	@mkdir -p functionsystem/runtime-launcher/bin
	@echo "Generating protobuf files..."
	@export PATH=/usr/local/go/bin:~/bin:~/go/bin:$$PATH; \
	cd functionsystem/runtime-launcher && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/runtime/v1/runtime_launcher.proto
	@echo "Compiling runtime-launcher..."
	@export PATH=/usr/local/go/bin:~/bin:~/go/bin:$$PATH; \
	cd functionsystem/runtime-launcher && \
	go build -buildvcs=false -o bin/runtime/runtime-launcher ./cmd/runtime-launcher/ && \
	go build -buildvcs=false -o bin/rl-client ./cmd/rl-client/
	@echo "Runtime-launcher built successfully!"

functionsystem:
	cd functionsystem && bash run.sh build -j $(JOBS) && bash run.sh pack && cd -
	cp -ar functionsystem/output/metrics ./
	cp functionsystem/output/yr-functionsystem*.tar.gz output/

yuanrong:
	@echo "Building yuanrong runtime..."
	@if [ -n "$(REMOTE_CACHE_URL)" ]; then \
		bash build.sh -P -r $(REMOTE_CACHE_URL); \
	else \
		bash build.sh -P; \
	fi

dashboard:
	@echo "Building dashboard..."
	cd go && bash build.sh && cd -
	cp go/output/yr-dashboard*.tar.gz output/
	cp go/output/yr-faas*.tar.gz output/

yuanrong:
	@echo "Building runtime..."
	@if [ -n "$(REMOTE_CACHE)" ]; then \
		echo "Using remote cache: $(REMOTE_CACHE)"; \
		bash build.sh -P -r $(REMOTE_CACHE); \
	else \
		echo "Building without remote cache (REMOTE_CACHE not provided)"; \
		bash build.sh -P; \
	fi

aio:
	@echo "Copying packages to example/aio/pkg/..."
	@mkdir -p example/aio/pkg
	@cp datasystem/output/sdk/openyuanrong_datasystem_sdk-*.whl example/aio/pkg/ 2>/dev/null || true
	@cp datasystem/output/openyuanrong_datasystem-*.whl example/aio/pkg/ 2>/dev/null || true
	@cp functionsystem/output/openyuanrong_functionsystem-*.whl example/aio/pkg/ 2>/dev/null || true
	@cp output/openyuanrong-*.whl example/aio/pkg/ 2>/dev/null || true
	@cp output/openyuanrong_sdk-*.whl example/aio/pkg/ 2>/dev/null || true
	@cp functionsystem/runtime-launcher/bin/runtime/runtime-launcher example/aio/pkg/runtime-launcher 2>/dev/null || true
	@mkdir -p example/aio/docs
	@cp example/aio/TRAEFIK_ETCD.md example/aio/docs/ 2>/dev/null || true
	@echo "Packages copied successfully!"
	@ls -la example/aio/pkg/
	@echo "Building Docker image aio:latest..."
	@cd example/aio && docker build -t openyuanrongaio:latest -f Dockerfile . && cd - || (cd -; exit 1)

all: frontend datasystem functionsystem runtime_launcher dashboard yuanrong
	@echo "Build completed!"
	@echo "Copying outputs to output/..."
	@mkdir -p example/aio/pkg
	@cp functionsystem/runtime-launcher/bin/runtime/runtime-launcher example/aio/pkg/runtime-launcher
