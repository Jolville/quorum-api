# quorum-api

Monorepo containing a Go GraphQL API, services, and postgres database. Deployed to GCP cloud run as a docker container.

Frontend code found at https://github.com/Jolville/quorum-react

## Quickstart

Prerequisite: will need access to quorum GCP API project in GCP to decrypt secrets:

1. `make decrypt env=local`
2. `make dbup`
3. `make migrate`
4. `make watch`

## Structure

```
.
├── Dockerfile
├── Makefile # Commands (mostly) for local dev
├── database # Interface for common database calls and constructor to create connection
├── deploy.sh # Deploys to production
├── gqlgen.yml # GraphQL config
├── graph
│   ├── auth.go # Auth middleware
│   ├── loaders.go # Dataloaders
│   ├── schema.graphql # GraphQL SDL file
│   └── schema.resolvers.go # Resolver implementation
├── local.env
├── local.secrets.enc.env # Secrets encrypted with sops https://github.com/getsops/sops
├── local.secrets.env
├── migrate-prod.sh # Runs migration against prod
├── migrations # Database migration files
│   ├── *.up.sql
├── prod.env
├── prod.secrets.enc.env https://github.com/getsops/sops
├── prod.secrets.env
├── server.go # main function that starts server
├── services # Each service has a service definition/implementation and optional Dao layer
│   ├── communications
│   │   └── communications.go
│   ├── customer
│   │   ├── dao.go
│   │   └── user.go
│   └── post
│       ├── dao.go
│       └── post.go
└── tools.go
```

Callouts:

- All read functions in svc's are listy so dataloading is supported by default.
- Could be decomposed easily into microservices later by keeping services as sepearate packages.

## Workflow for adding fields to graph

1. Modify `./graph/schema.graphql`
2. Run `go generate ./...`
3. Implement functions in `graph/schema.resolvers.go`

## Auth

- User is sent a confirm login URL with temporary token encoded in query string.
  - Token has short exp time and doesn't have verified claim encoded.
- Login page submits this token, then backend validates it and returns verified token to client with longer exp time.
