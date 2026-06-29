module github.com/adedaryorh/logistics-platform/pkg/kafka

go 1.25.6

require (
	github.com/adedaryorh/logistics-platform/pkg/observability v0.0.0
	github.com/lib/pq v1.12.3
	github.com/segmentio/kafka-go v0.4.51
	google.golang.org/protobuf v1.36.11
)

replace github.com/adedaryorh/logistics-platform/pkg/observability => ../observability

require (
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	golang.org/x/net v0.55.0 // indirect
)
