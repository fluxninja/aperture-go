package main_test

import (
	"context"
	"log"
	"net/http"
	"time"

	aperture "github.com/fluxninja/aperture-go/sdk"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

// app struct contains the server and the Aperture client.
type app struct {
	server         *http.Server
	apertureClient aperture.Client
}

// This is an example of how the Aperture client can be used in a Go application. However, multiple ways of using the client are possible.
func Example() {
	ctx := context.Background()
	client, err := grpcClient(ctx, "localhost:50051")
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Client can set a tracer provider for their purposes.
	setExporterAndTracerProvider()

	// checkTimeout is the time that the client will wait for a response from the Flow Control Service.
	// if not provided, the default value value of 200 milliseconds will be used.
	options := aperture.Options{
		ClientConn:   client,
		CheckTimeout: 200 * time.Millisecond,
		Ctx:          ctx,
	}

	// initialize Aperture Client with the provided options.
	apertureClient, err := aperture.NewClient(options)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Create a server with passing it the Aperture client.
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
		// Need to call End on the Flow in order to provide telemetry to Aperture Agent for completing the control loop. The first argument catpures whether the feature captured by the Flow was successful or resulted in an error. The second argument is error message for further diagnosis.
		flow.End(aperture.Ok, "")
	} else {
		// Flow has been rejected by Aperture Agent.
		flow.End(aperture.Error, "flow rejected by aperture")
	}
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
	// creating a grpc client connection is essential to allow the Aperture client to communicate with the Flow Control Service.
	var grpcDialOptions []grpc.DialOption
	grpcDialOptions = append(grpcDialOptions, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 10,
	}))
	grpcDialOptions = append(grpcDialOptions, grpc.WithUserAgent(aperture.LibraryName))
	grpcDialOptions = append(grpcDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

	return grpc.DialContext(ctx, address, grpcDialOptions...)
}