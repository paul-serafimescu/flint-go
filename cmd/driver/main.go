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
	appPrefix     = os.Getenv("APP_PREFIX") // e.g., "ndn-compute"
	numWorkers, _ = strconv.Atoi(os.Getenv("WORKERS"))
)

func main() {
	log.Println("🚀 Starting driver...")

	// 1. Connect to NDND via UDP (or TCP if preferred)
	udpFace := face.NewStreamFace("udp", "ndnd:6363", false)
	app := engine.NewBasicEngine(udpFace)

	if err := app.Start(); err != nil {
		log.Fatalf("❌ Failed to start engine: %v", err)
	}
	defer app.Stop()

	// 2. Wait for each worker to respond to ping (max 10 tries per worker)
	maxTries := 10
	ready := make(map[int]bool)
	attempts := make(map[int]int)

	for len(ready) < numWorkers {
		for i := 1; i <= numWorkers; i++ {
			if ready[i] {
				continue
			}
			if attempts[i] >= maxTries {
				log.Fatalf("❌ Worker %d did not respond after %d tries", i, maxTries)
			}

			prefix := fmt.Sprintf("/%s/worker/%d/ready", appPrefix, i)
			log.Printf("🔄 Pinging %s (try %d)...", prefix, attempts[i]+1)

			if tryPing(app, prefix) {
				log.Printf("✅ Worker %d is ready.", i)
				ready[i] = true
			} else {
				attempts[i]++
			}

			// Short sleep between round-robin pings
			time.Sleep(500 * time.Millisecond)
		}
	}
	log.Println("✅ All workers are ready.")

	// 3. Send a test Interest to /add/3/5
	addPrefix := fmt.Sprintf("/%s/add/3/5", appPrefix)
	name, _ := enc.NameFromStr(addPrefix)
	interest, _ := app.Spec().MakeInterest(name, &ndn.InterestConfig{
		MustBeFresh: true,
		Lifetime:    optional.Some(3 * time.Second),
		Nonce:       utils.ConvertNonce(app.Timer().Nonce()),
	}, nil, nil)

	log.Printf("📤 Sending Interest: %s", name.String())

	ch := make(chan struct{})
	_ = app.Express(interest, func(args ndn.ExpressCallbackArgs) {
		switch args.Result {
		case ndn.InterestResultData:
			log.Printf("✅ Got Data: %s", args.Data.Name())
			log.Printf("📦 Result = %s", string(args.Data.Content().Join()))
		case ndn.InterestResultNack:
			log.Printf("🚫 Nacked: reason = %d", args.NackReason)
		case ndn.InterestResultTimeout:
			log.Printf("⏳ Timed out")
		default:
			log.Printf("❓ Unknown Interest result")
		}
		ch <- struct{}{}
	})
	<-ch

	// 4. Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("👋 Driver shutting down.")
}

func tryPing(app ndn.Engine, prefix string) bool {
	name, _ := enc.NameFromStr(prefix)
	interest, _ := app.Spec().MakeInterest(name, &ndn.InterestConfig{
		MustBeFresh: true,
		Lifetime:    optional.Some(2 * time.Second),
		Nonce:       utils.ConvertNonce(app.Timer().Nonce()),
	}, nil, nil)

	result := make(chan bool, 1)
	_ = app.Express(interest, func(args ndn.ExpressCallbackArgs) {
		result <- args.Result == ndn.InterestResultData
	})

	select {
	case ok := <-result:
		return ok
	case <-time.After(3 * time.Second):
		return false
	}
}
