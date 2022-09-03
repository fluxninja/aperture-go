# Aperture-Go SDK

Aperture-Go SDK provides APIs to interact with Aperture Agent. These APIs enable flow control functionality on fine-grained features inside service code.

## Usage

### ApertureClient Interface

An `ApertureClient` maintains a grpc connection with ApertureAgent.

https://github.com/fluxninja/aperture-go/blob/2c5414a29e2a3cbc98481383b3a7dee6ccfa596e/example/example_test.go#L39-L46

### Flow Interface

A `Flow` is created every time a `ApertureClient.BeginFlow` is called.

https://github.com/fluxninja/aperture-go/blob/2c5414a29e2a3cbc98481383b3a7dee6ccfa596e/example/example_test.go#L77-L92

## Relevant Resources

[Aperture GitHub](https://github.com/fluxninja/aperture)
