package main

import (
	"log"
	"net/http"
	"os"
	"quorum-api/database"
	"quorum-api/graph"
	srvuser "quorum-api/services/user"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("expected \"JWT_SECRET\" environment variable")
	}

	dbConnString, err := database.GetConnectionStringFromEnv()
	if err != nil {
		log.Fatalf("getting db conn string: %v", err)
	}

	db, err := database.New(dbConnString)
	if err != nil {
		log.Fatalf("connecting to db: %v", err)
	}

	services := graph.Services{
		User: srvuser.New(db),
	}

	var srv http.Handler = handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{Resolvers: &graph.Resolver{
				JWTSecret: jwtSecret,
				Services: services,
			}},
		),
	)

	srv = graph.LoadersMiddleware(services, srv)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
