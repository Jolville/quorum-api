//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	srvcommunications "quorum-api/services/communications"
	srvcustomer "quorum-api/services/customer"
	srvpost "quorum-api/services/post"
)

type Resolver struct {
	JWTSecret string
	Services  Services
}

type Services struct {
	Customer       srvcustomer.SRVCustomer
	Post           srvpost.SRVPost
	Communications srvcommunications.SRVCommunications
}
