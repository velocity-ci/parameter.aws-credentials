#!/bin/sh -e

scripts/install-deps.sh

export CGO_ENABLED=0 
export GOOS=linux 

go build -a -installsuffix cgo -o dist/aws-credentials
