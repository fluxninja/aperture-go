package aperture

import (
	"time"

	flowcontrolv1 "github.com/fluxninja/aperture/api/gen/proto/go/aperture/flowcontrol/v1"
	otel "github.com/fluxninja/aperture/pkg/otelcollector"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Flows is the interface that is returned to the user everytime a check call is performed.
// The user can check the status of the check call, the response from the server and once the feature is executed, end the flow.
type Flow interface {
	Accepted() bool
	End(statusCode Code, errDescription string)
	CheckResponse() *flowcontrolv1.CheckResponse
}

type flow struct {
	checkResponse   *flowcontrolv1.CheckResponse
	fcsLatencyStart time.Time
	clientIP        string
	span            trace.Span
}

// Accepted returns the state of the connection with the server.
// ApertureClient is a faile-to-wire system so it will return true even if the connenction did not happen.
func (f *flow) Accepted() bool {
	if f.checkResponse == nil {
		return true
	}
	if f.checkResponse.DecisionType == flowcontrolv1.DecisionType_DECISION_TYPE_ACCEPTED {
		return true
	}
	return false
}

// CheckResponse returns the response from the server.
func (f *flow) CheckResponse() *flowcontrolv1.CheckResponse {
	return f.checkResponse
}

// End is used to end the flow, the user will have to pass a status code and an error description which will be used to investigate the response in case of failure.
func (f *flow) End(statusCode Code, errDescription string) {
	defer f.span.End()
	f.span.SetStatus(codes.Ok, errDescription) // Find a way to pass statusCode instead of codes.Ok
	f.span.SetAttributes(
		attribute.String(otel.FeatureAddressLabel, f.clientIP),
		attribute.String(otel.MarshalledCheckResponseLabel, asString(f.checkResponse)),
	)
	if errDescription != "" {
		f.span.SetAttributes(attribute.String(otel.DecisionErrorReasonLabel, errDescription))
	}
}
