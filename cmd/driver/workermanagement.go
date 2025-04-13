package main

import (
	"fmt"
	"log"
	"time"

	enc "github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/ndn"
	"github.com/named-data/ndnd/std/types/optional"
	"github.com/named-data/ndnd/std/utils"
)

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

// Current implementation: ping each worker periodically until they ACK, fail after 9 retries
func ensureWorkersAvailable(app ndn.Engine, numWorkers int, maxTries int) {
	ready := make(map[int]bool)
	attempts := make(map[int]int)

	for len(ready) < numWorkers {
		for i := 1; i <= numWorkers; i++ {
			if ready[i] {
				continue
			}
			if attempts[i] >= maxTries {
				log.Fatalf("Worker %d did not respond after %d tries", i, maxTries)
			}

			prefix := fmt.Sprintf("/%s/worker/%d/ready", appPrefix, i)
			log.Printf("Pinging %s (try %d)...", prefix, attempts[i]+1)

			if tryPing(app, prefix) {
				log.Printf("Worker %d is ready.", i)
				ready[i] = true
			} else {
				attempts[i]++
			}

			// Short sleep between round-robin pings
			time.Sleep(500 * time.Millisecond)
		}
	}
}
