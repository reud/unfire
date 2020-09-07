#!/bin/bash

runner() {
  if [ "$1" = "direct" ]; then
    go run main.go
  elif [ "$1" = "docker" ]; then
    docker build -t unfire .
    # read dotenv
    eval "$(cat .env <(echo) <(declare -x))"
    docker run -e PORT=8080 -e TWITTER_CONSUMER_KEY="$TWITTER_CONSUMER_KEY" -e TWITTER_CONSUMER_SECRET="$TWITTER_CONSUMER_SECRET" -p 8080:8080 -t unfire
  else
    echo "usage: ./manager.sh { docker | run }"
  fi
}

deploy() {
  echo "工事中！"
}


allocator() {
  if [ "$1" = "run" ]; then
    runner "$2"
  elif [ "$1" = "deploy" ]; then
    deploy
  else
    echo "usage: ./manager.sh { run | deploy } { command argument }"
  fi
}

allocator "$1" "$2"