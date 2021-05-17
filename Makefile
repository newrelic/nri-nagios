WORKDIR      := $(shell pwd)
NATIVEOS	 := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH	 := $(shell go version | awk -F '[ /]' '{print $$5}')
INTEGRATION  := nagios
GO_FILES     := ./src/
GOFLAGS          = -mod=readonly
BINARY_NAME   = nri-$(INTEGRATION)
GOLANGCI_LINT	 = github.com/golangci/golangci-lint/cmd/golangci-lint

all: build

build: clean validate test compile

build-container:
	docker build -t nri-nagios .

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and coverage file..."
	@rm -rfv bin coverage.xml

validate:
	@printf "=== $(INTEGRATION) === [ validate ]: running golangci-lint & semgrep... "
	@go run  $(GOFLAGS) $(GOLANGCI_LINT) run --verbose
	@[ -f .semgrep.yml ] && semgrep_config=".semgrep.yml" || semgrep_config="p/golang" ; \
	docker run --rm -v "${PWD}:/src:ro" --workdir /src returntocorp/semgrep -c "$$semgrep_config"


compile:
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) $(GO_FILES)


compile: bin/$(BINARY_NAME)

test:
	@echo "=== $(INTEGRATION) === [ test ]: running unit tests..."
	@go test -race ./... -count=1


integration-test:
	@echo "=== $(INTEGRATION) === [ test ]: running integration tests..."
	@chmod 0600 tests/testdata/*
	@docker-compose -f tests/docker-compose.yml pull
	@go test -v -tags=integration ./tests/. || (ret=$$?; docker-compose -f tests/docker-compose.yml down && exit $$ret)
	@docker-compose -f tests/docker-compose.yml down

install: compile
	@echo "=== $(INTEGRATION) === [ install ]: installing bin/$(BINARY_NAME)..."
	@sudo install -D --mode=755 --owner=root --strip $(ROOT)bin/$(BINARY_NAME) $(INTEGRATIONS_DIR)/bin/$(BINARY_NAME)
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-definition.yml $(INTEGRATIONS_DIR)/$(INTEGRATION)-definition.yml
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-config.yml.sample $(CONFIG_DIR)/$(INTEGRATION)-config.yml.sample

# Include thematic Makefiles
include $(CURDIR)/build/ci.mk
include $(CURDIR)/build/release.mk

.PHONY: all build clean tools tools-update deps validate compile test integration-test
