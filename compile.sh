#!/bin/sh
env GOOS=linux go build -o apicrawler *.go
zip apicrawler.zip apicrawler dummyPayload.json .ENV 

