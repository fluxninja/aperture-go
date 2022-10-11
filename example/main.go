package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	aperture "github.com/fluxninja/aperture-go/sdk"
)

const (
	defaultAgentHost = "localhost"
	defaultAgentPort = "8089"
	defaultAppPort   = "8080"
)

// app struct contains the server and the Aperture client.
type app struct {
	server         *http.Server
	grpcClient     *grpc.ClientConn
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
	agentHost := getEnvOrDefault("FN_AGENT_HOST", defaultAgentHost)
	agentPort := getEnvOrDefault("FN_AGENT_PORT", defaultAgentPort)

	ctx := context.Background()

	apertureAgentGRPCClient, err := grpcClient(ctx, net.JoinHostPort(agentHost, agentPort))
	if err != nil {
		log.Fatalf("failed to create flow control client: %v", err)
	}

	opts := aperture.Options{
		ApertureAgentGRPCClientConn: apertureAgentGRPCClient,
		CheckTimeout:                200 * time.Millisecond,
	}

	// initialize Aperture Client with the provided options.
	apertureClient, err := aperture.NewClient(ctx, opts)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	appPort := getEnvOrDefault("FN_APP_PORT", defaultAppPort)
	// Create a server with passing it the Aperture client.
	mux := http.NewServeMux()
	a := &app{
		server: &http.Server{
			Addr:    net.JoinHostPort("localhost", appPort),
			Handler: mux,
		},
		apertureClient: apertureClient,
		grpcClient:     apertureAgentGRPCClient,
	}

	mux.HandleFunc("/super", a.SuperHandler)
	mux.HandleFunc("/connected", a.ConnectedHandler)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %+v", err)
		}
	}()

	<-done
	if err := apertureClient.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown aperture client: %+v", err)
	}
	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %+v", err)
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
		w.WriteHeader(http.StatusAccepted)
		// Simulate work being done
		time.Sleep(2 * time.Second)
		// Need to call End() on the Flow in order to provide telemetry to Aperture Agent for completing the control loop.
		// The first argument captures whether the feature captured by the Flow was successful or resulted in an error.
		// The second argument is error message for further diagnosis.
		_ = flow.End(aperture.OK)
	} else {
		w.WriteHeader(http.StatusForbidden)
		// Flow has been rejected by Aperture Agent.
		_ = flow.End(aperture.Error)
	}
}

// ConnectedHandler handles HTTP requests on /connected API endpoint.
func (a *app) ConnectedHandler(w http.ResponseWriter, r *http.Request) {
	a.grpcClient.Connect()
	state := a.grpcClient.GetState()
	if state != connectivity.Ready {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_, _ = w.Write([]byte(state.String()))
}

func getEnvOrDefault(envName, defaultValue string) string {
	val := os.Getenv(envName)
	if val == "" {
		return defaultValue
	}
	return val
}
