package roshambo

import "context"

// RashamboTelemetry is RoshamboService specific RashamboTelemetry that our
// business finds very important. As Telemetry is required by the business
// it lives here in the root with our other core domain type with our other
// core domain types.
type RashamboTelemetry interface {
	RecordResult(ctx context.Context, res Result)
}

// ServerTelemetry could record any metrics about the server
// that our business finds valuable.
type ServerTelemetry interface {
	// Random example
	RecordValidationError(ctx context.Context, err error)
}
