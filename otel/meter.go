package otel

import (
	"context"

	"github.com/petenilson/roshambo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
)

func NewMetricExporter(ctx context.Context, conn *grpc.ClientConn) (*otlpmetricgrpc.Exporter, error) {
	return otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
	)
}

func NewPushReader(exporter metric.Exporter) *metric.PeriodicReader {
	return metric.NewPeriodicReader(
		exporter,
		// sdkmetric.WithInterval() is set to 1m by default and can be configured via OTEL_METRIC_EXPORT_INTERVAL
		// sdkmetric.WithTimeout() is set to 30s by default and can be configured via OTEL_METRIC_EXPORT_TIMEOUT
	)
}

func NewMeterProvider(reader metric.Reader) *metric.MeterProvider {
	return metric.NewMeterProvider(
		metric.WithResource( // if not resource is given, resource.Default() would be called
			resource.NewWithAttributes(
				semconv.SchemaURL,
				attribute.String("service.version", roshambo.Version),
				// add other attribute here
				// attribute.String("version", "production"),
				// attribute.String("region", "us-west-2"),
			)),
		metric.WithReader(reader), // if no reader is given, no metrics are exported
	)
}
