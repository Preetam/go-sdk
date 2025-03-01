#!/usr/bin/env bash

echo 'Building main package..'
go mod tidy && go mod vendor && go build -v ./...
echo 'Building gin package..'
(cd descope/gin && go mod tidy && go mod vendor && go build)
echo 'Building mux web app example..'
(cd examples/webapp && go mod tidy && go mod vendor && go build)
echo 'Building gin web app example..'
(cd examples/ginwebapp && go mod tidy && go mod vendor && go build)