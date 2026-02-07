default: build

build:
	go build -o terraform-provider-shoehorn

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/shoehorn/shoehorn/0.0.1/$(shell go env GOOS)_$(shell go env GOARCH)
	cp terraform-provider-shoehorn ~/.terraform.d/plugins/registry.terraform.io/shoehorn/shoehorn/0.0.1/$(shell go env GOOS)_$(shell go env GOARCH)/

test:
	go test ./... -v -count=1

testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 120m

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet
	go test ./... -count=1

.PHONY: build install test testacc fmt vet lint
