#!/bin/sh
echo REMOVING PREVIOUS COMPILED FILE
rm apicrawler
echo COMPILING FOR LINUX 🐧
env GOOS=linux go build -o apicrawler *.go
echo ZIPPING FOR AWS LAMBDA 👷️
zip apicrawler.zip apicrawler dummyPayload.json .env 
echo DONE 🍾
