package aperture

import (
	"time"
)

const (
	clientIPHeaderName = "client-ip"

	// Library name and version can be used by the user to create a resource that connects to telemetry expoert.
	libraryName    = "aperture-go"
	libraryVersion = "v0.1.0"

	// Config defaults.
	defaultRPCTimeout           = 200 * time.Millisecond
	defaultGRPCReconnectionTime = 10 * time.Second

	// Label keys.
	// Label to hold source of flow.
	sourceLabel = "aperture.source"
	// Label to hold status of the feature.
	featureStatusLabel = "aperture.feature_status"
	// Label to hold IP address of client hosting the feature.
	featureIPLabel = "aperture.feature_ip"
	// Label to hold JSON encoded check response struct.
	checkResponseLabel = "aperture.check_response"
	// Label to hold flow's start timestamp in Unix nanoseconds since Epoch.
	flowStartTimestampLabel = "aperture.flow_start_timestamp"
	// Label to hold flow's stop timestamp in Unix nanoseconds since Epoch.
	flowStopTimestampLabel = "aperture.flow_stop_timestamp"
	// Label to hold check response timestamp in Unix nanoseconds since Epoch.
	checkResponseTimestampLabel = "aperture.check_response_timestamp"
)
