# Aperture-Go SDK

Aperture-Go SDK provides APIs to interact with Aperture Agent. These APIs enable flow control functionality on fine-grained features inside service code.

### ApertureClient Interface

An `ApertureClient` maintains a grpc connection with ApertureAgent.

#### Usage/Examples
https://github.com/fluxninja/aperture-go/blob/465f7be06869a534a8495f7a119c5574088611ab/example/example_test.go#L39-L46

### Flow Interface

A `Flow` is created every time a `ApertureClient.BeginFlow` is called.

https://github.com/fluxninja/aperture-go/blob/465f7be06869a534a8495f7a119c5574088611ab/example/example_test.go#L77-L92

## 🔗 Relevant Resources:

[Aperture GitHub](https://github.com/fluxninja/aperture)
