package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/petenilson/roshambo"
	"github.com/petenilson/roshambo/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	ln     net.Listener
	server *http.Server
	router *http.ServeMux

	telemetry       ServerTelemetry
	roshamboService *roshambo.Service
}

type ServerTelemetry interface {
	// instruments http server specific telemetry
	CountGreetingsServed(ctx context.Context)
}

func NewServer(ctx context.Context) (*Server, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	conn, err := grpc.NewClient(
		roshambo.ExporterUri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	tProvider, err := otel.NewTracerProvider(ctx, conn)
	if err != nil {
		return nil, err
	}

	// metrics
	mExporter, err := otel.NewMetricExporter(ctx, conn)
	if err != nil {
		return nil, err
	}
	mProvider := otel.NewMeterProvider(otel.NewPushReader(mExporter))

	// these structs could conceivably be set up in our main
	// and pass through as a dependency
	roshamboServiceMetrics, err := otel.NewRoshamboMetrics(mProvider)
	if err != nil {
		return nil, err
	}
	s := &Server{
		server: &http.Server{Addr: roshambo.Address},
		router: http.NewServeMux(),
		roshamboService: roshambo.New(
			roshamboServiceMetrics,
		),
	}

	// Intrument the Server with our Middleware
	s.server.Handler = Middleware(tProvider, mProvider)(s.router)

	// Register Routes
	s.router.HandleFunc("/play", s.play)

	return s, nil
}

func (s *Server) play(w http.ResponseWriter, r *http.Request) {
	// A trace-id is present at this point in our context as the middleware
	// has injected it. If we're calling to an external service then that
	// service will be able to pull the trace out of the context and
	// add further spans that capture the telemetry that's important to our
	// business.

	var selection roshambo.Selection
	if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
		http.Error(w, "Invalid Choice", http.StatusBadRequest)
		return
	}
	result := s.roshamboService.Shoot(r.Context(), selection)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) Open() (err error) {
	if s.ln, err = net.Listen("tcp", roshambo.Address); err != nil {
		return err
	}

	go s.server.Serve(s.ln)

	return nil
}

func (s *Server) Close() error {
	return s.server.Shutdown(context.Background())
}

// Middleware provides instrumented middleware for our http server.
func Middleware(tp trace.TracerProvider, mp metric.MeterProvider) func(http.Handler) http.Handler {
	// spanNameFormattter allows each endpoint to be instrumented under
	// the operation name that shares it's name with the path.
	// ex. HTTP GET "/greeting"
	spanNameFormattter := func(operation string, r *http.Request) string {
		return fmt.Sprintf("%s %s %s", operation, r.Method, r.URL.Path)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// exclude health checks from intrunmentation here if you'd like
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			h := otelhttp.NewHandler(next, "HTTP", []otelhttp.Option{
				otelhttp.WithTracerProvider(tp),
				otelhttp.WithMeterProvider(mp),
				otelhttp.WithSpanNameFormatter(spanNameFormattter),
				otelhttp.WithMetricAttributesFn(
					func(r *http.Request) []attribute.KeyValue {
						attrs := []attribute.KeyValue{
							attribute.String("http.path", r.URL.Path),
						}
						return attrs
					},
				),
			}...,
			)
			h.ServeHTTP(w, r)
		})
	}
}
