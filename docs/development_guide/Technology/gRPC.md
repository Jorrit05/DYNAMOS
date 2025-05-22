# Introduction

**Links**
[Official documentation](https://protobuf.dev/)
[[Protoc]] compiler instruction

`Google Remote Procedure Calls`(gRPC) is a communication protocol developed by Google, used with strictly defined protocol buffer messages. 

This page tries to explain a few simple concepts in the context of DYNAMOS but is not meant to replace any official documentation and does not claim to be exact on the details. For exactness, please look on the internet.

## General idea

Execute a (remote) function on a different service/server. 

In the context of DYNAMOS this is mostly used in sync with a sidecar container. So the `main` container needs to execute some code, but is not interested in the exact implementation, this is handled by the `sidecar`.

TODO: Add examples and drawings how a sidecar starts a GRPC server. and the main container connects with a Client...