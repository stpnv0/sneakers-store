module fav_service

go 1.24.6

replace github.com/stpnv0/protos => ../protos

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/lib/pq v1.10.9
	github.com/stpnv0/protos v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.69.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/protobuf v1.36.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
