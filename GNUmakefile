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

integration:
	cd ./tests/provider && terraform init && terraform plan && terraform apply -input=false -auto-approve
	cd ./tests/terraform && go test --tags=integration ./...
	cd ./tests/provider && terraform destroy -input=false -auto-approve && rm terraform.tfstate terraform.tfstate.backup

upgrade_test_provider:
	cd ./tests/provider && terraform init -upgrade

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

build_test_local_plugin:
	@echo "Build local Terraform provider plugin and store to tests/provider/.terraform folder"
	@go build -o terraform-provider-indykite$$(go env GOEXE)
	@mkdir -p ./tests/provider/terraform.d/plugins/registry.terraform.io/indykite/indykite/0.0.1/$$(go env GOHOSTOS)_$$(go env GOHOSTARCH)/
	@cp terraform-provider-indykite$$(go env GOEXE) ./tests/provider/terraform.d/plugins/registry.terraform.io/indykite/indykite/0.0.1/$$(go env GOHOSTOS)_$$(go env GOHOSTARCH)/
	@echo "Clean up all files that are outdated for Terraform provider"
	@rm -f ./tests/provider/.terraform.lock.hcl
	@cd ./tests/provider && terraform init -backend=false
