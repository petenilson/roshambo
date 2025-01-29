package roshambo

import "context"

var (
	// used as an example, this would be injected during buildtime
	Version     string = "1.23"
	ExporterUri string = "otel-collector:4317"
	Address     string = "0.0.0.0:8080"
)

type Result string

var (
	WIN  Result = "You Won!"
	LOSE Result = "You Lost!"
	DRAW Result = "You Draw!"
)

type Form string

var (
	ROCK     Form = "rock"
	PAPER    Form = "paper"
	SCISSORS Form = "scissors"
)

func New(rt RashamboTelemetry) *Service {
	return &Service{
		telemetry: rt,
	}
}

type Service struct {
	telemetry RashamboTelemetry
}

type Selection struct {
	Form Form
}

// Shoot simulates a round and meters the result.
// We call another private method with the same name to seperate
// our concerns of metrics with that of our business logic.
func (r Service) Shoot(ctx context.Context, selection Selection) Result {
	result := r.shoot(selection)
	r.telemetry.RecordResult(ctx, result)
	return result
}

func (r Service) shoot(selection Selection) Result {
	// TODO: make this actually random
	randomChoice := PAPER
	switch selection.Form {
	case PAPER:
		switch randomChoice {
		case PAPER:
			return DRAW
		case ROCK:
			return WIN
		case SCISSORS:
			return LOSE
		}

	case ROCK:
		switch randomChoice {
		case PAPER:
			return LOSE
		case ROCK:
			return DRAW
		case SCISSORS:
			return WIN
		}

	case SCISSORS:
		switch randomChoice {
		case PAPER:
			return WIN
		case ROCK:
			return LOSE
		case SCISSORS:
			return DRAW
		}
	}
	panic("Bad Choice")
}
