.PHONY: proto build test pqclean-lib pq-wasm

# PQClean from pq-wasm (https://github.com/noxy-network/pq-wasm) - fetched at build time
PQ_WASM := $(abspath .)/.pq-wasm
PQCL := $(PQ_WASM)/pqclean
MLKEM := $(PQCL)/crypto_kem/ml-kem-768/clean
COMMON := $(PQCL)/common
BUILD := $(abspath .)/internal/kyber/build
LIBPQCL := $(BUILD)/libpqclean.a

proto:
	mkdir -p grpc/noxy
	protoc --go_out=. --go_opt=module=github.com/noxy-network/go-sdk \
		--go-grpc_out=. --go-grpc_opt=module=github.com/noxy-network/go-sdk \
		-I proto proto/noxy.proto

# Fetch pq-wasm from GitHub (with pqclean submodule) if not present
pq-wasm:
	@if [ ! -d "$(PQ_WASM)/pqclean" ]; then \
		echo "Fetching pq-wasm from https://github.com/noxy-network/pq-wasm..."; \
		rm -rf "$(PQ_WASM)"; \
		git clone --depth 1 --recurse-submodules https://github.com/noxy-network/pq-wasm.git "$(PQ_WASM)"; \
		echo "pq-wasm ready."; \
	else \
		echo "pq-wasm already present."; \
	fi

# Build PQClean static library for CGO (required for Kyber interoperability with Node.js/Rust SDKs)
pqclean-lib: pq-wasm $(LIBPQCL)

# Match Go's target architecture for CGO compatibility
PQCL_ARCH := $(shell go env GOARCH 2>/dev/null || echo "arm64")
PQCL_CFLAGS := -O2 -c -fPIC
ifeq ($(shell uname -s),Darwin)
	ifeq ($(PQCL_ARCH),amd64)
		PQCL_CFLAGS += -arch x86_64
	else
		PQCL_CFLAGS += -arch arm64
	endif
endif

$(LIBPQCL):
	@mkdir -p $(BUILD)
	@echo "Building PQClean ML-KEM 768 for $(PQCL_ARCH)..."
	@cd $(MLKEM) && $(CC) $(PQCL_CFLAGS) -I. -I$(COMMON) \
		kem.c indcpa.c polyvec.c poly.c ntt.c cbd.c reduce.c symmetric-shake.c verify.c
	@cd $(COMMON) && $(CC) $(PQCL_CFLAGS) -I. randombytes.c fips202.c
	@cd $(MLKEM) && ar rcs $(LIBPQCL) *.o
	@cd $(COMMON) && ar rcs $(LIBPQCL) randombytes.o fips202.o
	@rm -f $(MLKEM)/*.o $(COMMON)/randombytes.o $(COMMON)/fips202.o
	@echo "Built $(LIBPQCL)"

build: pqclean-lib proto
	CGO_ENABLED=1 go build ./...

test: build
	go test ./...
