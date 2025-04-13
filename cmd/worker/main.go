package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UCLA-IRL/flint-go/pkg/manifest"
	"github.com/UCLA-IRL/flint-go/pkg/security"

	"github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/engine"
	"github.com/named-data/ndnd/std/engine/face"
	"github.com/named-data/ndnd/std/ndn"
	"github.com/named-data/ndnd/std/types/optional"
)

var (
	appPrefix    = os.Getenv("APP_PREFIX")
	manifestPath = "/app/manifest/fs-manifest.json"
	app          ndn.Engine // set this globally for the handler
	workerId     = os.Getenv("WORKER_ID")
)

func main() {
	time.Sleep(1 * time.Second)
	log.Println("Worker awake")

	// 1. Load signer
	signer := security.LoadOrCreateECDSASigner()

	// 2. Connect to forwarder
	tcpFace := face.NewStreamFace("tcp", "ndnd:6363", false)
	app = engine.NewBasicEngine(tcpFace)
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}
	defer app.Stop()

	// 3. Register handler
	handler := &WorkerHandler{Signer: signer}
	mustAttach(app, fmt.Sprintf("/%s/add", appPrefix), handler.OnAddInterest)

	// 4. (Optional) Load manifest and announce shards
	_, err := manifest.LoadManifest(manifestPath)
	if err != nil {
		log.Printf("Warning: failed to load manifest: %v", err)
		// not fatal if unused
	}

	mustAttach(app, fmt.Sprintf("/%s/worker/%s/ready", appPrefix, workerId), func(args ndn.InterestHandlerArgs) {
		data, _ := app.Spec().MakeData(
			args.Interest.Name(),
			&ndn.DataConfig{
				ContentType: optional.Some(ndn.ContentTypeBlob),
				Freshness:   optional.Some(1 * time.Second),
			},
			encoding.Wire{[]byte("ready")},
			nil, // no signer needed
		)
		args.Reply(data.Wire)
	})

	log.Println("✅ Worker is serving. Awaiting Interests...")

	// 5. Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("👋 Worker shutting down.")
}

func mustAttach(app ndn.Engine, prefixStr string, handler ndn.InterestHandler) {
	name, err := encoding.NameFromStr(prefixStr)
	if err != nil {
		log.Fatalf("Invalid prefix: %s: %v", prefixStr, err)
	}
	if err := app.AttachHandler(name, handler); err != nil {
		log.Fatalf("AttachHandler failed for %s: %v", prefixStr, err)
	}
	if err := app.RegisterRoute(name); err != nil {
		log.Fatalf("RegisterRoute failed for %s: %v", prefixStr, err)
	}
	log.Printf("✓ Attached handler to %s", prefixStr)
}
