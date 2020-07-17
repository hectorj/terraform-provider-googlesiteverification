vendor: go.mod go.sum
	go mod vendor

.PHONY: build
build: terraform-provider-googlesiteverification
terraform-provider-googlesiteverification: vendor $(wildcard *.go)
	CGO_ENABLED=0 go build -o terraform-provider-googlesiteverification .

.PHONY: test
test: vendor $(wildcard *.go)
	TF_ACC=true go test -v .
