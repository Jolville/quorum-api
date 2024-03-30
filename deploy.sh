#!/bin/bash
set -e

env_vars_string=$(cat prod.env | tr '\n' ',')

git diff --exit-code

current_tag=$(git describe --tags --abbrev=0)
echo "Current tag: $current_tag"
echo 'New tag (set to "latest" to re-deploy latest image)?'
read new_tag

if [$new_tag != "latest"]; then
    docker build --platform linux/amd64 -t australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag .
    docker push australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag
fi

gcloud run deploy three-tier-app-api \
    --image australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum/quorum-api:$new_tag \
    --update-secrets=JWT_SECRET=JWT_SECRET:latest,MJ_API_KEY=MJ_API_KEY:latest,MJ_SECRET_KEY=MJ_SECRET_KEY:latest \
    --set-env-vars $env_vars_string \
    --region australia-southeast1

git tag $new_tag

git push origin $new_tag
