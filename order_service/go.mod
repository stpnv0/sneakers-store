module order_service

go 1.24.6

require (
	github.com/golang-migrate/migrate/v4 v4.18.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/jackc/pgx/v5 v5.5.4
	github.com/segmentio/kafka-go v0.4.47
	github.com/stpnv0/protos v0.0.0
	google.golang.org/grpc v1.69.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.16 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
)

replace github.com/stpnv0/protos => ../protos
