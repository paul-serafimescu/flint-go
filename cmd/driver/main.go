package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/UCLA-IRL/flint-go/cmd/driver/server"
	"github.com/UCLA-IRL/flint-go/pkg/proto"
	"github.com/named-data/ndnd/std/engine"
	"github.com/named-data/ndnd/std/engine/face"
	"google.golang.org/grpc"
)

var (
	appPrefix      = os.Getenv("APP_PREFIX")            // "ndn-compute" or whatever
	numWorkers, _  = strconv.Atoi(os.Getenv("WORKERS")) // number of workers available
	managementPort = os.Getenv("MANAGEMENT_PORT")       // port for the RPC server
)

func main() {
	log.Println("Starting driver...")

	// UDP should be good enough here
	udpFace := face.NewStreamFace("udp", "ndnd:6363", false)
	app := engine.NewBasicEngine(udpFace)

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	defer app.Stop()

	// Make sure all workers are available
	ensureWorkersAvailable(app, numWorkers, 10)
	log.Println("All workers are ready.")

	// create rpc server
	grpcServer := grpc.NewServer()

	// register all rpcs
	proto.RegisterStaticComputeServiceServer(grpcServer, server.NewStaticComputeServer(appPrefix))

	// bind to some port, who cares
	lis, err := net.Listen("tcp", ":"+managementPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("gRPC server listening on :" + managementPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Driver shutting down.")
}
