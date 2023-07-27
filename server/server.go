// Bootz server reference implementation.
//
// The bootz server will provide a simple file based bootstrap
// implementation for devices.  The service can be extended by
// provding your own implementation of the entity manager.
package main

import (
	"flag"
	"fmt"
	"net"

	log "github.com/golang/glog"
	"github.com/openconfig/bootz/proto/bootz"
	"github.com/openconfig/bootz/server/entitymanager"
	"github.com/openconfig/bootz/server/service"
	"google.golang.org/grpc"
)

var (
	port = flag.String("port", "", "The port to start the Bootz server on localhost")
)

func main() {
	flag.Parse()
	em := &entitymanager.EM{}
	c := service.New(em)
	s := grpc.NewServer()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", *port))
	if err != nil {
		log.Exitf("Error listening on port: %v", err)
	}
	s.RegisterService(&bootz.Bootstrap_ServiceDesc, c)
	err = s.Serve(lis)
	if err != nil {
		log.Exitf("Error serving grpc: %v", err)
	}
}
