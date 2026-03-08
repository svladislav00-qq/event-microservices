package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountUrl  string `envconfig:"ACCOUNT_SERVICE_URL"`
	EventUrl    string `envconfig:"EVENT_SERVICE_URL"`
	AttendeeUrl string `envconfig:"ATTENDEE_SERVICE_URL"`
}

func main() {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	s, err := NewGraphQLServer(cfg.AccountUrl, cfg.EventUrl, cfg.AttendeeUrl)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/graphql", handler.NewDefaultServer(s.ToExecutableSchema()))
	http.Handle("/playground", playground.Handler("vlad", "/graphql"))

	log.Fatal(http.ListenAndServe("8080", nil))
}
