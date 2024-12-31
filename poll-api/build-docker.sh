#!/bin/bash
go build main.go
docker build --tag poll-container:v1  -f ./Dockerfile .
