module github.com/adedaryorh/logistics-platform/pkg/redis

go 1.25.6

require (
	github.com/adedaryorh/logistics-platform/pkg/config v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/observability v0.0.0
)

replace github.com/adedaryorh/logistics-platform/pkg/config => ../config

replace github.com/adedaryorh/logistics-platform/pkg/observability => ../observability
