#!/bin/bash -eux

sed dp-content-resolver/appspec.yml dp-content-resolver/scripts/codedeploy/* -i \
  -e s/\${CODEDEPLOY_USER}/$CODEDEPLOY_USER/g                                   \
  -e s/^CONFIG_BUCKET=.*/CONFIG_BUCKET=$CONFIGURATION_BUCKET/                   \
  -e s/^ECR_REPOSITORY_URI=.*/ECR_REPOSITORY_URI=$ECR_REPOSITORY_URI/           \
  -e s/^GIT_COMMIT=.*/GIT_COMMIT=$(cat build/revision)/                         \
  -e s/^AWS_REGION=.*/AWS_REGION=$AWS_REGION/

mkdir -p artifacts/scripts/codedeploy

cp dp-content-resolver/appspec.yml artifacts/
cp dp-content-resolver/scripts/codedeploy/* artifacts/scripts/codedeploy
