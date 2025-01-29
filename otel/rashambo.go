package otel

import (
	"context"

	"github.com/petenilson/roshambo"
	"go.opentelemetry.io/otel/metric"
)

var _ roshambo.RashamboTelemetry = (*RoshamboMetrics)(nil)

// RoshamboMetrics handles the logic for
// metering the very important business metrics.
type RoshamboMetrics struct {
	wins  metric.Int64Counter
	loses metric.Int64Counter
	draws metric.Int64Counter
}

func NewRoshamboMetrics(provider metric.MeterProvider) (*RoshamboMetrics, error) {
	meter := provider.Meter("ShootMeter")
	var err error
	r := &RoshamboMetrics{}

	if r.wins, err = meter.Int64Counter(
		"user_wins",
		metric.WithDescription("amount of time the user has won"),
	); err != nil {
		return nil, err
	}

	if r.loses, err = meter.Int64Counter(
		"user_loses",
		metric.WithDescription("amount of time the user has lost"),
	); err != nil {
		return nil, err
	}

	if r.draws, err = meter.Int64Counter(
		"user_draws",
		metric.WithDescription("amount of time the user has drawn"),
	); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *RoshamboMetrics) RecordResult(ctx context.Context, result roshambo.Result) {
	switch result {
	case roshambo.DRAW:
		r.draws.Add(ctx, 1)
	case roshambo.WIN:
		r.wins.Add(ctx, 1)
	case roshambo.LOSE:
		r.loses.Add(ctx, 1)
	}
}
