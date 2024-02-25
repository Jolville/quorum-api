//go:generate go run github.com/99designs/gqlgen generate

package graph

import "quorum-api/database"

type Resolver struct{
	JWTSecret string
	DB database.Q
}
