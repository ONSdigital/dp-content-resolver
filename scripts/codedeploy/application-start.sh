#!/bin/bash

AWS_REGION=
CONFIG_BUCKET=
ECR_REPOSITORY_URI=
GIT_COMMIT=

INSTANCE=$(curl -s http://instance-data/latest/meta-data/instance-id)
CONFIG=$(aws --region $AWS_REGION ec2 describe-tags --filters "Name=resource-id,Values=$INSTANCE" "Name=key,Values=Configuration" --output text | awk '{print $5}')

if [[ $DEPLOYMENT_GROUP_NAME =~ [a-z]+-publishing ]]; then
  CONFIG_DIRECTORY=publishing
  DOCKER_NETWORK=publishing
else
  CONFIG_DIRECTORY=web
  DOCKER_NETWORK=website
fi

(aws s3 cp s3://$CONFIG_BUCKET/content-resolver/$CONFIG_DIRECTORY/$CONFIG.asc . && gpg --decrypt $CONFIG.asc > $CONFIG) || exit $?

source $CONFIG && docker run -d  \
  --env=BABBAGE_URL=$BABBAGE_URL \
  --env=BIND_ADDR=$BIND_ADDR     \
  --env=HUMAN_LOG=$HUMAN_LOG     \
  --env=ZEBEDEE_URL=$ZEBEDEE_URL \
  --name=content-resolver        \
  --net=$DOCKER_NETWORK          \
  --restart=always               \
  $ECR_REPOSITORY_URI/content-resolver:$GIT_COMMIT
