#!/bin/sh
./compile.sh
echo RUNNING DOCKER COMMAND 🐳
docker run --network mongonetwork --rm -v "$PWD":/var/task lambci/lambda:go1.x apicrawler "$(< dummyPayload.json)" 
