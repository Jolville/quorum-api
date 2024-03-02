//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	srvcustomer "quorum-api/services/customer"
)

type Resolver struct {
	JWTSecret string
	Services  Services
}

type Services struct {
	Customer srvcustomer.SRVCustomer
}
