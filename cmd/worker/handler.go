package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/ndn"
	"github.com/named-data/ndnd/std/types/optional"
)

type WorkerHandler struct {
	Signer ndn.Signer
}

func (h *WorkerHandler) OnAddInterest(args ndn.InterestHandlerArgs) {
	name := args.Interest.Name()
	log.Printf("Received add interest: %s", name.String())

	strippedName := name.Prefix(-1)
	n := len(strippedName)
	if n < 2 {
		log.Printf("Not enough name components for operands")
		h.sendData(args, name, []byte("error: not enough components"))
		return
	}

	// Parse last two components as operands
	x, err1 := strconv.Atoi(strippedName.At(n - 2).String())
	y, err2 := strconv.Atoi(strippedName.At(n - 1).String())
	if err1 != nil || err2 != nil {
		log.Printf("Failed to parse operands: %v, %v", err1, err2)
		h.sendData(args, name, []byte("error: invalid operands"))
		return
	}

	sum := x + y
	log.Printf("Computed %d + %d = %d", x, y, sum)
	result := fmt.Sprintf("%d", sum)

	h.sendData(args, name, []byte(result))
}

func (h *WorkerHandler) sendData(args ndn.InterestHandlerArgs, name encoding.Name, content []byte) {
	data, err := app.Spec().MakeData(
		name,
		&ndn.DataConfig{
			Freshness:   optional.Some(5 * time.Second),
			ContentType: optional.Some(ndn.ContentTypeBlob),
		},
		encoding.Wire{content},
		h.Signer,
	)
	if err != nil {
		log.Printf("Failed to build Data: %v", err)
		return
	}
	if err := args.Reply(data.Wire); err != nil {
		log.Printf("Failed to reply with Data: %v", err)
		return
	}
	log.Printf("Sent Data: %s (len=%d)", name, len(content))
}
