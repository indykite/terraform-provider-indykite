GO111MODULE=on

default:

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w .

goimports: gci

gci:
	@echo "==> Fixing imports code with gci..."
	gci write -s standard -s default -s "prefix(github.com/indykite/terraform-provider-indykite)" -s blank -s dot .

lint:
	@echo "==> Checking source code against linters..."
	golangci-lint run --timeout 2m0s ./...

install-tools:
	@echo Installing tools
	@go install github.com/daixiang0/gci@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest
	@echo Installation completed

test:
	go test -v -cpu 4 -covermode=count -coverpkg github.com/indykite/terraform-provider-indykite/... -coverprofile=coverage.out ./...

upgrade:
	@echo "==> Upgrading Go"
	@GO111MODULE=on go get -u all && go mod tidy
	@echo "==> Upgrading pre-commit"
	@pre-commit autoupdate
	@echo "Please, upgrade workflows manually"

tidy:
	@GO111MODULE=on go mod tidy

tfdocs_generate:
	tfplugindocs generate --rendered-provider-name "IndyKite"
