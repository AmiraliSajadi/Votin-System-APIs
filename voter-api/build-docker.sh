#!/bin/bash
go build main.go
docker build --tag voter-container:v1  -f ./Dockerfile .
