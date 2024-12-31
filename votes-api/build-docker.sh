#!/bin/bash
go build main.go
docker build --tag votes-container:v1  -f ./Dockerfile .
