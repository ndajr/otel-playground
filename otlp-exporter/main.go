package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

// https://github.com/open-telemetry/opentelemetry-go/blob/master/example/otel-collector/main.go#L44
func initProvider() func() {
	ctx := context.Background()
	fmt.Println("Initializing provider..")
	exporter, err := otlp.NewExporter(
		otlp.WithInsecure(),
		otlp.WithAddress("localhost:30080"),
		otlp.WithGRPCDialOption(grpc.WithBlock()),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("otlp-exporter-app"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return func() {
		tracerProvider.Shutdown(ctx)
		exporter.Shutdown(ctx)
	}
}

func main() {
	shutdown := initProvider()
	defer shutdown()
	fmt.Println("Starting app..")

	tr := otel.Tracer("otlp-exporter-app")

	props := otelhttptrace.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	err := func(ctx context.Context) error {
		ctx, span := tr.Start(ctx, "send event")
		defer span.End()

		ctx = baggage.ContextWithValues(ctx, label.String("env", "dev"))
		req, err := http.NewRequestWithContext(ctx, "GET", "https://github.com", nil)
		if err != nil {
			panic(err)
		}
		otelhttptrace.Inject(ctx, req, props)

		fmt.Println("Sending request...")
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		return nil
	}(context.Background())
	if err != nil {
		panic("unexpected error: " + err.Error())
	}
}
