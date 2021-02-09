module github.com/hectorj/terraform-provider-googlesiteverification

go 1.14

require (
	github.com/cloudflare/terraform-provider-cloudflare v1.18.2-0.20201126031502-995f63ac2526
	github.com/google/uuid v1.1.2
	github.com/hashicorp/terraform-plugin-docs v0.4.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.16.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.29.0
)

replace google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 => google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c
