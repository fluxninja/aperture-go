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

	"github.com/go-logr/stdr"
	"github.com/gorilla/mux"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"

	aperturego "github.com/fluxninja/aperture-go/v2/sdk"
	aperturegomiddleware "github.com/fluxninja/aperture-go/v2/sdk/middleware"
)

const (
	defaultAgentAddress = "localhost:8089"
	defaultAppPort      = "8080"
)

// app struct contains the server and the Aperture client.
type app struct {
	server         *http.Server
	apertureClient aperturego.Client
}

// grpcOptions creates a new gRPC client that will be passed in order to initialize the Aperture client.
func grpcOptions() []grpc.DialOption {
	var grpcDialOptions []grpc.DialOption
	grpcDialOptions = append(grpcDialOptions, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 10,
	}))
	grpcDialOptions = append(grpcDialOptions, grpc.WithUserAgent("aperture-go"))

	return grpcDialOptions
}

func main() {
	ctx := context.Background()

	stdr.SetVerbosity(2)

	opts := aperturego.Options{
		Address:         getEnvOrDefault("APERTURE_AGENT_ADDRESS", defaultAgentAddress),
		GRPCDialOptions: grpcOptions(),
		AgentAPIKey:     getEnvOrDefault("APERTURE_AGENT_API_KEY", ""),
		Insecure:        getBoolEnvOrDefault("APERTURE_AGENT_INSECURE", false),
		SkipVerify:      getBoolEnvOrDefault("APERTURE_AGENT_SKIP_VERIFY", false),
	}

	// initialize Aperture Client with the provided options.
	apertureClient, err := aperturego.NewClient(ctx, opts)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	appPort := getEnvOrDefault("FN_APP_PORT", defaultAppPort)
	// Create a server with passing it the Aperture client.
	mux := mux.NewRouter()
	a := &app{
		server: &http.Server{
			Addr:    net.JoinHostPort("localhost", appPort),
			Handler: mux,
		},
		apertureClient: apertureClient,
	}

	// Adding the http middleware to be executed before the actual business logic execution.
	superRouter := mux.PathPrefix("/super").Subrouter()
	superRouter.HandleFunc("", a.SuperHandler)
	superRouter.Use(aperturegomiddleware.NewHTTPMiddleware(apertureClient, "awesomeFeature", nil, nil, false, 2000*time.Millisecond).Handle)

	mux.HandleFunc("/connected", a.ConnectedHandler)
	mux.HandleFunc("/health", a.HealthHandler)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting example server")

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

// SuperHandler handles HTTP requests on /super endpoint.
func (a *app) SuperHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	// Simulate work being done
	time.Sleep(2 * time.Second)
}

// ConnectedHandler handles HTTP requests on /connected endpoint.
func (a *app) ConnectedHandler(w http.ResponseWriter, r *http.Request) {
	a.apertureClient.GetGRPClientConn().Connect()
	state := a.apertureClient.GetGRPClientConn().GetState()
	if state != connectivity.Ready {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_, _ = w.Write([]byte(state.String()))
}

// HealthHandler handles HTTP requests on /health endpoint.
func (a *app) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Healthy"))
}

func getBoolEnvOrDefault(envName string, defaultValue bool) bool {
	val := os.Getenv(envName)
	if val == "" {
		return defaultValue
	}
	return cast.ToBool(val)
}

func getEnvOrDefault(envName, defaultValue string) string {
	val := os.Getenv(envName)
	if val == "" {
		return defaultValue
	}
	return val
}
