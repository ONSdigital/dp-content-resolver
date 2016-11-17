#!/bin/bash

AWS_REGION=
CONFIG_BUCKET=
ECR_REPOSITORY_URI=
GIT_COMMIT=

INSTANCE=$(curl -s http://instance-data/latest/meta-data/instance-id)
CONFIG=$(aws --region $AWS_REGION ec2 describe-tags --filters "Name=resource-id,Values=$INSTANCE" "Name=key,Values=Configuration" --output text | awk '{print $5}')

(aws s3 cp s3://$CONFIG_BUCKET/content-resolver/$CONFIG.asc . && gpg --decrypt $CONFIG.asc > $CONFIG) || exit $?

if [[ $DEPLOYMENT_GROUP_NAME =~ [a-z]+-publishing ]]; then
  DOCKER_NETWORK=publishing
else
  DOCKER_NETWORK=website
fi

source $CONFIG && docker run -d \
  --env=BIND_ADDR=$BIND_ADDR    \
  --name=content-resolver       \
  --net=$DOCKER_NETWORK         \
  --restart=always              \
  $ECR_REPOSITORY_URI/content-resolver:$GIT_COMMIT