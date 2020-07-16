vendor: go.mod go.sum
	go mod vendor

terraform-provider-googlesiteverification: vendor $(wildcard *.go)
	go build .

.PHONY: test
test: vendor $(wildcard *.go)
	TF_ACC=true go test -v .
