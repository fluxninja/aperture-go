package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	aperture "github.com/fluxninja/aperture-go/sdk"
)

const (
	defaultAppPort   = "18080"
	defaultAgentHost = "aperture-agent.aperture-agent.svc.cluster.local"
	defaultAgentPort = "8089"
)

// app struct contains the server and the Aperture client.
type app struct {
	server                  *http.Server
	apertureClient          aperture.Client
	apertureAgentGRPCClient *grpc.ClientConn
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
	agentHost := getEnvOrDefault("FN_AGENT_HOST", defaultAgentHost)
	agentPort := getEnvOrDefault("FN_AGENT_PORT", defaultAgentPort)

	ctx := context.Background()

	apertureAgentGRPCClient, err := grpcClient(ctx, net.JoinHostPort(agentHost, agentPort))
	if err != nil {
		log.Fatalf("failed to create flow control client: %v", err)
	}

	options := aperture.Options{
		ApertureAgentGRPCClientConn: apertureAgentGRPCClient,
		CheckTimeout:                200 * time.Millisecond,
	}

	// initialize Aperture Client with the provided options.
	apertureClient, err := aperture.NewClient(options)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	appPort := getEnvOrDefault("FN_APP_PORT", defaultAppPort)
	// Create a server with passing it the Aperture client.
	mux := http.NewServeMux()
	a := &app{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%s", appPort),
			Handler: mux,
		},
		apertureClient:          apertureClient,
		apertureAgentGRPCClient: apertureAgentGRPCClient,
	}

	mux.HandleFunc("/super", a.SuperHandler)
	mux.HandleFunc("/connected", a.ConnectedHandler)

	err = a.server.ListenAndServe()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

// SuperHandler handles HTTP requests on /super API endpoint.
func (a *app) SuperHandler(w http.ResponseWriter, r *http.Request) {
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
		time.Sleep(2 * time.Second)
		// Need to call End() on the Flow in order to provide telemetry to Aperture Agent for completing the control loop.
		// The first argument captures whether the feature captured by the Flow was successful or resulted in an error.
		// The second argument is error message for further diagnosis.
		_ = flow.End(aperture.OK)
		w.WriteHeader(http.StatusAccepted)
	} else {
		// Flow has been rejected by Aperture Agent.
		_ = flow.End(aperture.Error)
		w.WriteHeader(http.StatusForbidden)
	}
}

// ConnectedHandler handles HTTP requests on /connected API endpoint.
func (a *app) ConnectedHandler(w http.ResponseWriter, r *http.Request) {
	a.apertureAgentGRPCClient.Connect()
	state := a.apertureAgentGRPCClient.GetState()
	_, _ = w.Write([]byte(state.String()))
	if state != connectivity.Ready {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getEnvOrDefault(envName, defaultValue string) string {
	val := os.Getenv(envName)
	if envName == "" {
		return defaultValue
	}
	return val
}
