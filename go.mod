module github.com/Nerzal/gocloak/v13

go 1.23.0

require (
	github.com/go-resty/resty/v2 v2.16.5
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/segmentio/ksuid v1.0.4
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.36.0
	golang.org/x/mod v0.24.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Nerzal/gocloak/v13 => github.com/darashenka/gocloak/v13 v13.9.1-0.20250708174939-75e7d64bd7ac
