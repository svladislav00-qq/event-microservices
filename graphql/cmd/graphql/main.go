package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"

	graph "github.com/svladislav00-qq/event-microservices/graphql"
)

type AppConfig struct {
	AccountUrl  string `envconfig:"ACCOUNT_SERVICE_URL"`
	EventUrl    string `envconfig:"EVENT_SERVICE_URL"`
	AttendeeUrl string `envconfig:"ATTENDEE_SERVICE_URL"`
}

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("No .env file found")
	}

	authAddr := os.Getenv("AUTH_GRPC_ADDRESS")
	eventAddr := os.Getenv("EVENT_GRPC_ADDRESS")
	attendeeAddr := os.Getenv("ATTENDEE_GRPC_ADDRESS")

	srv, err := graph.NewServer(authAddr, eventAddr, attendeeAddr)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	resolver := &graph.Resolver{
		Server: srv,
	}

	gqlServer := handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{
			Resolvers: resolver,
		}),
	)

	http.Handle("/query", graph.AuthMiddleware(gqlServer))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/download-table", graph.AuthMiddleware(graph.DownloadAttendeesHandler(srv)))

	port := "8080"
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
