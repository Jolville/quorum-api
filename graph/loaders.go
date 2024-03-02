package graph

import (
	"context"
	"net/http"
	srvcustomer "quorum-api/services/customer"
	"time"

	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

type loadersCtxKey struct{}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	CustomerLoader *dataloadgen.Loader[uuid.UUID, srvcustomer.Customer]
}

type getters struct {
	services Services
}

func NewLoaders(services Services) *Loaders {
	getters := getters{services: services}
	return &Loaders{
		CustomerLoader: dataloadgen.NewLoader(
			getters.getCustomers, dataloadgen.WithWait(time.Millisecond),
		),
	}
}

func LoadersMiddleware(services Services, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewLoaders(services)
		r = r.WithContext(context.WithValue(r.Context(), loadersCtxKey{}, loader))
		next.ServeHTTP(w, r)
	})
}

func (g *getters) getCustomers(
	ctx context.Context, customerIDs []uuid.UUID,
) ([]srvcustomer.Customer, []error) {
	customers, err := g.services.Customer.GetCustomersByFilter(
		ctx, srvcustomer.GetCustomersByFilterRequest{
			IDs: customerIDs,
		},
	)
	if err != nil {
		return nil, []error{err}
	}
	return customers, nil
}

func GetLoaders(ctx context.Context) *Loaders {
	return ctx.Value(loadersCtxKey{}).(*Loaders)
}
