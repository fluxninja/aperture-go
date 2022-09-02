package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"

	aperture "github.com/fluxninja/aperture-go/sdk"
)

// app struct contains the server and the Aperture client.
type app struct {
	server         *http.Server
	apertureClient aperture.Client
}

func main() {
	ctx := context.Background()
	client, err := grpcClient(ctx, "localhost:50051")
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Client can set a tracer provider for their purposes.
	setExporterAndTracerProvider()

	options := aperture.Options{
		ClientConn:   client,
		CheckTimeout: 200 * time.Millisecond,
		Ctx:          ctx,
	}

	// initialize Aperture Client and pass the config that needs to be loaded.
	apertureClient, err := aperture.NewClient(options)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Create a server with passing it the Aperture client
	mux := http.NewServeMux()
	a := app{
		server: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
		apertureClient: apertureClient,
	}

	mux.HandleFunc("/feature", a.handleFeature)

	err = a.server.ListenAndServe()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

// handleFeature is a handler function where all the work from the user is executed.
func (a app) handleFeature(w http.ResponseWriter, r *http.Request) {
	// keep empty when connection is successful, otherwise implement with error message description.
	ctx := context.Background()

	// do some business logic to collect labels
	labels := map[string]string{
		"user": "kenobi",
	}

	// BeginFlow performs a flowcontrolv1.Check call to Aperture Agent. It returns a Flow and an error if any.
	flow, err := a.apertureClient.BeginFlow(ctx, "awesomeFeature", labels)
	if err != nil {
		log.Printf("Aperture flow control got error. Returned flow defaults to Allowed. flow.Accepted(): %t", flow.Accepted())
	}

	// See whether flow was accepted by Aperture Agent.
	if flow.Accepted() {
		// Simulate work being done
		time.Sleep(5 * time.Second)
	} else {
		// Flow has been rejected by Aperture Agent, return appropriate response to caller of this feature
		log.Printf("Flow rejected by Aperture Agent.")
	}

	// Need to call End on the Flow in order to provide telemetry to Aperture Agent for completing the control loop. The first argument catpures whether the feature captured by the Flow was successful or resulted an error. The second argument is error message for further diagnosis.
	flow.End(aperture.OK, "")
}

func setExporterAndTracerProvider() {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("exporter setup failed")
		return
	}
	resourceDefault := resource.Default()
	r, err := resource.Merge(
		resourceDefault,
		resource.NewWithAttributes(
			resourceDefault.SchemaURL(),
			semconv.ServiceNameKey.String("aperture-library-test-app"),
			semconv.ServiceVersionKey.String("v0.1.0"),
		),
	)
	if err != nil {
		log.Fatalf("resource setup failed")
		return
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(r),
	)

	otel.SetTracerProvider(tp)
}

// grpcClient creates a new gRPC client that will be passed in order to initialize the Aperture client.
func grpcClient(ctx context.Context, address string) (*grpc.ClientConn, error) {
	var grpcDialOptions []grpc.DialOption
	grpcDialOptions = append(grpcDialOptions, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 10,
	}))
	grpcDialOptions = append(grpcDialOptions, grpc.WithUserAgent(aperture.LibraryName))
	grpcDialOptions = append(grpcDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

	return grpc.DialContext(ctx, address, grpcDialOptions...)
}
