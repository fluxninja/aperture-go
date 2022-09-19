package aperture

import (
	"context"
	"net/url"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	flowcontrolgrpc "go.buf.build/grpc/go/fluxninja/aperture/aperture/flowcontrol/v1"
)

// Client is the interface that is provided to the user upon which they can perform Check calls for their service and eventually shut down in case of error.
type Client interface {
	StartFlow(ctx context.Context, feature string, labels map[string]string) (Flow, error)
}

type apertureClient struct {
	flowControlClient flowcontrolgrpc.FlowControlServiceClient
	tracer            oteltrace.Tracer
	timeout           time.Duration
}

// Options that the user can pass to Aperture in order to receive a new Client. ClientConn and Ctx are required.
type Options struct {
	ClientConn   *grpc.ClientConn
	CheckTimeout time.Duration
}

// NewClient returns a new Client that can be used to perform Check calls.
// The user will pass in options which will be used to create a connection with otel and a tracerProvider retrieved from such connection.
func NewClient(options Options) (Client, error) {
	exporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(options.ClientConn), otlptracegrpc.WithReconnectionPeriod(defaultGRPCReconnectionTime))
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

	tracer := tracerProvider.Tracer(libraryName)

	fcClient := flowcontrolgrpc.NewFlowControlServiceClient(options.ClientConn)

	var timeout time.Duration
	if options.CheckTimeout == 0 {
		timeout = defaultRPCTimeout
	} else {
		timeout = options.CheckTimeout
	}

	c := &apertureClient{
		flowControlClient: fcClient,
		tracer:            tracer,
		timeout:           timeout,
	}
	runtime.SetFinalizer(c, exporter.Shutdown(context.Background()))
	return c, nil
}

// StartFlow takes a feature name and labels that get passed to Aperture Agent via flowcontrolv1.Check call.
// Return value is a Flow.
// The call returns immediately in case connection with Aperture Agent is not established.
// The default semantics are fail-to-wire. If StartFlow fails, calling Flow.Accepted() on returned Flow returns as true.
func (c *apertureClient) StartFlow(ctx context.Context, feature string, explicitLabels map[string]string) (Flow, error) {
	context, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	labels := make(map[string]string)

	// Inherit labels from baggage
	baggageCtx := baggage.FromContext(context)
	for _, member := range baggageCtx.Members() {
		value, err := url.QueryUnescape(member.Value())
		if err != nil {
			continue
		}
		labels[member.Key()] = value
	}

	// Explicit labels override baggage
	for key, value := range explicitLabels {
		labels[key] = value
	}

	req := &flowcontrolgrpc.CheckRequest{
		Feature: feature,
		Labels:  labels,
	}

	var header metadata.MD

	_, span := c.tracer.Start(context, "Aperture Check")
	span.SetAttributes(
		attribute.Int64(flowStartTimestampLabel, time.Now().UnixNano()),
		attribute.String(sourceLabel, "sdk"),
	)

	res, err := c.flowControlClient.Check(context, req, grpc.Header(&header))
	ipValue := ""
	ipHeader := header.Get(clientIPHeaderName)
	if len(ipHeader) == 1 {
		ipValue = ipHeader[0]
	}

	span.SetAttributes(
		attribute.Int64(checkResponseTimestampLabel, time.Now().UnixNano()),
	)

	if err != nil {
		return &flow{
			checkResponse: nil,
			clientIP:      ipValue,
			span:          span,
		}, err
	}

	return &flow{
		checkResponse: res,
		clientIP:      ipValue,
		span:          span,
	}, nil
}

// newResource returns a resource describing the running process, containing the library name and version.
func newResource() (*resource.Resource, error) {
	resourceDefault := resource.Default()
	r, err := resource.Merge(
		resourceDefault,
		resource.NewWithAttributes(
			resourceDefault.SchemaURL(),
			semconv.ServiceNameKey.String(libraryName),
			semconv.ServiceVersionKey.String(libraryVersion),
		),
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}
