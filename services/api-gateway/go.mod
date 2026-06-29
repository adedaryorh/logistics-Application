module github.com/adedaryorh/logistics-platform/services/api-gateway

go 1.25.6

require (
	github.com/adedaryorh/logistics-platform/pkg/auth v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/config v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/errors v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/redis v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/servicehttp v0.0.0
	github.com/gin-gonic/gin v1.12.0
)

replace github.com/adedaryorh/logistics-platform/pkg/servicehttp => ../../pkg/servicehttp

replace github.com/adedaryorh/logistics-platform/pkg/auth => ../../pkg/auth

replace github.com/adedaryorh/logistics-platform/pkg/config => ../../pkg/config

replace github.com/adedaryorh/logistics-platform/pkg/errors => ../../pkg/errors

replace github.com/adedaryorh/logistics-platform/pkg/redis => ../../pkg/redis
