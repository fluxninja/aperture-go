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

[![Github](https://camo.githubusercontent.com/cca71357fe98ec5f8cd6ebab9044ad2901f4b64ebda379ac81608ed9f1caa1a0/68747470733a2f2f696d672e736869656c64732e696f2f7374617469632f76313f7374796c653d666f722d7468652d6261646765266d6573736167653d47697448756226636f6c6f723d313831373137266c6f676f3d476974487562266c6f676f436f6c6f723d464646464646266c6162656c3d)](https://github.com/fluxninja/aperture)
