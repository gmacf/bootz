package main

import (
	"flag"

	pb "github.com/openconfig/bootz/server/tests/proto/sut"
)

var (
	grpcPort = flag.Int("grpc_port", 50051, "The gRPC server port")
	httpPort = flag.Int("http_port", 8080, "The HTTP server port")
)

type server struct {
	pb.UnimplementedImageServiceServer
}

func main() {

}
