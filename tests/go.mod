module github.com/adedaryorh/logistics-platform/tests

go 1.25.6

require (
	github.com/adedaryorh/logistics-platform/pkg/kafka v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/redis v0.0.0
	github.com/lib/pq v1.10.9
	github.com/segmentio/kafka-go v0.4.51
)

replace github.com/adedaryorh/logistics-platform/pkg/kafka => ../pkg/kafka

replace github.com/adedaryorh/logistics-platform/pkg/redis => ../pkg/redis
