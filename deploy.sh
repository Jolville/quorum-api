#!/bin/bash
set -e

git diff --exit-code

current_tag=$(git tag)
echo "Current tag: $current_tag"
echo "New tag?"
read new_tag

docker build -t australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum:$new_tag .

docker push australia-southeast1-docker.pkg.dev/trusty-charmer-415303/quorum:$new_tag

git tag $new_tag

git push origin $new_tag
