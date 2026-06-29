module github.com/adedaryorh/logistics-platform/services/payment-service

go 1.25.6

require (
	github.com/adedaryorh/logistics-platform/pkg/clients v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/config v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/errors v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/kafka v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/middleware v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/observability v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/servicehttp v0.0.0
	github.com/adedaryorh/logistics-platform/pkg/validation v0.0.0
	github.com/gin-gonic/gin v1.12.0
)

replace github.com/adedaryorh/logistics-platform/pkg/servicehttp => ../../pkg/servicehttp

replace github.com/adedaryorh/logistics-platform/pkg/clients => ../../pkg/clients

replace github.com/adedaryorh/logistics-platform/pkg/config => ../../pkg/config

replace github.com/adedaryorh/logistics-platform/pkg/errors => ../../pkg/errors

replace github.com/adedaryorh/logistics-platform/pkg/kafka => ../../pkg/kafka

replace github.com/adedaryorh/logistics-platform/pkg/middleware => ../../pkg/middleware

replace github.com/adedaryorh/logistics-platform/pkg/observability => ../../pkg/observability

replace github.com/adedaryorh/logistics-platform/pkg/validation => ../../pkg/validation
