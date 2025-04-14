package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UCLA-IRL/flint-go/pkg/proto"
	"github.com/UCLA-IRL/flint-go/pkg/security"

	"github.com/named-data/ndnd/std/encoding"
	enc "github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/engine"
	"github.com/named-data/ndnd/std/engine/face"
	"github.com/named-data/ndnd/std/ndn"
	"github.com/named-data/ndnd/std/types/optional"
	"github.com/named-data/ndnd/std/utils"
)

type staticComputeServer struct {
	proto.UnimplementedStaticComputeServiceServer
	app       ndn.Engine
	signer    ndn.Signer
	appPrefix string
}

func NewStaticComputeServer(appPrefix string) *staticComputeServer {
	signer := security.LoadOrCreateECDSASigner()
	ndnFace := face.NewStreamFace("udp", "ndnd:6363", false)

	app := engine.NewBasicEngine(ndnFace)
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start NDN engine: %v", err)
	}

	return &staticComputeServer{
		app:       app,
		signer:    signer,
		appPrefix: appPrefix,
	}
}

func (s *staticComputeServer) Add(ctx context.Context, req *proto.AddRequest) (*proto.AddResponse, error) {
	// Construct NDN name: /ndn-compute/add/x/y
	nameStr := fmt.Sprintf("/%s/add/%d/%d", s.appPrefix, req.X, req.Y)
	name, err := enc.NameFromStr(nameStr)
	if err != nil {
		return &proto.AddResponse{Success: false, Result: "invalid name"}, nil
	}

	interest, err := s.app.Spec().MakeInterest(name, &ndn.InterestConfig{
		MustBeFresh: true,
		Lifetime:    optional.Some(4 * time.Second),
		Nonce:       utils.ConvertNonce(s.app.Timer().Nonce()),
	}, encoding.Wire{}, s.signer)
	if err != nil {
		log.Print(err)
		return &proto.AddResponse{Success: false, Result: "failed to make interest"}, nil
	}

	log.Printf("Sending Interest: %s", name.String())

	resultCh := make(chan *proto.AddResponse, 1)
	err = s.app.Express(interest, func(cb ndn.ExpressCallbackArgs) {
		switch cb.Result {
		case ndn.InterestResultData:
			payload := string(cb.Data.Content().Join())
			resultCh <- &proto.AddResponse{Success: true, Result: payload}
		case ndn.InterestResultTimeout:
			resultCh <- &proto.AddResponse{Success: false, Result: "timeout"}
		case ndn.InterestResultNack:
			resultCh <- &proto.AddResponse{Success: false, Result: "nack"}
		default:
			resultCh <- &proto.AddResponse{Success: false, Result: "unknown"}
		}
	})
	if err != nil {
		return &proto.AddResponse{Success: false, Result: "express failed"}, nil
	}

	select {
	case res := <-resultCh:
		return res, nil
	case <-ctx.Done():
		return &proto.AddResponse{Success: false, Result: "client cancelled"}, nil
	}
}
