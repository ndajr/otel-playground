package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagators"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// https://github.com/open-telemetry/opentelemetry-go/blob/master/example/jaeger/main.go#L31
func initTracer() func() {
	fmt.Println("Initializing provider..")
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "jaeger-exporter-app",
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatal(err)
	}
	return flush
}

func main() {
	flush := initTracer()
	defer flush()
	fmt.Println("Starting app..")

	tr := global.Tracer("jaeger-exporter-app")

	props := otelhttptrace.WithPropagators(otel.NewCompositeTextMapPropagator(propagators.TraceContext{}, propagators.Baggage{}))
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	err := func(ctx context.Context) error {
		ctx, span := tr.Start(ctx, "send event")
		defer span.End()

		ctx = otel.ContextWithBaggageValues(ctx, label.String("env", "dev"))
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
		panic("unexpected error in http request: " + err.Error())
	}
}
