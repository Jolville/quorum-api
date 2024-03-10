# quorum-api

## Generating a JWT secret

`openssl rand -hex 32`

# Setting db password

gcloud sql users set-password postgres \
--instance=INSTANCE_NAME \
--password=PASSWORD
