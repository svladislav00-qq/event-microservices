package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	graph "github.com/svladislav00-qq/event-microservices/graphql"
)

type AppConfig struct {
	AccountUrl  string `envconfig:"ACCOUNT_SERVICE_URL"`
	EventUrl    string `envconfig:"EVENT_SERVICE_URL"`
	AttendeeUrl string `envconfig:"ATTENDEE_SERVICE_URL"`
}

func main() {
	srv, err := graph.NewServer(
		"localhost:44044",
		"localhost:44045",
		"localhost:44046",
	)
	if err != nil {
		log.Fatal(err)
	}

	resolver := &graph.Resolver{
		Server: srv,
	}

	http.Handle("/query", graph.AuthMiddleware(handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{Resolvers: resolver}),
	)))

	http.Handle("/", playground.Handler("GraphQL", "/query"))

	log.Println("server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
