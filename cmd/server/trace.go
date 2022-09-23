package main

import (
	"github.com/nobugtodebug/ucloud-kratos-layout/internal/conf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

/// Set global trace provider
func SetTracerProvider(c *conf.Tracing, serviceName, serviceVersion string, env string) error {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost(c.GetHost()),
		jaeger.WithAgentPort(c.GetPort()),
	))
	if err != nil {
		return err
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1.0))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("env", env),
		)),
	)
	otel.SetTracerProvider(tp)

	return nil
}
