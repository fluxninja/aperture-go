package aperture

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

<<<<<<<< HEAD:sdk/aperture-client.go
|||||||| 4dfd08c:sdk/apertureClient.go
const (
	// Ok and Error indicate feature execution. User will have to pass one of the two values when ending a flow.
	Ok    Code = 0
	Error Code = 1

	clientIPHeaderName = "client-ip"

	// Library name and version can be used by the user to create a resource that connects to telemetry expoert.
	LibraryName    = "aperture-go"
	LibraryVersion = "v0.1.0"

	defaultRPCTimeout = 200 * time.Millisecond

	defaultGRPCReconnectionTime = 10 * time.Second
)

========
	flowcontrolgrpc "go.buf.build/grpc/go/fluxninja/aperture/flowcontrol/v1"
)

>>>>>>>> main:sdk/client.go
// Code is an 32-bit representation of a status state.
type Code uint32

// Client is the interface that is provided to the user upon which they can perform Check calls for their service and eventually shut down in case of error.
type Client interface {
	Check(ctx context.Context, feature string, labels map[string]string) (Flow, error)
}

type apertureClient struct {
	flowControlClient flowcontrolgrpc.FlowControlServiceClient
	tracer            oteltrace.Tracer
	timeout           time.Duration
	tracerProvider    *trace.TracerProvider
}

<<<<<<<< HEAD:sdk/aperture-client.go
// Options contains fields that are used to create an ApertureClient instance. ClientConn and Ctx are required.
|||||||| 4dfd08c:sdk/apertureClient.go
// Options that the user can pass to Aperture in order to receive a new ApertureClient. ClientConn and Ctx are required.
========
// Options that the user can pass to Aperture in order to receive a new Client. ClientConn and Ctx are required.
>>>>>>>> main:sdk/client.go
type Options struct {
	ClientConn   *grpc.ClientConn
	CheckTimeout time.Duration
	Ctx          context.Context
}

<<<<<<<< HEAD:sdk/aperture-client.go
// NewApertureClient returns a new ApertureClient that can be used to perform Check calls.
// The user will pass in options which will be used to create a grpc connection.
func NewApertureClient(options Options) (ApertureClient, error) {
|||||||| 4dfd08c:sdk/apertureClient.go
// NewApertureClient returns a new ApertureClient that can be used to perform Check calls.
// The user will pass in options which will be used to create a connection with otel and a tracerProvider retrieved from such connection.
func NewApertureClient(options Options) (ApertureClient, error) {
========
// NewClient returns a new Client that can be used to perform Check calls.
// The user will pass in options which will be used to create a connection with otel and a tracerProvider retrieved from such connection.
func NewClient(options Options) (Client, error) {
>>>>>>>> main:sdk/client.go
	var timeout time.Duration
	flowControlClient := flowcontrolgrpc.NewFlowControlServiceClient(options.ClientConn)

	if options.CheckTimeout == 0 {
		timeout = defaultRPCTimeout
	} else {
		timeout = options.CheckTimeout
	}

	exporter, err := otlptracegrpc.New(options.Ctx, otlptracegrpc.WithGRPCConn(options.ClientConn), otlptracegrpc.WithReconnectionPeriod(defaultGRPCReconnectionTime))
	if err != nil {
		return nil, err
	}

	newResource, err := newResource()
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter)),
		trace.WithResource(newResource),
	)

	otel.SetTracerProvider(tracerProvider)

	tracer := tracerProvider.Tracer(LibraryName)

	runtime.SetFinalizer(&exporter, exporter.Shutdown(options.Ctx))
	return &apertureClient{
		flowControlClient: flowControlClient,
		tracer:            tracer,
		timeout:           timeout,
		tracerProvider:    tracerProvider,
	}, nil
}

// Check is a call performed on the FlowControlServiceClient, passing in the feature name and labels that the user wants to send to Aperture.
// The user will receive a Flow interface return upon which they can perform End calls.
// The check call will return a nil check response in case connection with flow control service is not established.
func (apc *apertureClient) Check(ctx context.Context, feature string, labels map[string]string) (Flow, error) {
	context, cancel := context.WithTimeout(ctx, apc.timeout)
	defer cancel()

	overiddenLabels := make(map[string]string)
	fcsLatencyStart := time.Now()

	newBaggage := baggage.FromContext(context)

	labelsFromBaggage := newBaggage.Members()
	for _, label := range labelsFromBaggage {
		overiddenLabels[asString(label.Key())] = asString(label.Value())
	}

	for key, value := range labels {
		overiddenLabels[key] = value
	}

	req := &flowcontrolgrpc.CheckRequest{
		Feature: feature,
		Labels:  overiddenLabels,
	}

	var header metadata.MD

	_, span := apc.tracer.Start(context, "Aperture Check")

	res, err := apc.flowControlClient.Check(context, req, grpc.Header(&header))
	ipValue := ""
	ipHeader := header.Get(clientIPHeaderName)
	if len(ipHeader) == 1 {
		ipValue = ipHeader[0]
	}

	if err != nil {
		return &flow{
			checkResponse:   nil,
			fcsLatencyStart: fcsLatencyStart,
			clientIP:        ipValue,
			span:            span,
		}, err
	}

	return &flow{
		checkResponse:   res,
		fcsLatencyStart: fcsLatencyStart,
		clientIP:        ipValue,
		span:            span,
	}, nil
}

// newResource returns a resource describing the running process, containing the library name and version.
func newResource() (*resource.Resource, error) {
	resourceDefault := resource.Default()
	r, err := resource.Merge(
		resourceDefault,
		resource.NewWithAttributes(
			resourceDefault.SchemaURL(),
			semconv.ServiceNameKey.String(LibraryName),
			semconv.ServiceVersionKey.String(LibraryVersion),
		),
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// asString returns the string representation of a key or value.
func asString(val any) string {
	bytes, err := json.Marshal(val)
	if err != nil {
		fmt.Println("Error occurred marshaling json: ", err)
		return ""
	}
	return string(bytes)
}
