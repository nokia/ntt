This README is still a heavy work in progress

# TTCN-3 Architecture Overview

ETSI provides a TTCN-3 reference architecture including interfaces, using OMG
IDL files, which would allow for automatic stub creation. Unfortunately IDL is
very antiquated and support for younger languages is virtually not available.  
For that reason we give [gRPC](http://grpc.io) a go and provide a similar
architecture, using tools which are more available.

We also introduce a some adaptions of which we think are beneficial for
low-latency communication and test throughput.

* Logging interface (TCI-TL) will be removed in favor of specialized solutions
  like [Zipkin](https://zipkin.io/), [Jaeger](https://www.jaegertracing.io/) or
  [OpenTelemetry](https://opentelemetry.io/). Because traces become big and can 
  negatively tests. We don't want to send them via gRPC, but use side-channels.
* Centralized component handling (TCI-CH) will be removed. System adapters
  transmits messages directly to peer. This increases complexity of system
  adapters but reduces latency.
* We don't distribute components over the network, but tests: Every node can
  execute its own test, with its own components and adapters.


```


    +------------------------------+
    |                              |
    |          CONTROLLER          |
    |                              |
    +--[TCI-TM]-[TCI-CH]-[TCI-TL]--+
           ^        ^        ^
           |        |        |
           |        |        |
           v        v        |
    +------------------------------+   +------------+
    |                              |+  |            |
    |           RUNTIME          <------->  T3XF    |
    |                              ||  |            |
    +-+[TCI-CD]-[TRI-SA]-[TRI-PA]--+|  +------------+
      +-----------------------------+




                             +-----------------+     
                             |                 | <--[TCI-CD]-->  Codecs
 Controller  <--[TCI-TM]-->  |      T3RTS      | 
                             |                 | <--[TRI-SA]--> System Adapters (Connectors)
                             |                 | <--[TRI-PA]--> Extfuncs  (TRI-SA)
                             +-----------------+

```
## Controller

A controller or _Test Managemant and Control_ (TMC) is the component responsible
for:

* Executing tests (TCI-TM)
* Logging (TCI-TL)
* Distributing components and communication (TCI-CH)

## Component Handling


    CreateTestComponent
    StopTestComponent
    Connect
    Disconnect
    Map
    Unmap
    ExecuteTest
    TestComponentDone
    Reset
}

## Codecs

ETSI distinugishes between external codecs (TRI-CD) and built-in codecs:

> The external codecs can be used in parallel with, or instead of, the built-in
> codecs associated with the TE. Unlike the built-in codecs the external codecs
> have a standardized interface which makes them portable between different
> TTCN-3 systems and tools.

The external codecs are typically used for message and procedure based
communication. The built-in codecs are typically used by `decvalue` and
`encvalue` and have slightly different requirements.


## Runtime

The runtime or _Test Executable_ (TE) the composed of three sub-systems:

* Executable Test Suite (ETS)
* TTCN-3 RunTime System (T3RTS)
* (optional) Encoding/Decoding System (EDS)


# Generate Stubs

To generate the protobuf files, make sure you have go, protoc, grpc and
protoc-gen-go installed:

    sudo dnf install protobuf-compiler
    go get -u google.golang.org/grpc
    go get -u github.com/golang/protobuf/protoc-gen-go

Don't forget to put the `$GOPATH/bin` into your `PATH` variable, or `protoc`
won't find `protoc-gen-go`. And then, from this directory, call:

    go generate

You should now have various updated Go files, like `ttcn3.pb.go`.


# Various Notes from ETSI Standards

## Communication (TRI-SA)

 * SA <-- Reset()
 * SA <-- ExecuteTestCase(id, tsi-ports)
 * SA <-- Map(compport, tsiport, params) return err
 * SA <-- Unmap(compport, tsiport, params)
 * SA <-- EndTestCase()
 * SA <-- SendMessage
 * SA <-- Call
 * SA <-- Reply
 * SA <-- Raise

**Notifications**

 * TE <-- Error
 * TE <-- EnqueueMessage
 * TE <-- EnqueueCall
 * TE <-- EnqueueReply
 * TE <-- EnqueueException

## Platform (TRI-PA)

 * PA <-- StartTimer
 * PA <-- StopTimer
 * PA <-| ReadTimer(id) returns elapsedtime
 * PA <-| TimerRunning(id) returns bool
 * PA <-- ExtFunc

**Notifications**

 * TE <-- Error
 * TE <-- Timeout
 * TE <-| Self() return compId
 * TE <-| Rnd() return Integer


## Test Management (TCI-TM)

 * TE <-- RootModule
 * TE <-| ImportedModules returns modules
 * TE <-| ModuleParameters returns params
 * TE <-| Testcases returns tests
 * TE <-| TestcaseParameters returns params
 * TE <-| TestcaseTSI returns ports
 * TE <-- StartTestcase
 * TE <-- StopTestcase
 * TE <-- StartControl
 * TE <-- StartControlWithParams
 * TE <-- StopControl
 * TE <-- ControlParameters

**Notifications**

 * TM <-- TestcaseStarted
 * TM <-- TastcaseTerminated
 * TM <-- ControlTerminated
 * TM <-- Log
 * TM <-- Error
 * TM <-| GetModulePar returns Value

## Component Handling (TCI-CH)


service SystemAdapter {
  rpc Reset() {}
  rpc Prepare(PrepareMessage) {}
  rpc Map(MapRequest) {}
  rpc UnMap(MapRequest) {}
  rpc Finish() {}
  rpc Send() {}
  rpc Call() {}
  rpc Reply() {}
  rpc Raise() {}
  rpc CommunicationEvents() returns (stream CommunicationEvent) {}
}

message PrepareMessage {
  string test_id = 1;
  repeated string port_ids = 2;
}

message MapRequest {
  string port_id = 1;
  string comp_id = 2;
  repeated Parameter params = 3;
}

message CommunicationEvent {
  //* Error
  //* EnqueueMessage
  //* EnqueueCall
  //* EnqueueReply
  //* EnqueueException
}

/*

## Test Management (TCI-TM)

 * TE <-- RootModule
 * TE <-| ImportedModules returns modules
 * TE <-| ModuleParameters returns params
 * TE <-| Testcases returns tests
 * TE <-| TestcaseParameters returns params
 * TE <-| TestcaseTSI returns ports
 * TE <-- StartTestcase
 * TE <-- StopTestcase
 * TE <-- StartControl
 * TE <-- StartControlWithParams
 * TE <-- StopControl
 * TE <-- ControlParameters

**Notifications**

 * TM <-- TestcaseStarted
 * TM <-- TastcaseTerminated
 * TM <-- ControlTerminated
 * TM <-- Log
 * TM <-- Error
 * TM <-| GetModulePar returns Value
 */
service Runtime { rpc ListTests() return }

