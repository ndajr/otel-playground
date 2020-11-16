module github.com/neemiasjnr/otel-playground/jaeger-exporter

go 1.14

require (
	github.com/gofiber/fiber v1.14.6
	go.opencensus.io v0.22.5
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.13.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.13.0
	go.opentelemetry.io/otel v0.13.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.13.0
	go.opentelemetry.io/otel/sdk v0.13.0
	google.golang.org/grpc v1.33.2
)
