# quorum-api

## Workflow for adding fields to graph

1. Modify `./graph/schema.graphql`
2. Implement functions in `graph/schema.resolvers.go`

## Generating a JWT secret

`openssl rand -hex 32`

# Setting db password

gcloud sql users set-password postgres \
--instance=INSTANCE_NAME \
--password=PASSWORD
