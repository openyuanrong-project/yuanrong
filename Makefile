.PHONY: help frontend datasystem functionsystem yuanrong dashboard all clean

# Number of parallel jobs for build (can be overridden via environment variable)
# functionsystem uses half of CPU cores because it requires large memory for compilation
JOBS ?= $(shell nproc)
FUNCTIONSYSTEM_JOBS ?= $(shell echo $$(( $$(nproc) / 2 )))

help:
	@echo "Available targets:"
	@echo "  make frontend       - Build frontend"
	@echo "  make datasystem    - Build datasystem"
	@echo "  make functionsystem - Build functionsystem"
	@echo "  make yuanrong      - Build runtime"
	@echo "  make dashboard     - Build dashboard"
	@echo "  make all           - Build all targets and copy outputs to output/"
	@echo "  make clean         - Clean all build outputs"
	@echo ""
	@echo "Environment variables:"
	@echo "  JOBS                - Parallel jobs (default: nproc)"
	@echo "  FUNCTIONSYSTEM_JOBS - Functionsystem jobs (default: nproc/2)"

frontend:
	@echo "Building frontend..."
	if grep -q 'yuanrong.org/kernel/runtime.*=>.*\.\./yuanrong/api/go' "frontend/go.mod"; then \
		sed -i 's|yuanrong.org/kernel/runtime.*=>.*\.\./yuanrong/api/go|yuanrong.org/kernel/runtime => ../api/go|g' "frontend/go.mod"; \
		echo "Updated frontend/go.mod: yuanrong.org/kernel/runtime => ../api/go"; \
	else \
		echo "frontend/go.mod already correct"; \
	fi
	bash frontend/build.sh
	mkdir -p output
	cp frontend/output/yr-frontend*.tar.gz output/

datasystem:
	@echo "Building datasystem..."
	bash datasystem/build.sh -X off -G on -i on
	mkdir -p output
	cp datasystem/output/yr-datasystem*.tar.gz output/

functionsystem:
	@echo "Building functionsystem..."
	mkdir -p functionsystem/vendor/src
	rm -rf functionsystem/vendor/src/datasystem
	rm -rf functionsystem/vendor/output/Install/datasystem
	cp datasystem/output/yr-datasystem-*.tar.gz functionsystem/vendor/src/yr-datasystem.tar.gz
	cd functionsystem/ && bash run.sh build -j $(FUNCTIONSYSTEM_JOBS) && bash run.sh pack && cd -
	mkdir -p output
	cp functionsystem/output/yr-functionsystem*.tar.gz output/

yuanrong:
	@echo "Building yuanrong runtime..."
	[ -d datasystem/output/sdk ] || tar --no-same-owner -zxf datasystem/output/yr-datasystem-*.tar.gz --strip-components=1 -C datasystem/output
	cp -ar functionsystem/output/metrics ./
	bash build.sh -l /tmp/bazelcache -P

dashboard:
	@echo "Building dashboard..."
	cd go && bash build.sh && cd -

all: frontend datasystem functionsystem dashboard yuanrong
	@echo "Build completed!"
	@echo "Outputs are in output/"

clean:
	@echo "Cleaning all build outputs..."
	rm -rf output/
	rm -rf frontend/output/
	rm -rf datasystem/output/
	rm -rf datasystem/build/
	rm -rf functionsystem/output/
	rm -rf functionsystem/functionsystem/build
	rm -rf functionsystem/functionsystem/output
	rm -rf functionsystem/common/logs/build
	rm -rf functionsystem/common/logs/output
	rm -rf functionsystem/common/litebus/build
	rm -rf functionsystem/common/litebus/output
	rm -rf functionsystem/common/metrics/build
	rm -rf functionsystem/common/metrics/output
	rm -rf metrics/
	rm -rf go/output/
	@echo "Clean completed!"
