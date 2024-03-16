package graph

import (
	"context"
	"net/http"
	srvcustomer "quorum-api/services/customer"
	srvpost "quorum-api/services/post"
	"time"

	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

type loadersCtxKey struct{}

func GetLoaders(ctx context.Context) *Loaders {
	return ctx.Value(loadersCtxKey{}).(*Loaders)
}

type Loaders struct {
	CustomerLoader   *dataloadgen.Loader[uuid.UUID, *srvcustomer.Customer]
	PostLoader       *dataloadgen.Loader[uuid.UUID, *srvpost.Post]
	PostOptionLoader *dataloadgen.Loader[uuid.UUID, *srvpost.Option]
	PostVoteLoader   *dataloadgen.Loader[uuid.UUID, *srvpost.Vote]
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
		PostLoader: dataloadgen.NewLoader(
			getters.getPosts, dataloadgen.WithWait(time.Millisecond),
		),
		PostOptionLoader: dataloadgen.NewLoader(
			getters.getPostOptions, dataloadgen.WithWait(time.Millisecond),
		),
		PostVoteLoader: dataloadgen.NewLoader(
			getters.getPostVotes, dataloadgen.WithWait(time.Millisecond),
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
	ctx context.Context, ids []uuid.UUID,
) ([]*srvcustomer.Customer, []error) {
	customers, err := g.services.Customer.GetCustomersByFilter(
		ctx, srvcustomer.GetCustomersByFilterRequest{
			IDs: ids,
		},
	)
	if err != nil {
		return nil, []error{err}
	}
	customersMap := map[uuid.UUID]*srvcustomer.Customer{}
	for _, c := range customers {
		customersMap[c.ID] = &c
	}
	result := []*srvcustomer.Customer{}
	for _, id := range ids {
		result = append(result, customersMap[id])
	}
	return result, nil
}

func (g *getters) getPostOptions(
	ctx context.Context, ids []uuid.UUID,
) ([]*srvpost.Option, []error) {
	options, err := g.services.Post.GetOptionsByFilter(
		ctx, srvpost.GetOptionsByFilterRequest{
			IDs: ids,
		},
	)
	if err != nil {
		return nil, []error{err}
	}
	oMap := map[uuid.UUID]*srvpost.Option{}
	for _, o := range options {
		oMap[o.ID] = &o
	}
	result := []*srvpost.Option{}
	for _, id := range ids {
		result = append(result, oMap[id])
	}
	return result, nil
}

func (g *getters) getPostVotes(
	ctx context.Context, ids []uuid.UUID,
) ([]*srvpost.Vote, []error) {
	votes, err := g.services.Post.GetVotesByFilter(
		ctx, srvpost.GetVotesByFilterRequest{
			IDs: ids,
		},
	)
	if err != nil {
		return nil, []error{err}
	}
	vMap := map[uuid.UUID]*srvpost.Vote{}
	for _, v := range votes {
		vMap[v.ID] = &v
	}
	result := []*srvpost.Vote{}
	for _, id := range ids {
		result = append(result, vMap[id])
	}
	return result, nil
}

func (g *getters) getPosts(
	ctx context.Context, ids []uuid.UUID,
) ([]*srvpost.Post, []error) {
	posts, err := g.services.Post.GetPostsByFilter(
		ctx, srvpost.GetPostsByFilterRequest{
			IDs: ids,
		},
	)
	if err != nil {
		return nil, []error{err}
	}
	pMap := map[uuid.UUID]*srvpost.Post{}
	for _, p := range posts {
		pMap[p.ID] = &p
	}
	result := []*srvpost.Post{}
	for _, id := range ids {
		result = append(result, pMap[id])
	}
	return result, nil
}
