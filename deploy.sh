#!/bin/bash
set -e

secrets_string=$(cat prod.secrets.env | tr '\n' ',')

git diff --exit-code

current_tag=$(git tag)
echo "Current tag: $current_tag"
echo "New tag?"
read new_tag

docker build -t australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag .

docker push australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag

gcloud run deploy SERVICE --image australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag \
    --update-secrets=JWT_SECRET=JWT_SECRET:1,DATABASE_URL=DATABASE_URL:1

git tag $new_tag

git push origin $new_tag
