#!/usr/bin/env bash

git fetch --all
git reset --hard origin/master


FILE_HASH=$(git rev-parse HEAD)

if make build hash="$FILE_HASH"; then
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
export GIT_HASH=$FILE_HASH
docker-compose -f docker-compose.yml pull
docker-compose -f docker-compose.yml up -d --build