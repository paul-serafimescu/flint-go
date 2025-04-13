package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	enc "github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/engine"
	"github.com/named-data/ndnd/std/engine/face"
	"github.com/named-data/ndnd/std/ndn"
	"github.com/named-data/ndnd/std/types/optional"
	"github.com/named-data/ndnd/std/utils"
)

var (
	appPrefix     = os.Getenv("APP_PREFIX")            // "ndn-compute" or whatever
	numWorkers, _ = strconv.Atoi(os.Getenv("WORKERS")) // number of workers available
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

	// Send a test Interest to /add/3/5
	addPrefix := fmt.Sprintf("/%s/add/3/5", appPrefix)
	name, _ := enc.NameFromStr(addPrefix)
	interest, _ := app.Spec().MakeInterest(name, &ndn.InterestConfig{
		MustBeFresh: true,
		Lifetime:    optional.Some(3 * time.Second),
		Nonce:       utils.ConvertNonce(app.Timer().Nonce()),
	}, nil, nil)

	log.Printf("Sending Interest: %s", name.String())

	ch := make(chan struct{})
	_ = app.Express(interest, func(args ndn.ExpressCallbackArgs) {
		switch args.Result {
		case ndn.InterestResultData:
			log.Printf("Got Data: %s", args.Data.Name())
			log.Printf("Result = %s", string(args.Data.Content().Join()))
		case ndn.InterestResultNack:
			log.Printf("Nacked: reason = %d", args.NackReason)
		case ndn.InterestResultTimeout:
			log.Printf("Timed out")
		default:
			log.Printf("Unknown Interest result")
		}
		ch <- struct{}{}
	})
	<-ch

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Driver shutting down.")
}
