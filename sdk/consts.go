package aperture

import (
	"time"
)

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

	// StatusCodeLabel describes HTTP status code of the response.
	StatusCodeLabel = "aperture.status_code"

	// FeatureAddressLabel describes feature address of the request.
	FeatureAddressLabel = "feature.ip"

	// MarshalledCheckResponseLabel contains JSON encoded check response struct.
	MarshalledCheckResponseLabel = "aperture.check_response"

	// DecisionErrorReasonLabel describes the error reason of the decision taken by policy.
	DecisionErrorReasonLabel = "fcs.decision_error_reason"

	// TimestampLabel describes timestamp of the request.
	TimestampLabel = "timestamp"
)
