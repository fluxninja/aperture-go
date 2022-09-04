package aperture

import (
	"time"
)

const (
	clientIPHeaderName = "client-ip"

	// Library name and version can be used by the user to create a resource that connects to telemetry expoert.
	libraryName    = "aperture-go"
	libraryVersion = "v0.1.0"

	defaultRPCTimeout = 200 * time.Millisecond

	defaultGRPCReconnectionTime = 10 * time.Second

	// status of the feature.
	featureStatusLabel = "aperture.feature_status"

	// IP address of client hosting the feature.
	featureIPLabel = "aperture.feature_ip"

	// checkResponseLabel contains JSON encoded check response struct.
	checkResponseLabel = "aperture.check_response"
)
