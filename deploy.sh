#!/bin/bash
set -e

env_vars_string=$(cat prod.env | tr '\n' ',')

git diff --exit-code

current_tag=$(git tag)
echo "Current tag: $current_tag"
echo "New tag?"
read new_tag

docker build --platform linux/amd64 -t australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag .

docker push australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag

gcloud run deploy three-tier-app-api \
    --image australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag \
    --update-secrets=JWT_SECRET=JWT_SECRET:1 \
    --set-env-vars $env_vars_string \
    --region australia-southeast1

git tag $new_tag

git push origin $new_tag
