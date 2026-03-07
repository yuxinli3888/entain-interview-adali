package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"git.neds.sh/matty/entain/api/proto/racing"
	"git.neds.sh/matty/entain/api/proto/sport"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var (
	apiEndpoint        = flag.String("api-endpoint", "localhost:8000", "API endpoint")
	racingGRPCEndpoint = flag.String("racing-grpc-endpoint", "localhost:9000", "Racing gRPC server endpoint")
	sportGRPCEndpoint  = flag.String("sport-grpc-endpoint", "localhost:9001", "Sport gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Printf("failed running api server: %s\n", err)
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	if err := racing.RegisterRacingHandlerFromEndpoint(
		ctx,
		mux,
		*racingGRPCEndpoint,
		[]grpc.DialOption{grpc.WithInsecure()},
	); err != nil {
		return err
	}

	if err := sport.RegisterSportHandlerFromEndpoint(
		ctx,
		mux,
		*sportGRPCEndpoint,
		[]grpc.DialOption{grpc.WithInsecure()},
	); err != nil {
		return err
	}

	log.Printf("API server listening on: %s\n", *apiEndpoint)

	return http.ListenAndServe(*apiEndpoint, mux)
}
