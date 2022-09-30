GO111MODULE=on

default: reload

reload:
	@echo Build and replace terraform-provider-indykite
	@go build -o terraform-provider-indykite$$(go env GOEXE)
	@mkdir -p ./example/terraform.d/plugins/terraform.indykite.com/indykite/indykite/0.1.0/$$(go env GOHOSTOS)_$$(go env GOHOSTARCH)/
	@cp terraform-provider-indykite$$(go env GOEXE) ./example/terraform.d/plugins/terraform.indykite.com/indykite/indykite/0.1.0/$$(go env GOHOSTOS)_$$(go env GOHOSTARCH)
	@rm -f ./example/.terraform.lock.hcl
	@cd ./example/ && terraform init -backend=false

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

download:
	@echo Download go.mod dependencies
	@go mod download

install-tools: download
	@echo Installing tools from tools.go
	@go install $$(go list -f '{{range .Imports}}{{.}} {{end}}' tools.go)

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
