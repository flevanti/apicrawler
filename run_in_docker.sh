#!/bin/sh
./compile.sh
echo "CREATING DOCKER NETWORK IF IT DOESN'T EXIST 🤔 "
docker network inspect mongonetwork > /dev/null 2>&1 || docker network create mongonetwork > /dev/null
echo RUNNING DOCKER COMMAND 🐳
docker run --network mongonetwork --rm -v "$PWD":/var/task lambci/lambda:go1.x apicrawler "$(< dummyPayload.json)" 
