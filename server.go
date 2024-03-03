package main

import (
	"log"
	"net/http"
	"os"
	"quorum-api/database"
	"quorum-api/graph"
	srvcustomer "quorum-api/services/customer"

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

	if os.Getenv("GO_ENV") == "local" {
		err := godotenv.Load("local.env")
		if err != nil {
			log.Fatal("Error loading local.env file")
		}
		err = godotenv.Load("local.secrets.env")
		if err != nil {
			log.Fatal("Error loading local.secrets.env file")
		}
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
		Customer: srvcustomer.New(db),
	}

	var srv http.Handler = handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{Resolvers: &graph.Resolver{
				JWTSecret: jwtSecret,
				Services:  services,
			}},
		),
	)

	srv = AddAccessControlHeaders(srv)
	srv = graph.LoadersMiddleware(services, srv)
	srv = graph.AuthMiddleware(jwtSecret)(srv)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func AddAccessControlHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		next.ServeHTTP(w, r)
	})
}
