package main

import (
	"database/sql"
	"flag"
	"log"
	"net"

	"git.neds.sh/matty/entain/sport/db"
	"git.neds.sh/matty/entain/sport/proto/sport"
	"git.neds.sh/matty/entain/sport/service"
	"google.golang.org/grpc"
)

var (
	grpcEndpoint = flag.String("grpc-endpoint", "localhost:9001", "gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running grpc server: %s\n", err)
	}
}

func run() error {
	conn, err := net.Listen("tcp", ":9001")
	if err != nil {
		return err
	}

	sportDB, err := sql.Open("sqlite3", "./db/sport.db")
	if err != nil {
		return err
	}

	eventsRepo := db.NewEventsRepo(sportDB)
	if err := eventsRepo.Init(); err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	sport.RegisterSportServer(
		grpcServer,
		service.NewSportService(eventsRepo),
	)

	log.Printf("sport gRPC server listening on: %s\n", *grpcEndpoint)

	if err := grpcServer.Serve(conn); err != nil {
		return err
	}

	return nil
}
