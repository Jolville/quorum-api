package graph

import (
	"context"
	"net/http"
	srvuser "quorum-api/services/user"
	"time"

	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

type loadersCtxKey struct{}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	UserLoader *dataloadgen.Loader[uuid.UUID, srvuser.User]
}

type getters struct {
	services Services
}

func NewLoaders(services Services) *Loaders {
	getters := getters{services: services}
	return &Loaders{
		UserLoader: dataloadgen.NewLoader(
			getters.getUsers, dataloadgen.WithWait(time.Millisecond),
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

func (g *getters) getUsers(
	ctx context.Context, userIDs []uuid.UUID,
) ([]srvuser.User, []error) {
	users, err := g.services.User.GetUsersByFilter(
		ctx, srvuser.GetUsersByFilterRequest{
			IDs: userIDs,
		},
	)
	if err != nil {
		return nil, []error{err}
	}
	return users, nil
}

func GetLoaders(ctx context.Context) *Loaders {
	return ctx.Value(loadersCtxKey{}).(*Loaders)
}
