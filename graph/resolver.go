//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	srvuser "quorum-api/services/user"
)

type Resolver struct{
	JWTSecret string
	Services Services
}

type Services struct {
	User srvuser.SRVUser
}
