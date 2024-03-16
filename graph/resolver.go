//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	srvcustomer "quorum-api/services/customer"
	srvpost "quorum-api/services/post"
)

type Resolver struct {
	JWTSecret string
	Services  Services
}

type Services struct {
	Customer srvcustomer.SRVCustomer
	Post     srvpost.SRVPost
}
