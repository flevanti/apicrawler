#!/bin/sh
./compile.sh
docker run --rm -v "$PWD":/var/task lambci/lambda:go1.x apicrawler "$(< dummyPayload.json)" 
