package otel

import (
	"context"

	"github.com/petenilson/roshambo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
)

var (
	// Provides attributes present in our middleware
	TraceAttributes []attribute.KeyValue = []attribute.KeyValue{
		semconv.ServiceNameKey.String("GreetingsServer"),
		semconv.ServiceVersionKey.String(roshambo.Version),
		// add other attribute here
		// attribute.String("version", "production"),
		// attribute.String("region", "us-west-2"),
	}
)

func NewTracerProvider(
	ctx context.Context,
	conn *grpc.ClientConn,
) (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithGRPCConn(conn),
	)
	if err != nil {
		return nil, err
	}

	provider := trace.NewTracerProvider(
		trace.WithResource( // if not resource is given, resource.Default() would be called
			resource.NewWithAttributes(
				semconv.SchemaURL,
				TraceAttributes...,
			)),
		trace.WithSampler(trace.AlwaysSample()), // if no sampler is given, AlwaysSample would be used
		trace.WithBatcher(exporter),
	)
	return provider, nil
}
