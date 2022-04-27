# Carbon Cub EX2 Builder's Log

## Dependencies

* [Go](https://golang.org/) language tooling
* [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)
  * [Protocol Buffers Basics: Go](https://developers.google.com/protocol-buffers/docs/gotutorial)
  * [Compiling Protocol Buffers](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)

### MacOS

* brew install golang protobuf protoc-gen-go

## Build

### Protocol Buffers
To rebuild the protocol buffer `.pb.go` file after changes to the `.proto` description:
```shell
protoc --go_out=$GOPATH/src metadata.proto
```
