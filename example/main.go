package main

import (
	"context"
	"log"
	"net/http"
	"time"

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

// grpcClient creates a new gRPC client that will be passed in order to initialize the Aperture client.
func grpcClient(ctx context.Context, address string) (*grpc.ClientConn, error) {
	// creating a grpc client connection is essential to allow the Aperture client to communicate with the Flow Control Service.
	var grpcDialOptions []grpc.DialOption
	grpcDialOptions = append(grpcDialOptions, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 10,
	}))
	grpcDialOptions = append(grpcDialOptions, grpc.WithUserAgent("aperture-go"))
	grpcDialOptions = append(grpcDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

	return grpc.DialContext(ctx, address, grpcDialOptions...)
}

func main() {
	ctx := context.Background()
	client, err := grpcClient(ctx, "aperture-agent.aperture-system.svc.cluster.local:80")
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// checkTimeout is the time that the client will wait for a response from Aperture Agent.
	// if not provided, the default value of 200 milliseconds will be used.
	options := aperture.Options{
		ClientConn:   client,
		CheckTimeout: 200 * time.Millisecond,
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

	mux.HandleFunc("/super-api", a.handleSuperAPI)

	err = a.server.ListenAndServe()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

// handleSuperAPI handles http requests on /super-apis endpoint.
func (a app) handleSuperAPI(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// do some business logic to collect labels
	labels := map[string]string{
		"user": "kenobi",
	}
	// StartFlow performs a flowcontrolv1.Check call to Aperture Agent. It returns a Flow and an error if any.
	flow, err := a.apertureClient.StartFlow(ctx, "awesomeFeature", labels)
	if err != nil {
		log.Printf("Aperture flow control got error. Returned flow defaults to Allowed. flow.Accepted(): %t", flow.Accepted())
	}

	// See whether flow was accepted by Aperture Agent.
	if flow.Accepted() {
		// Simulate work being done
		time.Sleep(5 * time.Second)
		// Need to call End on the Flow in order to provide telemetry to Aperture Agent for completing the control loop. The first argument catpures whether the feature captured by the Flow was successful or resulted in an error. The second argument is error message for further diagnosis.
		_ = flow.End(aperture.OK)
	} else {
		// Flow has been rejected by Aperture Agent.
		_ = flow.End(aperture.Error)
	}
}