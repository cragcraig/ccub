# Carbon Cub EX2 Builder's Log

## Dependencies

* [Go](https://golang.org/) language tooling
* [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)
  * [Protocol Buffers Basics: Go](https://developers.google.com/protocol-buffers/docs/gotutorial)
  * [Compiling you protocol buffers](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers)

## Build

### Protocol Buffers
To rebuild the protocol buffer `.pb.go` file after changes to the `.proto` description:
```shell
cd $SRC_DIR
protoc --go_out=. build_log.proto
```
