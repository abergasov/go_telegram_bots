#!/usr/bin/env bash

git fetch --all
git reset --hard origin/master

if make build ; then
  echo "build ok"
else
  echo "build failed"
  exit 1
fi

clean() {
  echo "drop old images"
  docker rmi $(docker images -f dangling=true -q)
}

clean

echo "run prod config"
echo "RUN docker-compose.yml "
docker-compose -f docker-compose.yml pull
docker-compose -f docker-compose.yml up -d --build