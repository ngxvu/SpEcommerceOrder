I create this directory to serve as a base source for other projects.

to start grpc run 
```bash
protoc --go_out=. --go-grpc_out=. pkg/proto/order.proto
```