# Aperture-Go SDK

Aperture-Go

## Aperture-Go SDK provides APIs to interact with Aperture Agent. These APIs enable flow control functionality on fine-grained features inside service code.

### ApertureClient Interface

An `ApertureClient` maintains a grpc connection with ApertureAgent.

### Usage/Examples

```golang
	options := aperture.Options{
		ClientConn:   client,
		CheckTimeout: 200 * time.Millisecond,
		Ctx:          ctx,
	}

	client, err := aperture.NewClient(options)
```

### Flow Interface

A `Flow` is created every time a `ApertureClient.BeginFlow` is called.

```golang
	// BeginFlow performs a flowcontrolv1.Check call to Aperture Agent. It returns a Flow and an error if any.
	flow, err := a.apertureClient.BeginFlow(ctx, "awesomeFeature", labels)
    	if err != nil {
		log.Warn("Aperture flow control got error. Returned flow defaults to Allowed. flow.Accepted(): %t", flow.Accepted())
	}

	// See whether flow was accepted by Aperture Agent
	if flow.Accepted() {
		// Simulation of work that client would do if the feature is enabled.
		time.Sleep(5 * time.Second)
	} else {
		// Flow has been rejected by Aperture Agent, return appropriate response to caller of this feature
		log.Info("Flow rejected by Aperture Agent")
	}
	// Need to call End on the Flow in order to provide telemetry to Aperture Agent for completing the control loop. The first argument catpures whether the feature captured by the Flow was successful or resulted an error. The second argument is error message for further diagnosis.
	flow.End(aperture.Ok, "")
```

## 🔗 Links to relevant Aperture Resources:

[![Github](https://camo.githubusercontent.com/cca71357fe98ec5f8cd6ebab9044ad2901f4b64ebda379ac81608ed9f1caa1a0/68747470733a2f2f696d672e736869656c64732e696f2f7374617469632f76313f7374796c653d666f722d7468652d6261646765266d6573736167653d47697448756226636f6c6f723d313831373137266c6f676f3d476974487562266c6f676f436f6c6f723d464646464646266c6162656c3d)](https://github.com/fluxninja/aperture)
