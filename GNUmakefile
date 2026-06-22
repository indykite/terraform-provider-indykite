GO111MODULE=on
GO_CPU_TEST ?= $(shell getconf _NPROCESSORS_ONLN) # use all available CPU cores for testing by default

# tooling choice defaults
PKG_MANAGER ?= brew
VALID_PKG_MANAGERS := brew native
PRE_COMMIT_BIN ?= prek
VALID_PRE_COMMIT_BINS := pre-commit prek
TERRAFORM_BIN ?= terraform
VALID_TERRAFORM_BINS := terraform tofu

# Optional local overrides. Create .env.local.mk for machine-specific settings
# and keep it out of version control.
-include .env.local.mk

# supplementary variables & validations
GOEXE := $(shell go env GOEXE)
GOHOSTOS := $(shell go env GOHOSTOS)
GOHOSTARCH := $(shell go env GOHOSTARCH)
TESTS_PROVIDER_DIR := ./tests/provider
TESTS_PROVIDER_TERRAFORM := $(TERRAFORM_BIN) -chdir=$(TESTS_PROVIDER_DIR)
TESTS_PROVIDER_PLUGIN_DIR := $(TESTS_PROVIDER_DIR)/terraform.d/plugins/registry.terraform.io/indykite/indykite/0.0.1
TESTS_PROVIDER_PLUGIN_ARCH_DIR := $(TESTS_PROVIDER_PLUGIN_DIR)/$(GOHOSTOS)_$(GOHOSTARCH)
SELF_MAKEFILE := $(lastword $(MAKEFILE_LIST))

ifneq ($(filter $(PKG_MANAGER),$(VALID_PKG_MANAGERS)),$(PKG_MANAGER))
$(error Invalid PKG_MANAGER '$(PKG_MANAGER)'. Allowed values: $(VALID_PKG_MANAGERS))
endif

ifneq ($(filter $(PRE_COMMIT_BIN),$(VALID_PRE_COMMIT_BINS)),$(PRE_COMMIT_BIN))
$(error Invalid PRE_COMMIT_BIN '$(PRE_COMMIT_BIN)'. Allowed values: $(VALID_PRE_COMMIT_BINS))
endif

ifneq ($(filter $(TERRAFORM_BIN),$(VALID_TERRAFORM_BINS)),$(TERRAFORM_BIN))
$(error Invalid TERRAFORM_BIN '$(TERRAFORM_BIN)'. Allowed values: $(VALID_TERRAFORM_BINS))
endif

TF_HASHICORP_PREFIX := $(shell brew --prefix hashicorp/tap/terraform 2>/dev/null)
TF_HASHICORP_PATH := $(if $(TF_HASHICORP_PREFIX),$(TF_HASHICORP_PREFIX)/bin/terraform,)
TF_ACC_TERRAFORM_PATH ?= $(or $(shell test -x "$(TF_HASHICORP_PATH)" && echo "$(TF_HASHICORP_PATH)"),$(shell command -v $(TERRAFORM_BIN) 2>/dev/null),$(shell command -v terraform 2>/dev/null),$(shell command -v tofu 2>/dev/null))


.PHONY: all clean default fmt goimports gci lint install-tools install-tools-brew install-tools-brew-hashicorp-terraform install-tools-native install-tools-go-only install-pre-commit-hooks upgrade-tools-brew upgrade-tools-native upgrade-tools-go-only test integration test-all upgrade_test_provider upgrade-pre-commit upgrade tidy tfdocs_generate build build_test_plugin tf_init lint-pre-commit lint-golangci lint-makefile

default:

all: clean upgrade fmt lint build tfdocs_generate build_test_plugin test-all

clean:
	go clean -cache -testcache
	@rm -f terraform-provider-indykite$(GOEXE) coverage.out $(TESTS_PROVIDER_DIR)/.terraform.lock.hcl

#
# Prerequisite tooling
#
install-tools: install-tools-$(PKG_MANAGER)
	@echo "==> All tools installed successfully"
	@$(MAKE) install-pre-commit-hooks

install-tools-brew: install-tools-go-only
	@echo "==> Installing brew-managed tools..."
	@command -v brew >/dev/null 2>&1 || (echo "Homebrew is required when PKG_MANAGER=brew"; exit 1)
	@brew install go gci golangci-lint $(PRE_COMMIT_BIN) yamlfmt yamllint actionlint markdownlint-cli2 shellcheck shfmt checkov trivy hadolint tflint opentofu
	@$(MAKE) install-tools-brew-hashicorp-terraform

install-tools-brew-hashicorp-terraform:
	@echo "[WARN] Acceptance tests and 'tfplugindocs' require HashiCorp Terraform CLI. OpenTofu alone is not enough because terraform-plugin-sdk v2 acceptance harness uses legacy provider address behavior."
	@brew tap hashicorp/tap
	@brew install hashicorp/tap/terraform

install-tools-native: install-tools-go-only
	@echo "==> Installing/upgrading native Go tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "==> Installation completed"
	@echo "[INFO] Note, '~/go/bin' must be added to $PATH in '~/.bash_profile' or similar for the tools to be available in the terminal"

# Tools installed via go install in all modes, because they are not available via 'brew'.
install-tools-go-only:
	@echo "==> Installing/upgrading Go-only tools..."
	@go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest

install-pre-commit-hooks:
	@echo "==> Installing pre-commit hooks..."
	$(PRE_COMMIT_BIN) install -f
	$(PRE_COMMIT_BIN) install -f -t post-checkout -t commit-msg

#
# Upgrade dependencies, both in-code and tooling
#
upgrade-tools-native: install-tools-native
upgrade-tools-go-only: install-tools-go-only
upgrade-tools-brew: upgrade-tools-go-only
	@echo "==> Upgrading brew-managed tools..."
	@command -v brew >/dev/null 2>&1 || (echo "Homebrew is required when PKG_MANAGER=brew"; exit 1)
	brew update && brew upgrade -g --yes && brew cleanup

upgrade_test_provider:
	$(TESTS_PROVIDER_TERRAFORM) init -upgrade

upgrade-pre-commit:
	@echo "==> Upgrading pre-commit"
	@$(PRE_COMMIT_BIN) autoupdate --freeze --jobs $$(getconf _NPROCESSORS_ONLN)

upgrade: upgrade-tools-$(PKG_MANAGER) upgrade-pre-commit upgrade_test_provider
	@echo "==> Upgrading Golang dependencies"
	GO111MODULE=on go get -u all
	@$(MAKE) tidy
	@echo "Please, upgrade workflows manually"

#
# Formatting & Linting
#
tidy:
	@GO111MODULE=on go mod tidy

fmt: tidy gci
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w .

goimports: gci

gci:
	@echo "==> Fixing imports code with gci..."
	gci write -s standard -s default -s "prefix(github.com/indykite/terraform-provider-indykite)" -s blank -s dot .

lint-makefile:
	go run github.com/checkmake/checkmake/cmd/checkmake@latest $(SELF_MAKEFILE)

lint-pre-commit:
	@echo "==> Running pre-commit hooks..."
	@$(PRE_COMMIT_BIN) run --all-files

lint-golangci:
	@echo "==> Checking source code against linters..."
	golangci-lint run ./...

lint: lint-golangci # TODO: consider switching to 'lint-pre-commit' instead

#
# Build and documentation generation
#
tfdocs_generate:
	@echo "==> Running tfplugindocs with TF_PLUGIN_CACHE_DIR disabled to avoid local Terraform provider install cache bug"
	@env -u TF_PLUGIN_CACHE_DIR PATH="$(dir $(TF_ACC_TERRAFORM_PATH)):$$PATH" tfplugindocs generate --rendered-provider-name "IndyKite"

build:
	go build -o terraform-provider-indykite$(GOEXE)

build_test_plugin: build
	@echo "Build local Terraform provider plugin and store to $(TESTS_PROVIDER_PLUGIN_ARCH_DIR)"
	@mkdir -p $(TESTS_PROVIDER_PLUGIN_ARCH_DIR)
	@cp terraform-provider-indykite$(GOEXE) $(TESTS_PROVIDER_PLUGIN_ARCH_DIR)/
	@rm -f $(TESTS_PROVIDER_DIR)/.terraform.lock.hcl
	@$(MAKE) tf_init

#
# Unit & Integration tests
#
test-all: test integration

test:
	GO_CPU_TEST="$(GO_CPU_TEST)" bash tests/test_unit.sh

tf_init:
	$(TESTS_PROVIDER_TERRAFORM) init -backend=false

integration:
	bash tests/test_integration.sh
