package aperture

import (
	"errors"
	"fmt"

	"github.com/fluxninja/aperture/pkg/otelcollector"
	flowcontrolproto "go.buf.build/grpc/go/fluxninja/aperture/aperture/flowcontrol/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

// Flow is the interface that is returned to the user everytime a Check call through ApertureClient is made.
// The user can check the status of the check call, the response from the server and once the feature is executed, end the flow.
type Flow interface {
	Accepted() bool
	End(statusCode Code) error
	CheckResponse() *flowcontrolproto.CheckResponse
}

type flow struct {
	span          trace.Span
	checkResponse *flowcontrolproto.CheckResponse
	clientIP      string
	ended         bool
}

// Accepted returns whether the Flow was accepted by Aperture Agent.
func (f *flow) Accepted() bool {
	if f.checkResponse == nil {
		return true
	}
	if f.checkResponse.DecisionType == flowcontrolproto.DecisionType_DECISION_TYPE_ACCEPTED {
		return true
	}
	return false
}

// CheckResponse returns the response from the server.
func (f *flow) CheckResponse() *flowcontrolproto.CheckResponse {
	return f.checkResponse
}

// End is used to end the flow, the user will have to pass a status code and an error description which will define the state and result of the flow.
func (f *flow) End(statusCode Code) error {
	if f.ended {
		fmt.Println("Flow already ended")
		return errors.New("flow already ended")
	}
	f.ended = true

	checkResponseJSONBytes, err := protojson.Marshal(f.checkResponse)
	if err != nil {
		fmt.Printf("Marshalling: %v\n", err)
		return err
	}
	fmt.Printf("Ending flow: %v\n", []string{featureStatusLabel, featureIPLabel, checkResponseLabel})
	f.span.SetAttributes(
		attribute.String(otelcollector.FeatureStatusLabel, statusCode.String()),
		attribute.String(otelcollector.FeatureAddressLabel, f.clientIP),
		attribute.String(otelcollector.MarshalledCheckResponseLabel, string(checkResponseJSONBytes)),
	)
	f.span.End()
	fmt.Println("YOYOYO this is nil")
	return nil
}
