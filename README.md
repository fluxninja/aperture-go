# FluxNinja's Aperture-Go SDK

Aperture library allows connection to the Flow Control Service in order to
perfom check calls upon feature handling.

## Aperture SDK API

### ApertureClient Interface

Requires a grpc client connection and context for initialization. Allows
Client's to perform check call with Flow Control Service.

### Usage/Examples

```golang
	options := aperture.Options{
		ClientConn:   client,
		CheckTimeout: 200 * time.Millisecond,
		Ctx:          ctx,
	}

	apertureClient, err := aperture.NewApertureClient(options)
```

### Flow Interface

Gets created everytime a check call is made through Aperture Client.

```golang
	// Perform a check call to Flow Control Service, which will return a flow and an error if any.
	flow, err := a.apertureClient.Check(ctx, "awesomeFeature", labels)
    	if err != nil {
		errMessage = err.Error()
	}

	if flow.Accepted() {
		// Simulation of work that client would do if the feature is enabled.
		time.Sleep(5 * time.Second)
	} else {
		_, err := json.Marshal(flow.CheckResponse())
		if err != nil {
			// in case of error, record it in a string and return it to the flow.
			errMessage = err.Error()
		}
	}
	// if the feature was correctly executed but a minor error occurred, send it when ending the flow. In case of not unsuccessflow flow send aperture.Error.
	flow.End(aperture.Ok, errMessage)
```

## 🔗 Links to Aperture repo and social media

[![Github](https://camo.githubusercontent.com/cca71357fe98ec5f8cd6ebab9044ad2901f4b64ebda379ac81608ed9f1caa1a0/68747470733a2f2f696d672e736869656c64732e696f2f7374617469632f76313f7374796c653d666f722d7468652d6261646765266d6573736167653d47697448756226636f6c6f723d313831373137266c6f676f3d476974487562266c6f676f436f6c6f723d464646464646266c6162656c3d)](https://github.com/fluxninja/aperture)
[![linkedin](https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/company/fluxninja/mycompany/)
[![twitter](https://img.shields.io/badge/twitter-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white)](https://twitter.com/fluxninjahq?lang=en)

