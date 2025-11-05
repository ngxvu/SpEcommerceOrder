I create this directory to serve as a base source for other projects.

The output of grpc will be defined by this cmd in the proto file:  option go_package = "pkg/proto/orderpb";

to start grpc run 
```bash
protoc --go_out=. --go-grpc_out=. pkg/proto/order.proto
```


[//]: # (proxy | with windows powershell)
```bash
protoc -I. -Ipkg/proto `
  --go_out=paths=source_relative:. `
  --go-grpc_out=paths=source_relative:. `
  --grpc-gateway_out=logtostderr=true,paths=source_relative:. `
  pkg/proto/order.proto
```